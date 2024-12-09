package parser

import (
	"errors"
	"fmt"
	"iter"
	"strconv"

	"github.com/EclesioMeloJunior/lotus/lexer"
)

var ErrVariableUndefined = errors.New("variable undefined")

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
	Type  Type
}

type ReassignVarStatement struct {
	VarName string
	Type    Type
	Value   Expression
}

type Argument struct {
	Name string
	Type Type
}

type FnStatement struct {
	Name                  string
	Args                  []*Argument
	Body                  []Node
	ReturnType            Type
	Defined               bool
	ExpressionsToEvaluate []Expression
}

type FnCall struct {
	FnName string
	Params []Expression
}

func (*FnCall) expressionNode() {}

// Identifier represents an identifier.
type UnboundedIdentifier struct {
	Value string
}

func (*UnboundedIdentifier) expressionNode() {}

// Identifier represents an identifier.
type Identifier struct {
	Value string
	Type  Type
}

func (*Identifier) expressionNode() {}

type StringLiteral struct {
	Value string
}

func (*StringLiteral) expressionNode() {}

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

	// TODO: currently all variables are "global" in the
	// parser's pov
	vars map[string]*VarStatement
	fns  map[string]*FnStatement
}

// NewParser returns a new instance of Parser.
func NewParser(tokens iter.Seq[lexer.Token]) *Parser {
	next, stop := iter.Pull(tokens)
	tokenStream := &TokenStream{next: next, stop: stop}

	p := &Parser{
		tokens: tokenStream,
		vars:   map[string]*VarStatement{},
		fns:    map[string]*FnStatement{},
	}
	p.nextToken()
	p.nextToken() // read two tokens, so curToken and peekToken are both set
	return p
}

// ErrParser represents a parsing error with line and column information.
type ErrParser struct {
	Line   int
	Column int
	Err    error
}

func (e *ErrParser) Error() string {
	return fmt.Sprintf("Error at line %d, column %d: %s", e.Line, e.Column, e.Err.Error())
}

// ParseProgram parses the tokens and returns a Program node.
func (p *Parser) ParseProgram() (*Program, error) {
	program := &Program{}
	for p.curToken.Type != lexer.EOF {
		if p.curToken.Type == lexer.NEXTLINE {
			p.nextToken()
			continue
		}

		stmt, err := p.parseStatement(Void)
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

func (p *Parser) parseStatement(tt Type) (Node, error) {
	switch p.curToken.Type {
	case lexer.VAR:
		return p.parseVarStatement()
	case lexer.FN:
		return p.parseFnStatement()
	case lexer.RETURN:
		return p.parseReturnStatement(tt)
	case lexer.IDENT:
		varStmt, exists := p.vars[p.curToken.Literal]
		// we are reassining a new value to the a already defined variable
		if exists {
			return p.parseReasignStatement(varStmt)
		}

		ident := &UnboundedIdentifier{Value: p.curToken.Literal}
		if err := p.consumeOrFail(lexer.LPAREN); err != nil {
			return nil, err
		}

		p.nextToken()
		expr, err := p.parseFnCallExpression(ident, Void)
		if err != nil {
			return nil, err
		}

		p.nextToken()
		if !endOfStatement(p.curToken.Type) {
			return nil, &ErrParser{
				Line:   p.curToken.Line,
				Column: p.curToken.Column + len(p.curToken.Literal),
				Err:    fmt.Errorf("after %s: expected end of statement or new line", p.curToken.Literal),
			}
		}
		return expr, nil
	default:
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column,
			Err:    fmt.Errorf("unexpected token: %s", p.curToken.Literal),
		}
	}
}

func (p *Parser) parseVarStatement() (*VarStatement, error) {
	stmt := &VarStatement{}

	if err := p.consumeOrFail(lexer.IDENT); err != nil {
		return nil, err
	}
	stmt.Name = p.curToken.Literal

	if p.peekToken.Type == lexer.COLON {
		p.nextToken()
		// TODO: should support custom types
		if err := p.consumeOrFail(lexer.RAWTYPE); err != nil {
			return nil, err
		}

		stmt.Type = getTypeFromLiteral(p.curToken.Literal)
	}

	if err := p.consumeOrFail(lexer.ASSIGN); err != nil {
		return nil, err
	}

	p.nextToken()
	expression, err := p.parseExpression(LOWEST, stmt.Type)
	if err != nil {
		return nil, err
	}

	stmt.Value = expression

	if !endOfStatement(p.peekToken.Type) {
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column + len(p.curToken.Literal),
			Err:    fmt.Errorf("after %s: expected end of statement or new line", p.curToken.Literal),
		}
	}

	p.vars[stmt.Name] = stmt

	p.nextToken()
	return stmt, nil
}

