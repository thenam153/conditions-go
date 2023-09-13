package ast

import (
	"github.com/thenam153/conditions-go/token"
)

type Expr interface {
	Node
}

type Node interface{}

type BinaryExpr struct {
	LHS Expr
	RHS Expr
	OP  token.Token
}

type ParenExpr struct {
	Expr Expr
}

type VarRef struct {
	Value string
}

type StringLiteral struct {
	Value string
}

type NumberLiteral struct {
	Value float64
}

type BooleanLiteral struct {
	Value bool
}

type SliceStringLiteral struct {
	Value []string
}

type SliceNumberLiteral struct {
	Value []float64
}
