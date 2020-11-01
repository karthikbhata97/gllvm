package test

import (
	"fmt"
	"github.com/SRI-CSL/gllvm/shared"
	"runtime"
	"testing"
)

func Test_basic_functionality(t *testing.T) {
	args := []string{"../data/helloworld.c", "-o", "../data/hello"}
	exitCode := shared.Compile(args, "clang")
	if exitCode != 0 {
		t.Errorf("Compile of %v returned %v\n", args, exitCode)
	} else {
		fmt.Println("Compiled OK")
	}
	args = []string{"get-bc", "-v", "../data/hello"}
	exitCode = shared.Extract(args)
	if exitCode != 0 {
		t.Errorf("Extraction of %v returned %v\n", args, exitCode)
	} else {
		fmt.Println("Extraction OK")
	}
}

func Test_more_functionality(t *testing.T) {
	objectFile := "../data/bhello.o"
	args := []string{"../data/helloworld.c", "-c", "-o", objectFile}
	exitCode := shared.Compile(args, "clang")
	if exitCode != 0 {
		t.Errorf("Compile of %v returned %v\n", args, exitCode)
	} else {
		fmt.Println("Compiled OK")
	}
	ok, err := shared.IsObjectFileForOS(objectFile, runtime.GOOS)
	if !ok {
		t.Errorf("isObjectFileForOS(%v, %v) = %v (err = %v)\n", objectFile, runtime.GOOS, ok, err)
	} else {
		fmt.Printf("isObjectFileForOS(%v, %v) = %v\n", objectFile, runtime.GOOS, ok)
	}
	args = []string{objectFile, "-o", "../data/bhello"}
	exitCode = shared.Compile(args, "clang")
	if exitCode != 0 {
		t.Errorf("Compile of %v returned %v\n", args, exitCode)
	} else {
		fmt.Println("Compiled OK")
	}
	args = []string{"get-bc", "-v", "../data/bhello"}
	exitCode = shared.Extract(args)
	if exitCode != 0 {
		t.Errorf("Extraction of %v returned %v\n", args, exitCode)
	} else {
		fmt.Println("Extraction OK")
	}
}

func Test_obscure_functionality(t *testing.T) {
	sourceFile := "../data/helloworld.c"
	objectFile := "../data/bhello.notanextensionthatwerecognize"
	exeFile := "../data/bhello"
	opSys := runtime.GOOS
	args := []string{sourceFile, "-c", "-o", objectFile}
	exitCode := shared.Compile(args, "clang")
	if exitCode != 0 {
		t.Errorf("Compile of %v returned %v\n", args, exitCode)
	} else {
		fmt.Println("Compiled OK")
	}
	ok, err := shared.IsObjectFileForOS(sourceFile, opSys)
	if ok {
		t.Errorf("isObjectFileForOS(%v, %v) = %v\n", sourceFile, opSys, ok)
	} else {
		fmt.Printf("isObjectFileForOS(%v, %v) = %v (err = %v)\n", sourceFile, opSys, ok, err)
	}
	ok, err = shared.IsObjectFileForOS(objectFile, opSys)
	if !ok {
		t.Errorf("isObjectFileForOS(%v, %v) = %v (err = %v)\n", objectFile, opSys, ok, err)
	} else {
		fmt.Printf("isObjectFileForOS(%v, %v) = %v\n", objectFile, opSys, ok)
	}
	args = []string{objectFile, "-o", exeFile}
	exitCode = shared.Compile(args, "clang")
	if exitCode != 0 {
		t.Errorf("Compile of %v returned %v\n", args, exitCode)
	} else {
		fmt.Println("Compiled OK")
	}
	ok, err = shared.IsObjectFileForOS(exeFile, opSys)
	if ok {
		t.Errorf("isObjectFileForOS(%v, %v) = %v\n", exeFile, opSys, ok)
	} else {
		fmt.Printf("isObjectFileForOS(%v, %v) = %v (err = %v)\n", exeFile, opSys, ok, err)
	}
	args = []string{"get-bc", "-v", exeFile}
	exitCode = shared.Extract(args)
	if exitCode != 0 {
		t.Errorf("Extraction of %v returned %v\n", args, exitCode)
	} else {
		fmt.Println("Extraction OK")
	}
}

func Test_file_type(t *testing.T) {
	fictionalFile := "HopefullyThereIsNotAFileCalledThisNearBy.txt"
	dataDir := "../data"
	sourceFile := "../data/helloworld.c"
	objectFile := "../data/bhello.notanextensionthatwerecognize"
	exeFile := "../data/bhello"
	fmt.Printf("GetBinaryType(%v) = %v\n", fictionalFile, shared.GetBinaryType(fictionalFile))
	fmt.Printf("GetBinaryType(%v) = %v\n", dataDir, shared.GetBinaryType(dataDir))
	fmt.Printf("GetBinaryType(%v) = %v\n", sourceFile, shared.GetBinaryType(sourceFile))
	fmt.Printf("GetBinaryType(%v) = %v\n", objectFile, shared.GetBinaryType(objectFile))
	fmt.Printf("GetBinaryType(%v) = %v\n", exeFile, shared.GetBinaryType(exeFile))

	plain := shared.IsPlainFile(fictionalFile)
	if plain {
		t.Errorf("shared.IsPlainFile(%v) returned %v\n", fictionalFile, plain)
	} else {
		fmt.Printf("shared.IsPlainFile(%v) returned %v\n", fictionalFile, plain)
	}
	plain = shared.IsPlainFile(dataDir)
	if plain {
		t.Errorf("shared.IsPlainFile(%v) returned %v\n", dataDir, plain)
	} else {
		fmt.Printf("shared.IsPlainFile(%v) returned %v\n", dataDir, plain)
	}
	plain = shared.IsPlainFile(sourceFile)
	if !plain {
		t.Errorf("shared.IsPlainFile(%v) returned %v\n", sourceFile, plain)
	} else {
		fmt.Printf("shared.IsPlainFile(%v) returned %v\n", sourceFile, plain)
	}
	plain = shared.IsPlainFile(objectFile)
	if !plain {
		t.Errorf("shared.IsPlainFile(%v) returned %v\n", objectFile, plain)
	} else {
		fmt.Printf("shared.IsPlainFile(%v) returned %v\n", objectFile, plain)
	}
	plain = shared.IsPlainFile(exeFile)
	if !plain {
		t.Errorf("shared.IsPlainFile(%v) returned %v\n", exeFile, plain)
	} else {
		fmt.Printf("shared.IsPlainFile(%v) returned %v\n", exeFile, plain)
	}

}