func (p *Parser) parseReasignStatement(old *VarStatement) (*ReassignVarStatement, error) {
	if err := p.consumeOrFail(lexer.ASSIGN); err != nil {
		return nil, err
	}

	reasign := &ReassignVarStatement{
		VarName: old.Name,
		Type:    old.Type,
	}

	p.nextToken()
	newExp, err := p.parseExpression(LOWEST, old.Type)
	if err != nil {
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column + len(p.curToken.Literal),
			Err:    err,
		}
	}

	reasign.Value = newExp

	if !endOfStatement(p.peekToken.Type) {
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column + len(p.curToken.Literal),
			Err:    fmt.Errorf("after %s: expected end of statement or new line", p.curToken.Literal),
		}
	}

	p.nextToken()

	return reasign, nil
}

func (p *Parser) parseFnStatement() (*FnStatement, error) {
	stmt := &FnStatement{}

	if err := p.consumeOrFail(lexer.IDENT); err != nil {
		return nil, err
	}

	stmt.Name = p.curToken.Literal

	if fnStmt, exists := p.fns[stmt.Name]; exists {
		if fnStmt.Defined {
			return nil, &ErrParser{
				Line:   p.curToken.Line,
				Column: p.curToken.Column,
				Err:    errors.New("function already defined"),
			}
		}
	}

	if err := p.consumeOrFail(lexer.LPAREN); err != nil {
		return nil, err
	}

	stmt.Args = []*Argument{}
	for {
		if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken()
			break
		}

		if err := p.consumeOrFail(lexer.IDENT); err != nil {
			return nil, err
		}

		arg := &Argument{}
		arg.Name = p.curToken.Literal

		if err := p.consumeOrFail(lexer.COLON); err != nil {
			return nil, err
		}

		if err := p.consumeOrFail(lexer.RAWTYPE); err != nil {
			return nil, err
		}

		arg.Type = getTypeFromLiteral(p.curToken.Literal)
		stmt.Args = append(stmt.Args, arg)

		if p.peekToken.Type == lexer.RPAREN {
			p.nextToken()
			break
		}

		if err := p.consumeOrFail(lexer.COMMA); err != nil {
			return nil, err
		}
	}

	if fnStmt, exists := p.fns[stmt.Name]; exists && !fnStmt.Defined {
		fmt.Println(len(fnStmt.ExpressionsToEvaluate))
		fmt.Println(len(stmt.Args))
		if len(fnStmt.ExpressionsToEvaluate) != len(stmt.Args) {
			return nil, &ErrParser{
				Line:   p.curToken.Line,
				Column: p.curToken.Column,
				Err:    errors.New("number of provided params do not match number of defined arguments"),
			}
		}

		for idx, arg := range stmt.Args {
			if err := arg.Type.Verify(fnStmt.ExpressionsToEvaluate[idx]); err != nil {
				return nil, &ErrParser{
					Line:   p.curToken.Line,
					Column: p.curToken.Column,
					Err:    err,
				}
			}
		}
	}

	mustHaveReturn := false
	if p.peekToken.Type == lexer.COLON {
		mustHaveReturn = true
		p.nextToken()
		// TODO: should support custom types
		if err := p.consumeOrFail(lexer.RAWTYPE); err != nil {
			return nil, err
		}

		stmt.ReturnType = getTypeFromLiteral(p.curToken.Literal)
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

		innerStmt, err := p.parseStatement(stmt.ReturnType)
		if err != nil {
			return nil, err
		}
		if innerStmt != nil {
			stmt.Body = append(stmt.Body, innerStmt)
		}

		p.nextToken()
	}

	if mustHaveReturn {
		if len(stmt.Body) == 0 {
			return nil, &ErrParser{
				Line:   p.curToken.Line,
				Column: p.curToken.Column,
				Err:    errors.New("function must have a return"),
			}
		}

		switch inner := stmt.Body[len(stmt.Body)-1].(type) {
		case *ReturnStatement:
			if inner.Type != stmt.ReturnType {
				return nil, &ErrParser{
					Line:   p.curToken.Line,
					Column: p.curToken.Column,
					Err:    fmt.Errorf("function expected return type: %T", stmt.ReturnType),
				}
			}
		default:
			return nil, &ErrParser{
				Line:   p.curToken.Line,
				Column: p.curToken.Column,
				Err:    errors.New("function must have a return"),
			}
		}
	}

	p.fns[stmt.Name] = stmt
	return stmt, nil
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Type  Type
	Value Expression
}

