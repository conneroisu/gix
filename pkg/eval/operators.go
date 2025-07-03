package eval

import (
	"errors"
	"fmt"

	"github.com/conneroisu/gix/internal/types"
	"github.com/conneroisu/gix/internal/value"
)

// evalBinary evaluates binary operators.
func (e *Evaluator) evalBinary(expr *types.BinaryExpr, env value.Environment) (value.Value, error) {
	// Handle short-circuit operators
	switch expr.Op {
	case types.OpAnd:
		return e.evalAnd(expr, env)
	case types.OpOr:
		return e.evalOr(expr, env)
	case types.OpImpl:
		return e.evalImpl(expr, env)
	}

	// Evaluate both operands for other operators
	left, err := e.evalExpr(expr.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := e.evalExpr(expr.Right, env)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	// Arithmetic
	case types.OpAdd:
		return evalAdd(left, right)
	case types.OpSub:
		return evalSub(left, right)
	case types.OpMul:
		return evalMul(left, right)
	case types.OpDiv:
		return evalDiv(left, right)

	// String/List operations
	case types.OpConcat:
		return evalConcat(left, right)

	// Comparison
	case types.OpEq:
		return value.Bool(left.Equals(right)), nil
	case types.OpNEq:
		return value.Bool(!left.Equals(right)), nil
	case types.OpLT:
		return evalLess(left, right)
	case types.OpGT:
		return evalGreater(left, right)
	case types.OpLTE:
		return evalLessEq(left, right)
	case types.OpGTE:
		return evalGreaterEq(left, right)

	// Attribute set update
	case types.OpUpdate:
		return evalUpdate(left, right)

	default:
		return nil, fmt.Errorf("unknown binary operator: %v", expr.Op)
	}
}

// evalUnary evaluates unary operators.
func (e *Evaluator) evalUnary(expr *types.UnaryExpr, env value.Environment) (value.Value, error) {
	operand, err := e.evalExpr(expr.Expr, env)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case types.OpNot:
		b, ok := operand.(value.Bool)
		if !ok {
			return nil, fmt.Errorf("! operator requires boolean operand, got %v", operand.Type())
		}

		return value.Bool(!bool(b)), nil

	case types.OpNeg:
		switch v := operand.(type) {
		case value.Int:
			return value.Int(-v), nil
		case value.Float:
			return value.Float(-v), nil
		default:
			return nil, fmt.Errorf("- operator requires numeric operand, got %v", operand.Type())
		}

	default:
		return nil, fmt.Errorf("unknown unary operator: %v", expr.Op)
	}
}

// Short-circuit operators.
func (e *Evaluator) evalAnd(expr *types.BinaryExpr, env value.Environment) (value.Value, error) {
	left, err := e.evalExpr(expr.Left, env)
	if err != nil {
		return nil, err
	}

	leftBool, ok := left.(value.Bool)
	if !ok {
		return nil, fmt.Errorf("&& requires boolean operands, got %v", left.Type())
	}

	if !leftBool {
		return value.Bool(false), nil
	}

	right, err := e.evalExpr(expr.Right, env)
	if err != nil {
		return nil, err
	}

	rightBool, ok := right.(value.Bool)
	if !ok {
		return nil, fmt.Errorf("&& requires boolean operands, got %v", right.Type())
	}

	return rightBool, nil
}

func (e *Evaluator) evalOr(expr *types.BinaryExpr, env value.Environment) (value.Value, error) {
	left, err := e.evalExpr(expr.Left, env)
	if err != nil {
		return nil, err
	}

	leftBool, ok := left.(value.Bool)
	if !ok {
		return nil, fmt.Errorf("|| requires boolean operands, got %v", left.Type())
	}

	if leftBool {
		return value.Bool(true), nil
	}

	right, err := e.evalExpr(expr.Right, env)
	if err != nil {
		return nil, err
	}

	rightBool, ok := right.(value.Bool)
	if !ok {
		return nil, fmt.Errorf("|| requires boolean operands, got %v", right.Type())
	}

	return rightBool, nil
}

func (e *Evaluator) evalImpl(expr *types.BinaryExpr, env value.Environment) (value.Value, error) {
	left, err := e.evalExpr(expr.Left, env)
	if err != nil {
		return nil, err
	}

	leftBool, ok := left.(value.Bool)
	if !ok {
		return nil, fmt.Errorf("-> requires boolean operands, got %v", left.Type())
	}

	if !leftBool {
		return value.Bool(true), nil
	}

	right, err := e.evalExpr(expr.Right, env)
	if err != nil {
		return nil, err
	}

	rightBool, ok := right.(value.Bool)
	if !ok {
		return nil, fmt.Errorf("-> requires boolean operands, got %v", right.Type())
	}

	return rightBool, nil
}

// Arithmetic operations.
func evalAdd(left, right value.Value) (value.Value, error) {
	switch l := left.(type) {
	case value.Int:
		switch r := right.(type) {
		case value.Int:
			return value.Int(l + r), nil
		case value.Float:
			return value.Float(float64(l) + float64(r)), nil
		default:
			return nil, fmt.Errorf("cannot add %v to int", right.Type())
		}

	case value.Float:
		switch r := right.(type) {
		case value.Int:
			return value.Float(float64(l) + float64(r)), nil
		case value.Float:
			return value.Float(l + r), nil
		default:
			return nil, fmt.Errorf("cannot add %v to float", right.Type())
		}

	case value.String:
		if r, ok := right.(value.String); ok {
			return value.String(string(l) + string(r)), nil
		}

		return nil, fmt.Errorf("cannot add %v to string", right.Type())

	default:
		return nil, fmt.Errorf("cannot add values of type %v", left.Type())
	}
}

