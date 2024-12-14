package tests

import (
	"testing"

	"github.com/EclesioMeloJunior/lotus/ir/llvm"
	"github.com/EclesioMeloJunior/lotus/lexer"
	"github.com/EclesioMeloJunior/lotus/parser"
	"github.com/stretchr/testify/require"
	gollvm "tinygo.org/x/go-llvm"
)

func TestSimpleMain(t *testing.T) {
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

func TestSimpleMainWithTypes(t *testing.T) {
	src := readInput(t, "./simple_main_with_types.lt")
	l := lexer.NewLexer(src)

	p := parser.NewParser(l.NextToken())

	program, err := p.ParseProgram()
	require.NoError(t, err)

	require.Len(t, program.Statements, 2)
	require.Equal(t, &parser.FnStatement{
		Name: "main",
		Args: []*parser.Argument{},
		Body: []parser.Node{
			&parser.VarStatement{
				Name:  "a",
				Value: &parser.IntegerLiteral{Value: 3},
				Type:  parser.Int32,
			},
			&parser.VarStatement{
				Name:  "b",
				Value: &parser.IntegerLiteral{Value: 0},
				Type:  parser.Int32,
			},
			&parser.VarStatement{
				Name:  "c",
				Value: &parser.Identifier{Value: "b", Type: parser.Int32},
				Type:  parser.Int32,
			},

			&parser.VarStatement{
				Name: "name",
				Value: &parser.InfixExpression{
					Left: &parser.InfixExpression{
						Left:     &parser.StringLiteral{Value: "Eclesio"},
						Operator: "+",
						Right:    &parser.StringLiteral{Value: " "},
					},
					Operator: "+",
					Right:    &parser.StringLiteral{Value: "sio"},
				},
				Type: parser.String,
			},

			&parser.FnCall{
				FnName: "print",
				Params: []parser.Expression{
					&parser.Identifier{
						Value: "name",
						Type:  parser.String,
					},
				},
			},

			&parser.ReassignVarStatement{
				VarName: "a",
				Value:   &parser.IntegerLiteral{Value: 1},
				Type:    parser.Int32,
			},

			&parser.ReturnStatement{
				Value: &parser.InfixExpression{
					Operator: "+",
					Left:     &parser.Identifier{Value: "a", Type: parser.Int32},
					Right:    &parser.Identifier{Value: "b", Type: parser.Int32},
				},
				Type: 1,
			},
		},
		ReturnType: parser.Int32,
	}, program.Statements[0])

	require.Equal(t, &parser.FnStatement{
		Name: "print",
		Args: []*parser.Argument{
			&parser.Argument{
				Name: "name",
				Type: parser.String,
			},
		},
		Body:       []parser.Node{},
		ReturnType: parser.Void,
	}, program.Statements[1])
}