// parseReturnStatement parses a return statement.
func (p *Parser) parseReturnStatement(tt Type) (*ReturnStatement, error) {
	fmt.Printf("parsing return stmt: %v\n", tt)

	stmt := &ReturnStatement{}
	p.nextToken()
	expression, err := p.parseExpression(LOWEST, tt)
	if err != nil {
		return nil, err
	}

	stmt.Type = tt
	stmt.Value = expression

	if !endOfStatement(p.peekToken.Type) {
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column + len(p.curToken.Literal),
			Err:    fmt.Errorf("after %s: expected end of statement or new line", p.curToken.Literal),
		}
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

func (p *Parser) parseExpression(precedence int, tt Type) (Expression, error) {
	var leftExp Expression

	switch p.curToken.Type {
	case lexer.INT:
		leftExp = p.parseIntegerLiteral()
	case lexer.STRING:
		leftExp = &StringLiteral{Value: p.curToken.Literal}
	case lexer.FLOAT:
		leftExp = p.parseFloatLiteral()
	case lexer.IDENT:
		varStmt, ok := p.vars[p.curToken.Literal]
		if ok {
			leftExp = &Identifier{Value: varStmt.Name, Type: varStmt.Type}
		} else {
			leftExp = &UnboundedIdentifier{Value: varStmt.Name}
		}

	case lexer.LPAREN:
		expression, err := p.parseGroupedExpression(tt)
		if err != nil {
			return nil, err
		}
		leftExp = expression
	default:
		return nil, nil
	}

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		switch p.peekToken.Type {
		case lexer.PLUS, lexer.MINUS, lexer.SLASH, lexer.STAR:
			p.nextToken()
			exp, err := p.parseInfixExpression(leftExp, tt)
			if err != nil {
				return nil, err
			}

			leftExp = exp
		case lexer.LPAREN:
			p.nextToken()
			exp, err := p.parseFnCallExpression(leftExp, tt)
			if err != nil {
				return nil, err
			}
			leftExp = exp
		default:
			return leftExp, nil
		}
	}

	if err := tt.Verify(leftExp); err != nil {
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column,
			Err:    fmt.Errorf("verifying expression: %w", err),
		}
	}

	return leftExp, nil
}

func (p *Parser) parseGroupedExpression(tt Type) (Expression, error) {
	p.nextToken()
	exp, err := p.parseExpression(LOWEST, tt)
	if err != nil {
		return nil, err
	}

	if err := p.consumeOrFail(lexer.RPAREN); err != nil {
		return nil, err
	}

	return exp, nil
}

