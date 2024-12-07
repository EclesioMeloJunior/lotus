package llvm_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/EclesioMeloJunior/lotus/ir/llvm"
	"github.com/EclesioMeloJunior/lotus/lexer"
	"github.com/EclesioMeloJunior/lotus/parser"
	"github.com/stretchr/testify/require"

	gollvm "tinygo.org/x/go-llvm"
)

func TestIRGenerator_GenerateIR(t *testing.T) {
	input := `fn main() {
	var x = 1;
	var y = 3;
	var z = x + y;
	return z;
}`

	l := lexer.NewLexer(strings.NewReader(input))
	tokens := l.NextToken()
	p := parser.NewParser(tokens)

	program, err := p.ParseProgram()
	require.NoError(t, err)

	irGen := llvm.NewIRGenerator()
	irGen.GenerateIR(program)

	fmt.Println(irGen.Module.String())

	err = gollvm.VerifyModule(irGen.Module, gollvm.PrintMessageAction)
	require.NoError(t, err)

	gollvm.LinkInMCJIT()
	gollvm.InitializeNativeTarget()
	gollvm.InitializeNativeAsmPrinter()

	options := gollvm.NewMCJITCompilerOptions()
	options.SetMCJITOptimizationLevel(2)
	options.SetMCJITEnableFastISel(true)
	options.SetMCJITNoFramePointerElim(true)
	options.SetMCJITCodeModel(gollvm.CodeModelJITDefault)

	engine, err := gollvm.NewMCJITCompiler(irGen.Module, options)
	require.NoError(t, err)
	defer engine.Dispose()

	main := irGen.Module.NamedFunction("main")
	require.NotNil(t, main)

	result := engine.RunFunction(main, nil)
	fmt.Printf("Result: %v\n", result.Int(false))

	//require.Equal(t, expectedIR, irGen.Module.String())
}
