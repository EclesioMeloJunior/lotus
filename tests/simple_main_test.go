package tests

import (
	"testing"

	"github.com/EclesioMeloJunior/lotus/ir/llvm"
	"github.com/EclesioMeloJunior/lotus/lexer"
	"github.com/EclesioMeloJunior/lotus/parser"
	"github.com/stretchr/testify/require"
	gollvm "tinygo.org/x/go-llvm"
)

func TestSimpleInput(t *testing.T) {
	src := readInput(t, "./simple_main.lt")

	l := lexer.NewLexer(src)
	p := parser.NewParser(l.NextToken())

	program, err := p.ParseProgram()
	require.NoError(t, err)

	irGen := llvm.NewIRGenerator()
	irGen.GenerateIR(program)

	runMainFn(t, irGen, func(ee gollvm.ExecutionEngine, gv gollvm.GenericValue) {
		require.Equal(t, uint64(3), gv.Int(false))
	})
}
