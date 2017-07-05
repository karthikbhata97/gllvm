package shared

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type bitcodeToObjectLink struct {
	bcPath  string
	objPath string
}

func Compile(args []string, compilerName string) (exitCode int) {
	exitCode = 0
	//in the configureOnly case we have to know the exit code of the compile
	//because that is how configure figures out what it can and cannot do.

	var ok bool = true

	var compilerExecName = getCompilerExecName(compilerName)
	var configureOnly bool
	if ConfigureOnly != "" {
		configureOnly = true
	}
	var pr = parse(args)

	var wg sync.WaitGroup
	// If configure only is set, just execute the compiler

	if configureOnly {
		wg.Add(1)
		go execCompile(compilerExecName, pr, &wg, &ok)
		wg.Wait()
		// Else try to build bitcode as well

		if !ok {
			exitCode = 1
		}

	} else {
		var bcObjLinks []bitcodeToObjectLink
		var newObjectFiles []string

		wg.Add(2)
		go execCompile(compilerExecName, pr, &wg, &ok)
		go buildAndAttachBitcode(compilerExecName, pr, &bcObjLinks, &newObjectFiles, &wg)
		wg.Wait()

		//grok the exit code
		if !ok {
			exitCode = 1
		} else {

			// When objects and bitcode are built we can attach bitcode paths
			// to object files and link
			for _, link := range bcObjLinks {
				attachBitcodePathToObject(link.bcPath, link.objPath)
			}
			if !pr.IsCompileOnly {
				compileTimeLinkFiles(compilerExecName, pr, newObjectFiles)
			}

		}
	}
	return
}

// Compiles bitcode files and mutates the list of bc->obj links to perform + the list of
// new object files to link
func buildAndAttachBitcode(compilerExecName string, pr parserResult, bcObjLinks *[]bitcodeToObjectLink, newObjectFiles *[]string, wg *sync.WaitGroup) {
	defer (*wg).Done()
	// If nothing to do, exit silently
	if !pr.IsEmitLLVM && !pr.IsAssembly && !pr.IsAssembleOnly &&
		!(pr.IsDependencyOnly && !pr.IsCompileOnly) && !pr.IsPreprocessOnly {
		var hidden = !pr.IsCompileOnly

		if len(pr.InputFiles) == 1 && pr.IsCompileOnly {
			var srcFile = pr.InputFiles[0]
			objFile, bcFile := getArtifactNames(pr, 0, hidden)
			buildBitcodeFile(compilerExecName, pr, srcFile, bcFile)
			*bcObjLinks = append(*bcObjLinks, bitcodeToObjectLink{bcPath: bcFile, objPath: objFile})
		} else {
			for i, srcFile := range pr.InputFiles {
				objFile, bcFile := getArtifactNames(pr, i, hidden)
				if hidden {
					buildObjectFile(compilerExecName, pr, srcFile, objFile)
					*newObjectFiles = append(*newObjectFiles, objFile)
				}
				if strings.HasSuffix(srcFile, ".bc") {
					*bcObjLinks = append(*bcObjLinks, bitcodeToObjectLink{bcPath: srcFile, objPath: objFile})
				} else {
					buildBitcodeFile(compilerExecName, pr, srcFile, bcFile)
					*bcObjLinks = append(*bcObjLinks, bitcodeToObjectLink{bcPath: bcFile, objPath: objFile})
				}
			}
		}
	}
	return
}

func attachBitcodePathToObject(bcFile, objFile string) {
	// We can only attach a bitcode path to certain file types
	switch filepath.Ext(objFile) {
	case
		".o",
		".lo",
		".os",
		".So",
		".po":
		// Store bitcode path to temp file
		var absBcPath, _ = filepath.Abs(bcFile)
		tmpContent := []byte(absBcPath + "\n")
		tmpFile, err := ioutil.TempFile("", "gllvm")
		if err != nil {
			LogFatal("attachBitcodePathToObject: %v\n", err)
		}
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.Write(tmpContent); err != nil {
			LogFatal("attachBitcodePathToObject: %v\n", err)
		}
		if err := tmpFile.Close(); err != nil {
			LogFatal("attachBitcodePathToObject: %v\n", err)
		}

		// Let's write the bitcode section
		var attachCmd string
		var attachCmdArgs []string
		if runtime.GOOS == "darwin" {
			attachCmd = "ld"
			attachCmdArgs = []string{"-r", "-keep_private_externs", objFile, "-sectcreate", DarwinSegmentName, DarwinSectionName, tmpFile.Name(), "-o", objFile}
		} else {
			attachCmd = "objcopy"
			attachCmdArgs = []string{"--add-section", ELFSectionName + "=" + tmpFile.Name(), objFile}
		}

		// Run the attach command and ignore errors
		execCmd(attachCmd, attachCmdArgs, "")

		// Copy bitcode file to store, if necessary
		if bcStorePath := os.Getenv(BitcodeStorePath); bcStorePath != "" {
			destFilePath := path.Join(bcStorePath, getHashedPath(absBcPath))
			in, _ := os.Open(absBcPath)
			defer in.Close()
			out, _ := os.Create(destFilePath)
			defer out.Close()
			io.Copy(out, in)
			out.Sync()
		}
	}
}

func compileTimeLinkFiles(compilerExecName string, pr parserResult, objFiles []string) {
	var outputFile = pr.OutputFilename
	if outputFile == "" {
		outputFile = "a.out"
	}
	args := pr.LinkArgs
	args = append(args, objFiles...)
	args = append(args, "-o", outputFile)
	success, err := execCmd(compilerExecName, args, "")
	if !success {
		LogFatal("Failed to link: %v.", err)
	}
}

// Tries to build the specified source file to object
func buildObjectFile(compilerExecName string, pr parserResult, srcFile string, objFile string) {
	args := pr.CompileArgs[:]
	args = append(args, srcFile, "-c", "-o", objFile)
	success, err := execCmd(compilerExecName, args, "")
	if !success {
		LogFatal("Failed to build object file for %s because: %v\n", srcFile, err)
	}
}

// Tries to build the specified source file to bitcode
func buildBitcodeFile(compilerExecName string, pr parserResult, srcFile string, bcFile string) {
	args := pr.CompileArgs[:]
	args = append(args, "-emit-llvm", "-c", srcFile, "-o", bcFile)
	success, err := execCmd(compilerExecName, args, "")
	if !success {
		LogFatal("Failed to build bitcode file for %s because: %v\n", srcFile, err)
	}
}

// Tries to build object file
func execCompile(compilerExecName string, pr parserResult, wg *sync.WaitGroup, ok *bool) {
	defer (*wg).Done()
	success, err := execCmd(compilerExecName, pr.InputList, "")
	if !success {
		LogError("Failed to compile using given arguments: %v\n", err)
		*ok = false
	}
}

func getCompilerExecName(compilerName string) string {
	switch compilerName {
	case "clang":
		if LLVMCCName != "" {
			return filepath.Join(LLVMToolChainBinDir, LLVMCCName)
		}
		return filepath.Join(LLVMToolChainBinDir, compilerName)
	case "clang++":
		if LLVMCXXName != "" {
			return filepath.Join(LLVMToolChainBinDir, LLVMCXXName)
		}
		return filepath.Join(LLVMToolChainBinDir, compilerName)
	default:
		LogFatal("The compiler %s is not supported by this tool.", compilerName)
		return ""
	}
}