func (p *Parser) parseFnCallExpression(left Expression, tt Type) (Expression, error) {
	fnIdentifier, ok := left.(*UnboundedIdentifier)
	if !ok {
		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column,
			Err:    errors.New("expected function identifier"),
		}
	}

	// check if there is already a function definition for this function
	// check its return type, and if this matches with the incoming type tt
	fnStmt, ok := p.fns[fnIdentifier.Value]
	if !ok {
		// if there is not definition there is no problem
		// this function might be defined down in the source file
		// but here we need to define some infos about the function
		// to when finally the function is defined we do the type checking
		fnStmt = &FnStatement{
			Name:       fnIdentifier.Value,
			Body:       []Node{},
			ReturnType: tt,
			Defined:    false,
		}
	}

	fnCallExp := &FnCall{
		FnName: fnIdentifier.Value,
	}

	// is a simple call without parameterss
	if p.curToken.Type == lexer.RPAREN {
		p.nextToken()
		// save fn informations if the fn is not yet defined
		if !ok {
			p.fns[fnIdentifier.Value] = fnStmt
		}

		return fnCallExp, nil
	}

	// if the function is already defined we will
	// constraint the function call args using the types already defined
	if ok {
		fnCallExp.Params = make([]Expression, len(fnStmt.Args))
		for idx := 0; idx < len(fnStmt.Args); idx++ {
			arg, err := p.parseExpression(LOWEST, fnStmt.Args[idx].Type)
			if err != nil {
				return nil, &ErrParser{
					Line:   p.curToken.Line,
					Column: p.curToken.Column,
					Err:    errors.New("wrong paramater type"),
				}
			}
			fnCallExp.Params[idx] = arg

			if p.peekTokenIs(lexer.RPAREN) {
				p.nextToken()
				return fnCallExp, nil
			}

			if err := p.consumeOrFail(lexer.COMMA); err != nil {
				return nil, err
			}
		}

		return nil, &ErrParser{
			Line:   p.curToken.Line,
			Column: p.curToken.Column,
			Err:    errors.New("function call unexpected end"),
		}
	} else {
		// lets parse fn parameters, also updating the fn stmt
		// if the function is not yet defined
		for {
			fmt.Printf("curr: %v\n", p.curToken.Type.String())
			arg, err := p.parseExpression(LOWEST, Void)
			if err != nil {
				return nil, &ErrParser{
					Line:   p.curToken.Line,
					Column: p.curToken.Column,
					Err:    errors.New("wrong paramater type"),
				}
			}

			fmt.Printf("exp: %T\n", arg)
			fmt.Printf("peek: %v\n", p.peekToken.Type.String())

			fnCallExp.Params = append(fnCallExp.Params, arg)

			if p.peekTokenIs(lexer.RPAREN) {
				p.nextToken()
				break
			}

			if err := p.consumeOrFail(lexer.COMMA); err != nil {
				return nil, err
			}
		}

		fnStmt.ExpressionsToEvaluate = make([]Expression, len(fnCallExp.Params))
		copy(fnStmt.ExpressionsToEvaluate, fnCallExp.Params)
		p.fns[fnIdentifier.Value] = fnStmt

		return fnCallExp, nil
	}
}

func (p *Parser) parseInfixExpression(left Expression, tt Type) (Expression, error) {
	expression := &InfixExpression{
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	exp, err := p.parseExpression(precedence, tt)
	if err != nil {
		return nil, err
	}

	expression.Right = exp
	return expression, nil
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
			Err:    fmt.Errorf("expected: %s, got: %s", expected.String(), p.curToken.Type.String()),
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

func getTypeFromLiteral(literal string) Type {
	switch literal {
	case "int32":
		return Int32
	case "string":
		return String
	default:
		panic("unreacheable")
	}
}
