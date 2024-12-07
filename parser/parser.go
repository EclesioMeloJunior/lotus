package parser

import (
	"fmt"
	"iter"
	"strconv"

	"github.com/EclesioMeloJunior/lotus/lexer"
)

// Node represents a node in the AST.
type Node interface{}

// Expression represents an expression node.
type Expression interface {
	Node
	expressionNode()
}

// Program represents the root node of the AST.
type Program struct {
	Statements []Node
}

// VarStatement represents a variable declaration.
type VarStatement struct {
	Name  string
	Value Expression
}

type FnStatement struct {
	Name string
	Args []string
	Body []Node
}

// Identifier represents an identifier.
type Identifier struct {
	Value string
}

func (*Identifier) expressionNode() {}

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	Value int64
}

func (*IntegerLiteral) expressionNode() {}

// FloatLiteral represents a float literal.
type FloatLiteral struct {
	Value float64
}

func (*FloatLiteral) expressionNode() {}

// InfixExpression represents an infix expression.
type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (*InfixExpression) expressionNode() {}

// PrefixExpression represents a prefix expression.
type PrefixExpression struct {
	Operator string
	Right    Expression
}

func (*PrefixExpression) expressionNode() {}

type TokenStream struct {
	next func() (lexer.Token, bool)
	stop func()
}

// Parser represents a parser.
type Parser struct {
	tokens    *TokenStream
	curToken  lexer.Token
	peekToken lexer.Token
}

// NewParser returns a new instance of Parser.
func NewParser(tokens iter.Seq[lexer.Token]) *Parser {
	next, stop := iter.Pull(tokens)
	tokenStream := &TokenStream{next: next, stop: stop}

	p := &Parser{tokens: tokenStream}
	p.nextToken()
	p.nextToken() // read two tokens, so curToken and peekToken are both set
	return p
}

// ErrParser represents a parsing error with line and column information.
type ErrParser struct {
	Line   int
	Column int
	Msg    string
}

func (e *ErrParser) Error() string {
	return fmt.Sprintf("Error at line %d, column %d: %s", e.Line, e.Column, e.Msg)
}

// ParseProgram parses the tokens and returns a Program node.
func (p *Parser) ParseProgram() (*Program, error) {
	program := &Program{}
	for p.curToken.Type != lexer.EOF {
		if p.curToken.Type == lexer.NEXTLINE {
			p.nextToken()
			continue
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program, nil
}

func (p *Parser) parseStatement() (Node, error) {
	switch p.curToken.Type {
	case lexer.VAR:
		return p.parseVarStatement()
	case lexer.FN:
		return p.parseFnStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	default:
		return nil, &ErrParser{Line: p.curToken.Line, Column: p.curToken.Column, Msg: fmt.Sprintf("unexpected token: %s", p.curToken.Literal)}
	}
}

func (p *Parser) parseVarStatement() (*VarStatement, error) {
	stmt := &VarStatement{}

	if err := p.consumeOrFail(lexer.IDENT); err != nil {
		return nil, err
	}
	stmt.Name = p.curToken.Literal

	if err := p.consumeOrFail(lexer.ASSIGN); err != nil {
		return nil, err
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if !endOfStatement(p.peekToken.Type) {
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column + len(p.curToken.Literal),
			Msg:    fmt.Sprintf("after %s: expected end of statement or new line", p.curToken.Literal)}
	}

	p.nextToken()
	return stmt, nil
}

func (p *Parser) parseFnStatement() (*FnStatement, error) {
	stmt := &FnStatement{}

	if err := p.consumeOrFail(lexer.IDENT); err != nil {
		return nil, err
	}

	stmt.Name = p.curToken.Literal

	if err := p.consumeOrFail(lexer.LPAREN); err != nil {
		return nil, err
	}

	stmt.Args = []string{}
	for {
		p.nextToken()
		if p.curToken.Type == lexer.RPAREN {
			break
		}
		if p.curToken.Type != lexer.IDENT {
			return nil, &ErrParser{Line: p.curToken.Line, Column: p.curToken.Column, Msg: "expected identifier"}
		}

		stmt.Args = append(stmt.Args, p.curToken.Literal)

		if p.peekToken.Type == lexer.RPAREN {
			p.nextToken()
			break
		}

		if err := p.consumeOrFail(lexer.COMMA); err != nil {
			return nil, err
		}
	}

	if err := p.consumeOrFail(lexer.LBRACE); err != nil {
		return nil, err
	}

	stmt.Body = []Node{}

	p.nextToken()
	for p.curToken.Type != lexer.RBRACE {
		if p.curToken.Type == lexer.NEXTLINE {
			p.nextToken()
			continue
		}

		innerStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmt.Body = append(stmt.Body, innerStmt)
		}

		p.nextToken()
	}

	return stmt, nil
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Value Expression
}

// parseReturnStatement parses a return statement.
func (p *Parser) parseReturnStatement() (*ReturnStatement, error) {
	stmt := &ReturnStatement{}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if !endOfStatement(p.peekToken.Type) {
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column + len(p.curToken.Literal),
			Msg:    fmt.Sprintf("after %s: expected end of statement or new line", p.curToken.Literal)}
	}

	p.nextToken()
	return stmt, nil
}

const (
	_ int = iota
	LOWEST
	SUM     // + or -
	PRODUCT // * or /
	PREFIX  // -X or !X
	CALL    // myFunction(X)
)

var precedences = map[lexer.TokenType]int{
	lexer.PLUS:   SUM,
	lexer.MINUS:  SUM,
	lexer.SLASH:  PRODUCT,
	lexer.STAR:   PRODUCT,
	lexer.LPAREN: CALL,
}

func (p *Parser) parseExpression(precedence int) Expression {
	var leftExp Expression

	switch p.curToken.Type {
	case lexer.INT:
		leftExp = p.parseIntegerLiteral()
	case lexer.FLOAT:
		leftExp = p.parseFloatLiteral()
	case lexer.IDENT:
		leftExp = &Identifier{Value: p.curToken.Literal}
	case lexer.LPAREN:
		leftExp = p.parseGroupedExpression()
	default:
		return nil
	}

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		switch p.peekToken.Type {
		case lexer.PLUS, lexer.MINUS, lexer.SLASH, lexer.STAR:
			p.nextToken()
			leftExp = p.parseInfixExpression(leftExp)
		default:
			return leftExp
		}
	}

	return leftExp
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Left:     left,
		Operator: p.curToken.Literal,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseIntegerLiteral() *IntegerLiteral {
	value, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
	return &IntegerLiteral{Value: value}
}

func (p *Parser) parseFloatLiteral() *FloatLiteral {
	value, _ := strconv.ParseFloat(p.curToken.Literal, 64)
	return &FloatLiteral{Value: value}
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	if p.curToken.Type == lexer.EOF {
		return
	}

	if tok, ok := p.tokens.next(); ok {
		p.peekToken = tok
	} else {
		p.tokens.stop()
		p.peekToken = lexer.Token{Type: lexer.EOF}
	}
}

func (p *Parser) consumeOrFail(expected lexer.TokenType) error {
	p.nextToken()
	if p.curToken.Type != expected {
		return &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column,
			Msg:    fmt.Sprintf("expected: %v, got: %v", expected, p.curToken.Type),
		}
	}
	return nil
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	return false
}

func endOfStatement(t lexer.TokenType) bool {
	return t == lexer.SEMICOLON || t == lexer.NEXTLINE || t == lexer.EOF
}
