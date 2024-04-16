package ast

import (
	"bytes"
	"llvm-lang/token"
	"strings"
)

// Interfaces

type (
	Node interface {
		TokenLiteral() string
		String() string
	}

	Stmt interface {
		Node
		statementNode()
	}

	Expr interface {
		Node
		expressionNode()
	}
)

// Node
type (
	Program struct {
		Stmts []Stmt
	}
)

// Statements
type (
	ExpressionStmt struct {
		Token token.Token
		Expr  Expr
	}
)

// Expressions and literals
type (
	// Literals
	NumberLiteral struct {
		Token token.Token
		Value float64
	}

	// Expressions
	Identifier struct {
		Token token.Token // token.Ident
		Value string
	}

	PrefixExpr struct {
		Token    token.Token
		Operator string
		Right    Expr
	}

	InfixExpr struct {
		Token    token.Token
		Left     Expr
		Operator string
		Right    Expr
	}

	CallExpr struct {
		Token     token.Token
		Function  Expr
		Arguments []Expr
	}
)

// Node interfaces
func (p *Program) TokenLiteral() string {
	if len(p.Stmts) > 0 {
		return p.Stmts[0].TokenLiteral()
	} else {
		return ""
	}
}

func (e *ExpressionStmt) TokenLiteral() string {
	return e.Token.Literal
}

func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *NumberLiteral) TokenLiteral() string {
	return i.Token.Literal
}

func (p *PrefixExpr) TokenLiteral() string {
	return p.Token.Literal
}

func (i *InfixExpr) TokenLiteral() string {
	return i.Token.Literal
}

func (c *CallExpr) TokenLiteral() string {
	return c.Token.Literal
}

// Statements
func (p *Program) String() string {
	var out bytes.Buffer

	for _, stmt := range p.Stmts {
		out.WriteString(stmt.String())
	}

	return out.String()
}

func (e *ExpressionStmt) String() string {
	if e.Expr != nil {
		return e.Expr.String()
	}
	return ""
}

// Expressions
func (i *Identifier) String() string {
	return i.Value
}

func (p *PrefixExpr) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(p.Operator)
	out.WriteString(p.Right.String())
	out.WriteString(")")

	return out.String()
}

func (i *InfixExpr) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(i.Left.String())
	out.WriteString(" " + i.Operator + " ")
	out.WriteString(i.Right.String())
	out.WriteString(")")

	return out.String()
}

func (c *CallExpr) String() string {
	var out bytes.Buffer
	args := make([]string, 0)
	for _, arg := range c.Arguments {
		args = append(args, arg.String())
	}

	out.WriteString(c.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// Literals
func (i *NumberLiteral) String() string {
	return i.Token.Literal
}

// Statements
func (e *ExpressionStmt) statementNode() {}

// Expressions
func (i *Identifier) expressionNode()    {}
func (n *NumberLiteral) expressionNode() {}
func (p *PrefixExpr) expressionNode()    {}
func (i *InfixExpr) expressionNode()     {}
func (c *CallExpr) expressionNode()      {}