func evalSub(left, right value.Value) (value.Value, error) {
	switch l := left.(type) {
	case value.Int:
		switch r := right.(type) {
		case value.Int:
			return value.Int(int64(l) - int64(r)), nil
		case value.Float:
			return value.Float(float64(l) - float64(r)), nil
		default:
			return nil, fmt.Errorf("cannot subtract %v from int", right.Type())
		}

	case value.Float:
		switch r := right.(type) {
		case value.Int:
			return value.Float(float64(l) - float64(r)), nil
		case value.Float:
			return value.Float(l - r), nil
		default:
			return nil, fmt.Errorf("cannot subtract %v from float", right.Type())
		}

	default:
		return nil, fmt.Errorf("cannot subtract from %v", left.Type())
	}
}

func evalMul(left, right value.Value) (value.Value, error) {
	switch l := left.(type) {
	case value.Int:
		switch r := right.(type) {
		case value.Int:
			return value.Int(int64(l) * int64(r)), nil
		case value.Float:
			return value.Float(float64(l) * float64(r)), nil
		default:
			return nil, fmt.Errorf("cannot multiply int by %v", right.Type())
		}

	case value.Float:
		switch r := right.(type) {
		case value.Int:
			return value.Float(float64(l) * float64(r)), nil
		case value.Float:
			return value.Float(l * r), nil
		default:
			return nil, fmt.Errorf("cannot multiply float by %v", right.Type())
		}

	default:
		return nil, fmt.Errorf("cannot multiply %v", left.Type())
	}
}

func evalDiv(left, right value.Value) (value.Value, error) {
	// Check for division by zero
	switch r := right.(type) {
	case value.Int:
		if r == 0 {
			return nil, errors.New("division by zero")
		}
	case value.Float:
		if r == 0 {
			return nil, errors.New("division by zero")
		}
	}

	switch l := left.(type) {
	case value.Int:
		switch r := right.(type) {
		case value.Int:
			// Integer division in Nix returns float
			return value.Float(float64(l) / float64(r)), nil
		case value.Float:
			return value.Float(float64(l) / float64(r)), nil
		default:
			return nil, fmt.Errorf("cannot divide int by %v", right.Type())
		}

	case value.Float:
		switch r := right.(type) {
		case value.Int:
			return value.Float(float64(l) / float64(r)), nil
		case value.Float:
			return value.Float(l / r), nil
		default:
			return nil, fmt.Errorf("cannot divide float by %v", right.Type())
		}

	default:
		return nil, fmt.Errorf("cannot divide %v", left.Type())
	}
}

// List concatenation.
func evalConcat(left, right value.Value) (value.Value, error) {
	lList, lOk := left.(*value.List)
	rList, rOk := right.(*value.List)

	if !lOk || !rOk {
		return nil, fmt.Errorf(
			"++ operator requires two lists, got %v and %v",
			left.Type(),
			right.Type(),
		)
	}

	elements := append(lList.Elements(), rList.Elements()...)

	return value.NewList(elements...), nil
}

// Comparison operations.
func evalLess(left, right value.Value) (value.Value, error) {
	switch l := left.(type) {
	case value.Int:
		switch r := right.(type) {
		case value.Int:
			return value.Bool(l < r), nil
		case value.Float:
			return value.Bool(float64(l) < float64(r)), nil
		default:
			return nil, fmt.Errorf("cannot compare int with %v", right.Type())
		}

	case value.Float:
		switch r := right.(type) {
		case value.Int:
			return value.Bool(float64(l) < float64(r)), nil
		case value.Float:
			return value.Bool(l < r), nil
		default:
			return nil, fmt.Errorf("cannot compare float with %v", right.Type())
		}

	case value.String:
		if r, ok := right.(value.String); ok {
			return value.Bool(l < r), nil
		}

		return nil, fmt.Errorf("cannot compare string with %v", right.Type())

	default:
		return nil, fmt.Errorf("cannot compare %v", left.Type())
	}
}

func evalGreater(left, right value.Value) (value.Value, error) {
	// a > b is equivalent to b < a
	return evalLess(right, left)
}

func evalLessEq(left, right value.Value) (value.Value, error) {
	less, err := evalLess(left, right)
	if err != nil {
		return nil, err
	}
	if less.(value.Bool) {
		return value.Bool(true), nil
	}

	return value.Bool(left.Equals(right)), nil
}

func evalGreaterEq(left, right value.Value) (value.Value, error) {
	greater, err := evalGreater(left, right)
	if err != nil {
		return nil, err
	}
	if greater.(value.Bool) {
		return value.Bool(true), nil
	}

	return value.Bool(left.Equals(right)), nil
}

// Attribute set update.
func evalUpdate(left, right value.Value) (value.Value, error) {
	lAttrs, lOk := left.(*value.Attrs)
	rAttrs, rOk := right.(*value.Attrs)

	if !lOk || !rOk {
		return nil, fmt.Errorf(
			"// operator requires two attribute sets, got %v and %v",
			left.Type(),
			right.Type(),
		)
	}

	// Create new attribute set with merged attributes
	result := value.NewAttrs()

	// Copy left attributes
	for _, k := range lAttrs.Keys() {
		v, _ := lAttrs.Get(k)
		result.Set(k, v)
	}

	// Override with right attributes
	for _, k := range rAttrs.Keys() {
		v, _ := rAttrs.Get(k)
		result.Set(k, v)
	}

	return result, nil
}
