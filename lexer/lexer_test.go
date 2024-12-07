package lexer_test

import (
	"strings"
	"testing"

	"github.com/EclesioMeloJunior/lotus/lexer"
	"github.com/stretchr/testify/require"
)

func TestLexerNextToken(t *testing.T) {
	input := "var x = 42 + 5 main fn continue if break { } ( ) 3.14 ;"
	l := lexer.NewLexer(strings.NewReader(input))

	exepectedTokens := []lexer.Token{
		{Type: lexer.VAR, Literal: "var", Line: 1, Column: 0},
		{Type: lexer.IDENT, Literal: "x", Line: 1, Column: 4},
		{Type: lexer.ASSIGN, Literal: "=", Line: 1, Column: 6},
		{Type: lexer.INT, Literal: "42", Line: 1, Column: 8},
		{Type: lexer.PLUS, Literal: "+", Line: 1, Column: 11},
		{Type: lexer.INT, Literal: "5", Line: 1, Column: 13},
		{Type: lexer.MAIN, Literal: "main", Line: 1, Column: 15},
		{Type: lexer.FN, Literal: "fn", Line: 1, Column: 20},
		{Type: lexer.CONTINUE, Literal: "continue", Line: 1, Column: 23},
		{Type: lexer.IF, Literal: "if", Line: 1, Column: 32},
		{Type: lexer.BREAK, Literal: "break", Line: 1, Column: 35},
		{Type: lexer.LBRACE, Literal: "{", Line: 1, Column: 41},
		{Type: lexer.RBRACE, Literal: "}", Line: 1, Column: 43},
		{Type: lexer.LPAREN, Literal: "(", Line: 1, Column: 45},
		{Type: lexer.RPAREN, Literal: ")", Line: 1, Column: 47},
		{Type: lexer.FLOAT, Literal: "3.14", Line: 1, Column: 49},
		{Type: lexer.SEMICOLON, Literal: ";", Line: 1, Column: 54},
		{Type: lexer.EOF, Literal: "", Line: 0, Column: 0},
	}

	var tokens []lexer.Token
	for tok := range l.NextToken() {
		tokens = append(tokens, tok)
	}

	require.Equal(t, exepectedTokens, tokens)
}
