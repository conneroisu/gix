package eval

import (
	"errors"
	"fmt"

	"github.com/conneroisu/gix/internal/types"
	"github.com/conneroisu/gix/internal/value"
)

// evalApply evaluates function application.
func (e *Evaluator) evalApply(expr *types.ApplyExpr, env value.Environment) (value.Value, error) {
	// Evaluate the function
	fnVal, err := e.evalExpr(expr.Func, env)
	if err != nil {
		return nil, err
	}

	// Evaluate the argument
	argVal, err := e.evalExpr(expr.Arg, env)
	if err != nil {
		return nil, err
	}

	// Apply based on function type
	switch fn := fnVal.(type) {
	case *value.Function:
		// Create new environment for function body
		fnEnv := fn.Env().Extend()
		fnEnv.Set(fn.Param(), argVal)

		// Evaluate function body
		body, ok := fn.Body().(types.Expr)
		if !ok {
			return nil, errors.New("invalid function body")
		}

		return e.evalExpr(body, fnEnv)

	case *value.Builtin:
		// Builtin functions expect a list of arguments
		return fn.Apply([]value.Value{argVal})

	default:
		return nil, fmt.Errorf("cannot apply non-function value of type %v", fnVal.Type())
	}
}

// evalSelect evaluates attribute selection.
func (e *Evaluator) evalSelect(expr *types.SelectExpr, env value.Environment) (value.Value, error) {
	// Evaluate the expression to select from
	val, err := e.evalExpr(expr.Expr, env)
	if err != nil {
		return nil, err
	}

	// Navigate through the attribute path
	current := val
	for i, key := range expr.AttrPath {
		attrs, ok := current.(*value.Attrs)
		if !ok {
			// If we have a default value, use it
			if expr.Default != nil {
				return e.evalExpr(expr.Default, env)
			}

			return nil, fmt.Errorf("cannot select attribute '%s' from %v", key, current.Type())
		}

		next, ok := attrs.Get(key)
		if !ok {
			// If we have a default value, use it
			if expr.Default != nil {
				return e.evalExpr(expr.Default, env)
			}

			return nil, fmt.Errorf("attribute '%s' not found", key)
		}

		// For the last key, return the value
		if i == len(expr.AttrPath)-1 {
			return next, nil
		}

		current = next
	}

	// This shouldn't happen
	return nil, errors.New("unexpected end of attribute path")
}

// evalHasAttr evaluates attribute existence test.
func (e *Evaluator) evalHasAttr(
	expr *types.HasAttrExpr,
	env value.Environment,
) (value.Value, error) {
	// Evaluate the expression to test
	val, err := e.evalExpr(expr.Expr, env)
	if err != nil {
		return nil, err
	}

	// Navigate through the attribute path
	current := val
	for i, key := range expr.AttrPath {
		attrs, ok := current.(*value.Attrs)
		if !ok {
			return value.Bool(false), nil
		}

		next, ok := attrs.Get(key)
		if !ok {
			return value.Bool(false), nil
		}

		// For the last key, we found it
		if i == len(expr.AttrPath)-1 {
			return value.Bool(true), nil
		}

		current = next
	}

	return value.Bool(true), nil
}
