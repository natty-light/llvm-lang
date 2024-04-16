package parser

import (
	"fmt"
	"llvm-lang/ast"
	"llvm-lang/lexer"
	"llvm-lang/token"
	"strconv"
)

type (
	prefixParseFn func() ast.Expr
	infixParseFn  func(ast.Expr) ast.Expr
)

type Precedence int

const (
	LOWEST Precedence = iota + 1
	ANDOR             // I think this is right
	EQUALS
	LESSGREATEREQUAL
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
)

var precedences = map[token.TokenType]Precedence{
	token.And:                ANDOR,
	token.Or:                 ANDOR,
	token.EqualTo:            EQUALS,
	token.NotEqualTo:         EQUALS,
	token.LessThan:           LESSGREATER,
	token.GreaterThan:        LESSGREATER,
	token.GreaterThanEqualTo: LESSGREATEREQUAL,
	token.LessThanEqualTo:    LESSGREATEREQUAL,
	token.Plus:               SUM,
	token.Minus:              SUM,
	token.Slash:              PRODUCT,
	token.Star:               PRODUCT,
	token.Modulo:             PRODUCT,
	token.LeftParen:          CALL,
	token.LeftSquareBracket:  INDEX,
}

type Parser struct {
	lexer *lexer.Lexer

	currToken token.Token
	peekToken token.Token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{lexer: l, errors: make([]string, 0)}

	// peekToken and currToken are initialized to the zero value of token.Token, so we advance twice
	p.nextToken() // set peek
	p.nextToken() // set curr and peek

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)

	p.registerPrefix(token.Identifier, p.parseIdentifier)
	p.registerPrefix(token.Number, p.parseNumberLiteral)
	p.registerPrefix(token.Bang, p.parsePrefixExpr)
	p.registerPrefix(token.Minus, p.parsePrefixExpr)
	p.registerPrefix(token.LeftParen, p.parseGroupedExpr)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.Plus, p.parseInfixExpr)
	p.registerInfix(token.Minus, p.parseInfixExpr)
	p.registerInfix(token.Slash, p.parseInfixExpr)
	p.registerInfix(token.Star, p.parseInfixExpr)
	p.registerInfix(token.Modulo, p.parseInfixExpr)
	p.registerInfix(token.EqualTo, p.parseInfixExpr)
	p.registerInfix(token.NotEqualTo, p.parseInfixExpr)
	p.registerInfix(token.GreaterThanEqualTo, p.parseInfixExpr)
	p.registerInfix(token.LessThanEqualTo, p.parseInfixExpr)
	p.registerInfix(token.GreaterThan, p.parseInfixExpr)
	p.registerInfix(token.LessThan, p.parseInfixExpr)
	p.registerInfix(token.And, p.parseInfixExpr)
	p.registerInfix(token.Or, p.parseInfixExpr)
	p.registerInfix(token.LeftParen, p.parseCallExpr)
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("Honk! no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// advances current and peek by one
func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// Checks whether current token matches given type
func (p *Parser) currTokenIs(t token.TokenType) bool {
	return p.currToken.Type == t
}

// checks whether peek token matches given type
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// Checks if peek token matches given type, advances tokens if true
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken() // eats
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Honk! expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() Precedence {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) currPrecedence() Precedence {
	if p, ok := precedences[p.currToken.Type]; ok {
		return p
	}
	return LOWEST
}

// Parsing methods
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Stmts = make([]ast.Stmt, 0)

	for !p.currTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Stmts = append(program.Stmts, stmt)
		}
		p.nextToken() // advance past semis?
	}
	return program
}

// Statements
func (p *Parser) parseStatement() ast.Stmt {
	return p.parseExpressionStmt()
}

func (p *Parser) parseExpressionStmt() *ast.ExpressionStmt {
	stmt := &ast.ExpressionStmt{Token: p.currToken}

	stmt.Expr = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.Semicolon) {
		p.nextToken()
	}
	return stmt
}

// Expressions
func (p *Parser) parseExpression(precedence Precedence) ast.Expr {
	prefix := p.prefixParseFns[p.currToken.Type] // look for prefix function for p.currToken
	if prefix == nil {
		p.noPrefixParseFnError(p.currToken.Type)
		return nil
	}

	left := prefix() // call prefix function

	// if the statement has not ended and the passed in precedence is lower than the precedence of the next token
	// if the precedence of the next token is higher, then we need to parse it as an infix expression because it is higher priority
	// otherwise we return the expression as parsed by the prefix
	for !p.peekTokenIs(token.Semicolon) && precedence < p.peekPrecedence() {
		// look for an infix parse fn
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return left
		}

		p.nextToken()

		// we bind left to the infix expression
		left = infix(left)
	}

	return left
}

// prefix and infix functions

// this is an prefixParseFn, so it will not call p.nextToken() at the end
func (p *Parser) parseIdentifier() ast.Expr {
	return &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
}

// this is an prefixParseFn, so it will not call p.nextToken() at the end
func (p *Parser) parseNumberLiteral() ast.Expr {
	literal := &ast.NumberLiteral{Token: p.currToken}

	value, err := strconv.ParseFloat(p.currToken.Literal, 64)

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	literal.Value = value

	return literal
}

// this is an prefixParseFn, so it will not call p.nextToken() at the end
func (p *Parser) parsePrefixExpr() ast.Expr {
	expr := &ast.PrefixExpr{Token: p.currToken, Operator: p.currToken.Literal}

	p.nextToken() // advance past operator
	expr.Right = p.parseExpression(PREFIX)

	return expr
}

// this is an infixParseFn, so it will not call p.nextToken() at the end
func (p *Parser) parseInfixExpr(left ast.Expr) ast.Expr {
	expr := &ast.InfixExpr{Token: p.currToken, Operator: p.currToken.Literal, Left: left}

	precedence := p.currPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)

	return expr
}

// func (p *Parser) parseBooleanLiteral() ast.Expr {
// 	return &ast.BooleanLiteral{Token: p.currToken, Value: p.currTokenIs(token.True)}
// }

// this is a prefixParseFn
func (p *Parser) parseGroupedExpr() ast.Expr {
	p.nextToken() // advance past (

	expr := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RightParen) {
		return nil
	}

	return expr
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expr {
	list := []ast.Expr{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken() // advance past opening toke

	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.Comma) {
		p.nextToken() // advance to commma
		p.nextToken() // advance past comma
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseCallExpr(function ast.Expr) ast.Expr {
	expr := &ast.CallExpr{Token: p.currToken, Function: function}
	expr.Arguments = p.parseExpressionList(token.RightParen)
	return expr
}
