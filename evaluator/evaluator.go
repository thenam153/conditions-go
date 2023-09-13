package evaluator

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/thenam153/conditions-go/ast"
	lerrors "github.com/thenam153/conditions-go/errors"
)

func Evaluate(expr ast.Expr, args map[string]any) (bool, error) {
	expr, err := evaluateTree(expr, args)
	if err != nil {
		return false, lerrors.NewWrap("Cannot evaluate expression", err)
	}
	if v, ok := expr.(*ast.BooleanLiteral); ok {
		return v.Value, nil
	}
	return false, lerrors.Newf("Wrong root expression, cannot return boolean value, type: %T", expr)
}

func evaluateTree(expr ast.Expr, args map[string]any) (ast.Expr, error) {
	if expr == nil || reflect.ValueOf(expr).IsNil() {
		return nil, lerrors.New("Expression must be not nil")
	}
	switch e := expr.(type) {
	case *ast.ParenExpr:
		return evaluateTree(e.Expr, args)
	case *ast.BinaryExpr:
		var (
			elhs, erhs ast.Expr
			err        error
		)
		if elhs, err = evaluateTree(e.LHS, args); err != nil {
			return nil, lerrors.NewWrap("Cannot evaluate LHS of binary expression", err)
		}
		if erhs, err = evaluateTree(e.RHS, args); err != nil {
			return nil, lerrors.NewWrap("Cannot evaluate RHS of binary expression", err)
		}
		return applyOperator(e.OP, elhs, erhs)
	case *ast.VarRef:
		index := e.Value
		if _, ok := args[index]; !ok {
			return nil, lerrors.Newf("Cannot get args with index %v", index)
		}
		kind := reflect.TypeOf(args[index]).Kind()
		switch kind {
		case reflect.Int:
			return &ast.NumberLiteral{Value: float64(args[index].(int))}, nil
		case reflect.Int32:
			return &ast.NumberLiteral{Value: float64(args[index].(int32))}, nil
		case reflect.Int64:
			return &ast.NumberLiteral{Value: float64(args[index].(int64))}, nil
		case reflect.Float32:
			return &ast.NumberLiteral{Value: float64(args[index].(float32))}, nil
		case reflect.Float64:
			return &ast.NumberLiteral{Value: float64(args[index].(float64))}, nil
		case reflect.String:
			return &ast.StringLiteral{Value: args[index].(string)}, nil
		case reflect.Bool:
			return &ast.BooleanLiteral{Value: args[index].(bool)}, nil
		case reflect.Slice, reflect.Array:
			kindEl := reflect.TypeOf(args[index]).Elem().Kind()
			switch kindEl {
			case reflect.Int:
				arrValue := args[index].([]int)
				arrNumber := make([]float64, len(arrValue))
				for _, v := range arrValue {
					arrNumber = append(arrNumber, float64(v))
				}
				return &ast.SliceNumberLiteral{Value: arrNumber}, nil
			case reflect.Int32:
				arrValue := args[index].([]int32)
				arrNumber := make([]float64, len(arrValue))
				for _, v := range arrValue {
					arrNumber = append(arrNumber, float64(v))
				}
				return &ast.SliceNumberLiteral{Value: arrNumber}, nil
			case reflect.Int64:
				arrValue := args[index].([]int64)
				arrNumber := make([]float64, len(arrValue))
				for _, v := range arrValue {
					arrNumber = append(arrNumber, float64(v))
				}
				return &ast.SliceNumberLiteral{Value: arrNumber}, nil
			case reflect.Float32:
				arrValue := args[index].([]float32)
				arrNumber := make([]float64, len(arrValue))
				for _, v := range arrValue {
					arrNumber = append(arrNumber, float64(v))
				}
				return &ast.SliceNumberLiteral{Value: arrNumber}, nil
			case reflect.Float64:
				arrNumber := args[index].([]float64)
				return &ast.SliceNumberLiteral{Value: arrNumber}, nil
			case reflect.String:
				arrString := args[index].([]string)
				return &ast.SliceStringLiteral{Value: arrString}, nil
			default:
				arrString := args[index].([]string)
				return &ast.SliceStringLiteral{Value: arrString}, nil
			}
		}
		return nil, lerrors.Newf("Cannot parse %T to array", expr)
	case *ast.JQRef:
		var (
			bytes  []byte
			err    error
			value  any
			values []any
		)
		if bytes, err = json.Marshal(args); err != nil {
			return nil, lerrors.Wrap(fmt.Errorf("cannot marshal %T to JSON", expr), err)
		}
		if err = json.Unmarshal(bytes, &value); err != nil {
			return nil, lerrors.Wrap(fmt.Errorf("cannot unmarshal %T to any", expr), err)
		}
		iter := e.Query.Run(value)
		for {
			v, n := iter.Next()
			values = append(values, v)
			if !n {
				break
			}
		}
		jqMode, ok := ast.JQModes[e.Mode]
		if !ok {
			jqMode = ast.JQFirst
		}
		switch jqMode {
		case ast.JQFirst:
			switch v := values[0].(type) {
			case float64:
				return &ast.NumberLiteral{Value: v}, nil
			case string:
				return &ast.StringLiteral{Value: v}, nil
			case nil:
				return nil, lerrors.New("JQ Query get nil value")
			default:
				return nil, lerrors.Newf("JQ unsupported type %T", v)
			}
		case ast.JQLast:
			switch v := values[len(values)-1].(type) {
			case float64:
				return &ast.NumberLiteral{Value: v}, nil
			case string:
				return &ast.StringLiteral{Value: v}, nil
			case nil:
				return nil, lerrors.New("JQ Query get nil value")
			default:
				return nil, lerrors.Newf("JQ unsupported type %T", v)
			}
		case ast.JQArray:
			switch v := values[0].(type) {
			case float64:
				arrayNumber := []float64{}
				for _, v := range values {
					arrayNumber = append(arrayNumber, v.(float64))
				}
				return &ast.SliceNumberLiteral{Value: arrayNumber}, nil
			case string:
				arrayString := []string{}
				for _, v := range values {
					arrayString = append(arrayString, v.(string))
				}
				return &ast.SliceStringLiteral{Value: arrayString}, nil
			case nil:
				return nil, lerrors.New("JQ Query get nil value")
			default:
				return nil, lerrors.Newf("JQ unsupported type %T", v)
			}
		default:
			return nil, lerrors.Newf("Not implemented JQMode, JQMode: %v", jqMode)
		}
	}
	return expr, nil
}
