package eval

import (
	"errors"
	"fmt"

	"github.com/conneroisu/gix/internal/types"
	"github.com/conneroisu/gix/internal/value"
)

// evalIf evaluates an if-then-else expression.
func (e *Evaluator) evalIf(expr *types.IfExpr, env value.Environment) (value.Value, error) {
	cond, err := e.evalExpr(expr.Cond, env)
	if err != nil {
		return nil, err
	}

	condBool, ok := cond.(value.Bool)
	if !ok {
		return nil, fmt.Errorf("if condition must be boolean, got %v", cond.Type())
	}

	if condBool {
		return e.evalExpr(expr.Then, env)
	} else {
		return e.evalExpr(expr.Else, env)
	}
}

// evalLet evaluates a let expression.
func (e *Evaluator) evalLet(expr *types.LetExpr, env value.Environment) (value.Value, error) {
	// Create new environment for let bindings
	letEnv := env.Extend()

	// Evaluate all bindings in order
	// Note: In real Nix, let bindings can be mutually recursive
	// For simplicity, we evaluate them in order
	for _, binding := range expr.Bindings {
		val, err := e.evalExpr(binding.Value, letEnv)
		if err != nil {
			return nil, fmt.Errorf("error in let binding %s: %w", binding.Name, err)
		}
		letEnv.Set(binding.Name, val)
	}

	// Evaluate body in the new environment
	return e.evalExpr(expr.Body, letEnv)
}

// evalWith evaluates a with expression.
func (e *Evaluator) evalWith(expr *types.WithExpr, env value.Environment) (value.Value, error) {
	// Evaluate the expression that provides the scope
	scopeVal, err := e.evalExpr(expr.Expr, env)
	if err != nil {
		return nil, err
	}

	// It must be an attribute set
	attrs, ok := scopeVal.(*value.Attrs)
	if !ok {
		return nil, fmt.Errorf("with expression requires attribute set, got %v", scopeVal.Type())
	}

	// Create new environment with attributes from the set
	withEnv := env.Extend()
	for _, key := range attrs.Keys() {
		val, _ := attrs.Get(key)
		withEnv.Set(key, val)
	}

	// Evaluate body in the new environment
	return e.evalExpr(expr.Body, withEnv)
}

// evalAssert evaluates an assert expression.
func (e *Evaluator) evalAssert(expr *types.AssertExpr, env value.Environment) (value.Value, error) {
	// Evaluate the assertion condition
	cond, err := e.evalExpr(expr.Cond, env)
	if err != nil {
		return nil, err
	}

	condBool, ok := cond.(value.Bool)
	if !ok {
		return nil, fmt.Errorf("assert condition must be boolean, got %v", cond.Type())
	}

	if !condBool {
		return nil, errors.New("assertion failed")
	}

	// If assertion passes, evaluate the body
	return e.evalExpr(expr.Body, env)
}
