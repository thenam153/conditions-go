package evaluator

import (
	"github.com/thenam153/conditions-go/ast"
	lerrors "github.com/thenam153/conditions-go/errors"
)

func getBool(e ast.Expr) (bool, error) {
	switch n := e.(type) {
	case *ast.BooleanLiteral:
		return n.Value, nil
	default:
		return false, lerrors.Newf("Literal is not a boolean: %v", n)
	}
}

func getString(e ast.Expr) (string, error) {
	switch n := e.(type) {
	case *ast.StringLiteral:
		return n.Value, nil
	default:
		return "", lerrors.Newf("Literal is not a string: %v", n)
	}
}

func getNumber(e ast.Expr) (float64, error) {
	switch n := e.(type) {
	case *ast.NumberLiteral:
		return n.Value, nil
	default:
		return 0, lerrors.Newf("Literal is not a number: %v", n)
	}
}

func getSliceString(e ast.Expr) ([]string, error) {
	switch n := e.(type) {
	case *ast.SliceStringLiteral:
		return n.Value, nil
	default:
		return nil, lerrors.Newf("Literal is not a slice string: %v", n)
	}
}

func getSliceNumber(e ast.Expr) ([]float64, error) {
	switch n := e.(type) {
	case *ast.SliceNumberLiteral:
		return n.Value, nil
	default:
		return nil, lerrors.Newf("Literal is not a slice number: %v", n)
	}
}
