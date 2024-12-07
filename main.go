package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/EclesioMeloJunior/lotus/ir/llvm"
	"github.com/EclesioMeloJunior/lotus/lexer"
	"github.com/EclesioMeloJunior/lotus/parser"

	gollvm "tinygo.org/x/go-llvm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: main <source file>")
		return
	}

	sourceFile := os.Args[1]
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Printf("Error reading source file: %v\n", err)
		return
	}

	l := lexer.NewLexer(strings.NewReader(string(source)))
	tokens := l.NextToken()
	p := parser.NewParser(tokens)

	program, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("Error parsing program: %v\n", err)
		return
	}

	irGen := llvm.NewIRGenerator()
	irGen.GenerateIR(program)

	err = gollvm.VerifyModule(irGen.Module, gollvm.PrintMessageAction)
	if err != nil {
		fmt.Printf("Error verifying module: %v\n", err)
		return
	}

	irFile, err := os.Create("output.ll")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer irFile.Close()
	irGen.Module.Dump()
	_, err = irFile.Write([]byte(irGen.Module.String()))
	if err != nil {
		fmt.Printf("Error writing to output file: %v\n", err)
		return
	}

	// gollvm.LinkInMCJIT()
	// gollvm.InitializeNativeTarget()
	// gollvm.InitializeNativeAsmPrinter()

	// options := gollvm.NewMCJITCompilerOptions()
	// options.SetMCJITOptimizationLevel(2)
	// options.SetMCJITEnableFastISel(true)
	// options.SetMCJITNoFramePointerElim(true)
	// options.SetMCJITCodeModel(gollvm.CodeModelJITDefault)

	// engine, err := gollvm.NewMCJITCompiler(irGen.Module, options)
	// if err != nil {
	// 	fmt.Printf("Error creating MCJIT compiler: %v\n", err)
	// 	return
	// }

	// defer engine.Dispose()
	// engine.RunFunction(irGen.Module.NamedFunction("main"), nil)
}
