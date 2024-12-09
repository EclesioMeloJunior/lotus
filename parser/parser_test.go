package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/EclesioMeloJunior/lotus/lexer"
	"github.com/EclesioMeloJunior/lotus/parser"
	"github.com/stretchr/testify/require"
)

func TestParser_ParseProgram(t *testing.T) {
	input := "var x = 42; var y = 3.14; var z = x + y;"
	l := lexer.NewLexer(strings.NewReader(input))
	tokens := l.NextToken()
	p := parser.NewParser(tokens)

	program, err := p.ParseProgram()
	require.NoError(t, err)

	require.NotNil(t, program)
	require.Len(t, program.Statements, 3)

	expected := &parser.Program{
		Statements: []parser.Node{
			&parser.VarStatement{
				Name: "x",
				Value: &parser.IntegerLiteral{
					Value: 42,
				},
			},
			&parser.VarStatement{
				Name: "y",
				Value: &parser.FloatLiteral{
					Value: 3.14,
				},
			},

			&parser.VarStatement{
				Name: "z",
				Value: &parser.InfixExpression{
					Left: &parser.Identifier{
						Value: "x",
					},
					Operator: "+",
					Right: &parser.Identifier{
						Value: "y",
					},
				},
			},
		},
	}

	require.Equal(t, expected, program)
}

func TestParser_ParsePrecendence(t *testing.T) {
	input := "var x = 1 + 2 * 3;"
	l := lexer.NewLexer(strings.NewReader(input))
	tokens := l.NextToken()
	p := parser.NewParser(tokens)

	program, err := p.ParseProgram()
	require.NoError(t, err)

	expected := &parser.Program{
		Statements: []parser.Node{
			&parser.VarStatement{
				Name: "x",
				Value: &parser.InfixExpression{
					Left: &parser.IntegerLiteral{
						Value: 1,
					},
					Operator: "+",
					Right: &parser.InfixExpression{
						Left: &parser.IntegerLiteral{
							Value: 2,
						},
						Operator: "*",
						Right: &parser.IntegerLiteral{
							Value: 3,
						},
					},
				},
			},
		},
	}

	require.Equal(t, expected, program)
}

func TestParser_ParseGroupedExpression(t *testing.T) {
	input := "var x = 2 * (42 + 3.14);"
	l := lexer.NewLexer(strings.NewReader(input))
	tokens := l.NextToken()
	p := parser.NewParser(tokens)

	program, err := p.ParseProgram()
	require.NoError(t, err)

	expected := &parser.Program{
		Statements: []parser.Node{
			&parser.VarStatement{
				Name: "x",
				Value: &parser.InfixExpression{
					Left: &parser.IntegerLiteral{
						Value: 2,
					},
					Operator: "*",
					Right: &parser.InfixExpression{
						Left: &parser.IntegerLiteral{
							Value: 42,
						},
						Operator: "+",
						Right: &parser.FloatLiteral{
							Value: 3.14,
						},
					},
				},
			},
		},
	}

	require.Equal(t, expected, program)
}

func TestParser_ParseFunction(t *testing.T) {
	input := `fn add(a: int32, b: int32): int32 {
	return a + b;
}`

	l := lexer.NewLexer(strings.NewReader(input))
	tokens := l.NextToken()
	p := parser.NewParser(tokens)

	program, err := p.ParseProgram()
	require.NoError(t, err)

	require.NotNil(t, program)
	require.Len(t, program.Statements, 1)

	expected := &parser.Program{
		Statements: []parser.Node{
			&parser.FnStatement{
				ReturnType: parser.Int32,
				Name:       "add",
				Args: []*parser.Argument{
					{Name: "a", Type: parser.Int32},
					{Name: "b", Type: parser.Int32},
				},
				Body: []parser.Node{
					&parser.ReturnStatement{
						Value: &parser.InfixExpression{
							Left: &parser.Identifier{
								Value: "a",
							},
							Operator: "+",
							Right: &parser.Identifier{
								Value: "b",
							},
						},
					},
				},
			},
		},
	}

	require.Equal(t, expected, program)
}

func TestParser_ShouldErrorIfReturnIsNotPresent(t *testing.T) {
	input := `fn add(a, b): int32 {
	var a: int32 = 10;
}`

	l := lexer.NewLexer(strings.NewReader(input))
	tokens := l.NextToken()
	p := parser.NewParser(tokens)

	_, err := p.ParseProgram()
	require.Error(t, err, &parser.ErrParser{
		Line:   3,
		Column: 0,
		Err:    errors.New("function must have a return"),
	})
}
