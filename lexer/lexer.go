package lexer

import (
	"bufio"
	"bytes"
	"io"
	"iter"
	"strings"
)

// TokenType represents the type of token.
type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	VAR
	IDENT
	INT
	ASSIGN
	PLUS
	MINUS
	STAR
	SLASH
	MAIN
	FN
	CONTINUE
	IF
	BREAK
	LBRACE
	RBRACE
	LPAREN
	RPAREN
	FLOAT
	SEMICOLON
	NEXTLINE
	RETURN
	COMMA
)

var Keywords = map[string]TokenType{
	"var":      VAR,
	"fn":       FN,
	"continue": CONTINUE,
	"if":       IF,
	"break":    BREAK,
	"return":   RETURN,
}

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t *Token) String() string {
	return t.Literal
}

// Lexer represents a lexical scanner.
type Lexer struct {
	r      *bufio.Reader
	line   int
	column int
}

// NewLexer returns a new instance of Lexer.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{r: bufio.NewReader(r), line: 1, column: 0}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() iter.Seq[Token] {
	return func(yield func(Token) bool) {
		for {
			ch := l.read()

			var tok Token
			tok.Line = l.line
			tok.Column = l.column

			switch {
			case ch == '\n':
				tok = Token{Type: NEXTLINE, Literal: string(ch), Line: l.line, Column: l.column}
				l.line++
				l.column = 0
			case isWhitespace(ch): // skip whitespace
				continue
			case isLetter(ch):
				l.unread()
				tok = l.readIdent()
				tok.Line = l.line

				tok.Column = l.column - len(tok.Literal)

				if _, isKeyword := Keywords[tok.Literal]; isKeyword {
					tok.Type = Keywords[tok.Literal]
				}
			case isDigit(ch):
				l.unread()
				tok = l.readNumber()
				tok.Line = l.line
				tok.Column = l.column - len(tok.Literal)
			case ch == '=':
				tok = Token{Type: ASSIGN, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == '+':
				tok = Token{Type: PLUS, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == '-':
				tok = Token{Type: MINUS, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == '*':
				tok = Token{Type: STAR, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == '/':
				tok = Token{Type: SLASH, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == '{':
				tok = Token{Type: LBRACE, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == '}':
				tok = Token{Type: RBRACE, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == '(':
				tok = Token{Type: LPAREN, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == ')':
				tok = Token{Type: RPAREN, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == ';':
				tok = Token{Type: SEMICOLON, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == ',':
				tok = Token{Type: COMMA, Literal: string(ch), Line: l.line, Column: l.column - 1}
			case ch == 0:
				tok = Token{Type: EOF, Literal: "", Line: 0, Column: 0}
			default:
				tok = Token{Type: ILLEGAL, Literal: string(ch), Line: 0, Column: 0}
			}

			if !yield(tok) {
				return
			}

			if tok.Type == EOF {
				return
			}
		}
	}
}

func (l *Lexer) read() rune {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return 0
	}
	l.column++
	return ch
}

func (l *Lexer) unread() {
	_ = l.r.UnreadRune()
	l.column--
}

func (l *Lexer) readIdent() Token {
	var buf bytes.Buffer
	buf.WriteRune(l.read())
	for {
		if ch := l.read(); isLetter(ch) {
			buf.WriteRune(ch)
		} else {
			l.unread()
			break
		}
	}
	return Token{Type: IDENT, Literal: buf.String()}
}

func (l *Lexer) readNumber() Token {
	var buf bytes.Buffer
	buf.WriteRune(l.read())
	for {
		if ch := l.read(); isDigit(ch) {
			buf.WriteRune(ch)
		} else {
			l.unread()
			break
		}
	}

	if l.peek() == '.' {
		buf.WriteRune(l.read())

		for {
			if ch := l.read(); isDigit(ch) {
				buf.WriteRune(ch)
			} else {
				l.unread()
				break
			}
		}

		return Token{Type: FLOAT, Literal: buf.String()}
	}

	return Token{Type: INT, Literal: buf.String()}
}

func (l *Lexer) readFloat() Token {
	var buf bytes.Buffer
	buf.WriteRune(l.read())
	for {
		if ch := l.read(); isDigit(ch) || ch == '.' {
			buf.WriteRune(ch)
		} else {
			l.unread()
			break
		}
	}
	return Token{Type: FLOAT, Literal: buf.String()}
}

func (l *Lexer) peek() rune {
	ch := l.read()
	l.unread()
	return ch
}

func isLetter(ch rune) bool {
	return strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", ch)
}

func isDigit(ch rune) bool {
	return strings.ContainsRune("0123456789", ch)
}

func isWhitespace(ch rune) bool {
	return strings.ContainsRune(" \t\r", ch)
}
