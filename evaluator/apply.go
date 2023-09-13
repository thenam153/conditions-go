package evaluator

import (
	"regexp"

	"github.com/thenam153/conditions-go/ast"
	lerrors "github.com/thenam153/conditions-go/errors"
	"github.com/thenam153/conditions-go/token"
)

func applyOperator(op token.Token, lhs, rhs ast.Expr) (*ast.BooleanLiteral, error) {
	switch op {
	case token.AND:
		return applyAND(lhs, rhs)
	case token.NAND:
		return applyNAND(lhs, rhs)
	case token.OR:
		return applyOR(lhs, rhs)
	case token.XOR:
		return applyXOR(lhs, rhs)
	case token.EQ:
		return applyEQ(lhs, rhs)
	case token.NEQ:
		return applyNEQ(lhs, rhs)
	case token.GT:
		return applyGT(lhs, rhs)
	case token.GTE:
		return applyGTE(lhs, rhs)
	case token.LT:
		return applyLT(lhs, rhs)
	case token.LTE:
		return applyLTE(lhs, rhs)
	case token.IN:
		return applyIN(lhs, rhs)
	case token.NOTIN:
		return applyNOTIN(lhs, rhs)
	case token.EREG:
		return applyEREG(lhs, rhs)
	case token.NEREG:
		return applyNEREG(lhs, rhs)
	default:
		return nil, lerrors.Newf("Not implemented operator, Op: %v", op.String())
	}
}

func applyAND(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv bool
		err    error
	)
	if lv, err = getBool(l); err != nil {
		return nil, err
	}
	if rv, err = getBool(r); err != nil {
		return nil, err
	}
	return &ast.BooleanLiteral{Value: lv && rv}, nil
}

func applyNAND(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv bool
		err    error
	)
	if lv, err = getBool(l); err != nil {
		return nil, err
	}
	if rv, err = getBool(r); err != nil {
		return nil, err
	}
	return &ast.BooleanLiteral{Value: !(lv && rv)}, nil
}

func applyOR(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv bool
		err    error
	)
	if lv, err = getBool(l); err != nil {
		return nil, err
	}
	if rv, err = getBool(r); err != nil {
		return nil, err
	}
	return &ast.BooleanLiteral{Value: lv || rv}, nil
}

func applyXOR(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv bool
		err    error
	)
	if lv, err = getBool(l); err != nil {
		return nil, err
	}
	if rv, err = getBool(r); err != nil {
		return nil, err
	}
	return &ast.BooleanLiteral{Value: (lv || rv) && (!lv || !rv)}, nil
}

func applyEQ(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lvs, rvs string
		lvn, rvn float64
		lvb, rvb bool
		err      error
	)
	if lvs, err = getString(l); err == nil {
		if rvs, err = getString(r); err != nil {
			return nil, lerrors.New("Cannot compare string with non-string")
		}
		return &ast.BooleanLiteral{Value: lvs == rvs}, nil
	}
	if lvn, err = getNumber(l); err == nil {
		if rvn, err = getNumber(r); err != nil {
			return nil, lerrors.New("Cannot compare number with non-number")
		}
		return &ast.BooleanLiteral{Value: lvn == rvn}, nil
	}
	if lvb, err = getBool(l); err == nil {
		if rvb, err = getBool(r); err != nil {
			return nil, lerrors.New("Cannot compare boolean with non-boolean")
		}
		return &ast.BooleanLiteral{Value: lvb == rvb}, nil
	}
	return &ast.BooleanLiteral{Value: false}, nil
}

func applyNEQ(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	result, err := applyEQ(l, r)
	if err != nil {
		return result, err
	}
	result.Value = !result.Value
	return result, nil
}

func applyEREG(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv string
		err    error
		match  bool
	)
	lv, err = getString(l)
	if err != nil {
		return nil, err
	}
	rv, err = getString(r)
	if err != nil {
		return nil, err
	}
	match, err = regexp.MatchString(rv, lv)
	return &ast.BooleanLiteral{Value: match}, err
}

func applyNEREG(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	result, err := applyEREG(l, r)
	if err != nil {
		return nil, err
	}
	result.Value = !result.Value
	return result, nil
}

func applyGT(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv float64
		err    error
	)
	if lv, err = getNumber(l); err != nil {
		return nil, err
	}
	if rv, err = getNumber(r); err != nil {
		return nil, err
	}
	return &ast.BooleanLiteral{Value: lv > rv}, nil
}

func applyGTE(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv float64
		err    error
	)
	if lv, err = getNumber(l); err != nil {
		return nil, err
	}
	if rv, err = getNumber(r); err != nil {
		return nil, err
	}
	return &ast.BooleanLiteral{Value: lv >= rv}, nil
}

func applyLT(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv float64
		err    error
	)
	if lv, err = getNumber(l); err != nil {
		return nil, err
	}
	if rv, err = getNumber(r); err != nil {
		return nil, err
	}
	return &ast.BooleanLiteral{Value: lv <= rv}, nil
}

func applyLTE(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		lv, rv float64
		err    error
	)
	if lv, err = getNumber(l); err != nil {
		return nil, err
	}
	if rv, err = getNumber(r); err != nil {
		return nil, err
	}
	return &ast.BooleanLiteral{Value: lv <= rv}, nil
}

func applyIN(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	var (
		err error
	)
	switch t := l.(type) {
	case *ast.StringLiteral:
		var (
			lv string
			rv []string
		)
		lv, _ = getString(l)
		if rv, err = getSliceString(r); err != nil {
			return nil, err
		}
		for _, v := range rv {
			if lv == v {
				return &ast.BooleanLiteral{Value: true}, nil
			}
		}
		return &ast.BooleanLiteral{Value: false}, nil
	case *ast.NumberLiteral:
		var (
			lv float64
			rv []float64
		)
		lv, _ = getNumber(l)
		if rv, err = getSliceNumber(r); err != nil {
			return nil, err
		}
		for _, v := range rv {
			if lv == v {
				return &ast.BooleanLiteral{Value: true}, nil
			}
		}
		return &ast.BooleanLiteral{Value: false}, nil
	default:
		return nil, lerrors.Newf("Cannot evaluate Literal of unknow type %s %T", t, t)
	}
}

func applyNOTIN(l, r ast.Expr) (*ast.BooleanLiteral, error) {
	result, err := applyIN(l, r)
	if err != nil {
		return result, err
	}
	result.Value = !result.Value
	return result, err
}
