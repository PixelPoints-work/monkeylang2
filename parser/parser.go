package parser

import (
	"fmt"
	"monkeylang2/ast"
	"monkeylang2/lexer"
	"monkeylang2/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.peekToken = p.l.NextToken()
	p.nextToken()
	return p
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	fmt.Printf("ParseProgram: starting with current token type=%s literal=%s\n", p.curToken.Type, p.curToken.Literal)
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Add this debug print
	fmt.Printf("ParseProgram: is current token EOF? %v\n", p.curToken.Type == token.EOF)

	for p.curToken.Type != token.EOF {
		fmt.Printf("ParseProgram loop: current token type=%s literal=%s\n", p.curToken.Type, p.curToken.Literal)
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
			fmt.Printf("ParseProgram: added statement of type %T\n", stmt)
		} else {
			fmt.Printf("ParseProgram: got nil statement\n")
		}
		p.nextToken()
		fmt.Printf("ParseProgram: after nextToken type=%s literal=%s\n", p.curToken.Type, p.curToken.Literal)
	}
	fmt.Printf("ParseProgram: ended with %d statements\n", len(program.Statements))
	return program
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseStatement() ast.Statement {
	fmt.Printf("parseStatement ENTRY: current token type=%s literal=%s\n", p.curToken.Type, p.curToken.Literal)
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: We're skipping the expressions until we
	// encounter a semicolon
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	// TODO We're skipping the expressions until we
	// encounter a semicolon

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	fmt.Printf("parseExpressionStatement: current token type=%s literal=%s\n", p.curToken.Type, p.curToken.Literal)
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	expr := p.parseExpression(LOWEST)
	fmt.Printf("parseExpressionStatement: expression result: %v\n", expr)
	stmt.Expression = expr

	if stmt.Expression == nil {
		fmt.Printf("parseExpressionStatement: got nil expression\n")
	} else {
		fmt.Printf("parseExpressionStatement: got expression of type %T\n", stmt.Expression)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	fmt.Printf("parseExpression: current token type=%s literal=%s\n", p.curToken.Type, p.curToken.Literal)
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		fmt.Printf("parseExpression: no prefix parse function for %s found\n", p.curToken.Type)
		fmt.Printf("parseExpression: registered prefixes: %v\n", p.prefixParseFns)
		return nil
	}
	leftExp := prefix()
	fmt.Printf("parseExpression: leftExp result type: %T\n", leftExp)
	return leftExp
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	fmt.Printf("parseIntegerLiteral: trying to parse %s\n", p.curToken.Literal)
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		fmt.Printf("parseIntegerLiteral: error parsing: %v\n", err)
		return nil
	}
	lit.Value = value
	fmt.Printf("parseIntegerLiteral: successfully parsed %d\n", value)
	return lit
}
