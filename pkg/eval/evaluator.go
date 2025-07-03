package eval

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/conneroisu/gix/internal/types"
	"github.com/conneroisu/gix/internal/value"
)

// Evaluator implements the semantic evaluation engine for Nix expressions.
// It traverses Abstract Syntax Trees (ASTs) and computes their runtime values,
// implementing the complete Nix evaluation semantics including scoping,
// function application, and built-in operations.
type Evaluator struct {
	baseDir  string                 // Base directory for resolving relative paths
	builtins map[string]value.Value // Built-in functions and constants
}

// New creates a new evaluator instance with the specified base directory.
// The base directory is used for resolving relative path literals in expressions.
// The evaluator is initialized with all standard built-in functions and constants
// that form the Nix standard library.
func New(baseDir string) *Evaluator {
	e := &Evaluator{
		baseDir:  baseDir,                      // Store base directory for path resolution
		builtins: make(map[string]value.Value), // Initialize built-ins registry
	}
	// Populate the built-ins registry with standard functions
	e.registerBuiltins()

	return e
}

// Eval evaluates a Nix expression in a fresh environment.
// This is the main entry point for expression evaluation, creating a new
// environment populated with built-in functions and constants.
// Used for evaluating top-level expressions and standalone evaluations.
func (e *Evaluator) Eval(expr types.Expr) (value.Value, error) {
	// Create a fresh environment for this evaluation
	env := value.NewEnv()

	// Populate environment with all built-in functions and constants
	// This makes functions like 'length', 'map', etc. available to expressions
	for name, builtin := range e.builtins {
		env.Set(name, builtin)
	}

	// Delegate to the main evaluation dispatcher
	return e.evalExpr(expr, env)
}

// EvalWithEnv evaluates an expression in an existing environment.
// This method is used for evaluating expressions within a specific context,
// such as function bodies (which capture their defining environment) or
// expressions within let bindings where variables are already bound.
func (e *Evaluator) EvalWithEnv(expr types.Expr, env value.Environment) (value.Value, error) {
	// Use the provided environment directly without modification
	return e.evalExpr(expr, env)
}

// evalExpr is the central evaluation dispatcher that implements the Nix evaluation semantics.
// It performs pattern matching on AST node types and delegates to specialized evaluation
// methods. This tree-walking evaluator processes expressions recursively, maintaining
// proper scoping and environment management throughout the evaluation process.
func (e *Evaluator) evalExpr(expr types.Expr, env value.Environment) (value.Value, error) {
	switch expr := expr.(type) {
	// Literal expressions - direct value representations
	case *types.IntExpr:
		// Integer literals: 42, -10, 0
		return value.Int(expr.Value), nil

	case *types.FloatExpr:
		// Floating-point literals: 3.14, -0.5, 1.0
		return value.Float(expr.Value), nil

	case *types.StringExpr:
		// String literals: "hello", "world"
		return value.String(expr.Value), nil

	case *types.BoolExpr:
		// Boolean literals: true, false
		return value.Bool(expr.Value), nil

	case *types.NullExpr:
		// Null literal: null
		return value.Null{}, nil

	case *types.PathExpr:
		// Path literals: ./file.txt, /absolute/path
		// Resolve relative paths against the evaluator's base directory
		path := e.resolvePath(expr.Value)

		return value.Path(path), nil

	case *types.IdentExpr:
		// Variable references: look up in current environment
		return e.evalIdent(expr.Name, env)

	// Compound data structure expressions
	case *types.ListExpr:
		// List literals: [1, 2, 3] - evaluate each element
		return e.evalList(expr, env)

	case *types.AttrSetExpr:
		// Attribute sets: { x = 1; y = 2; } - handle recursive sets
		return e.evalAttrSet(expr, env)

	// Operator expressions - binary and unary operations
	case *types.BinaryExpr:
		// Binary operators: +, -, *, ==, &&, ++, etc.
		return e.evalBinary(expr, env)

	case *types.UnaryExpr:
		// Unary operators: -, ! (negation and logical NOT)
		return e.evalUnary(expr, env)

	// Control flow expressions - conditional evaluation and scoping
	case *types.IfExpr:
		// Conditional expressions: if condition then value else alternative
		return e.evalIf(expr, env)

	case *types.LetExpr:
		// Let bindings: let x = 1; y = 2; in x + y
		return e.evalLet(expr, env)

	case *types.WithExpr:
		// With expressions: with attrs; expression
		return e.evalWith(expr, env)

	case *types.AssertExpr:
		// Assert expressions: assert condition; expression
		return e.evalAssert(expr, env)

	// Function expressions - creation and application
	case *types.FunctionExpr:
		// Function definitions: x: x + 1 (create closure capturing environment)
		return value.NewFunction(expr.Param, expr.Body, env), nil

	case *types.ApplyExpr:
		// Function application: f x (apply function to argument)
		return e.evalApply(expr, env)

	// Attribute access expressions - working with attribute sets
	case *types.SelectExpr:
		// Attribute selection: obj.attr.subattr
		return e.evalSelect(expr, env)

	case *types.HasAttrExpr:
		// Attribute existence test: obj ? attr
		return e.evalHasAttr(expr, env)

	default:
		// Unrecognized AST node type - this indicates a bug in the parser
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// evalIdent resolves variable references by looking up identifiers in the environment.
// This implements lexical scoping - variables are resolved in the environment where
// they are referenced, following the scope chain established by let bindings,
// function parameters, and with expressions.
func (e *Evaluator) evalIdent(name string, env value.Environment) (value.Value, error) {
	// Attempt to resolve the variable in the current environment
	if val, ok := env.Get(name); ok {
		// Variable found - return its value
		return val, nil
	}

	// Variable not found in any accessible scope
	return nil, fmt.Errorf("undefined variable: %s", name)
}

// evalList evaluates list literals by recursively evaluating each element.
// Lists in Nix are heterogeneous sequences that can contain any combination
// of value types. Evaluation is eager - all elements are evaluated immediately
// when the list expression is encountered.
func (e *Evaluator) evalList(expr *types.ListExpr, env value.Environment) (value.Value, error) {
	// Pre-allocate result slice with known size for efficiency
	elements := make([]value.Value, len(expr.Elements))
	// Evaluate each element expression in the current environment
	for i, elem := range expr.Elements {
		val, err := e.evalExpr(elem, env)
		if err != nil {
			// Propagate evaluation error with element context
			return nil, err
		}
		elements[i] = val
	}

	// Create and return the completed list value
	return value.NewList(elements...), nil
}

// evalAttrSet evaluates attribute set expressions with support for recursion.
// Attribute sets are key-value mappings that can be either:
// - Regular sets: { x = 1; y = 2; } - bindings cannot reference each other
// - Recursive sets: rec { x = 1; y = x + 1; } - bindings can reference each other
//
// For recursive sets, we implement a two-pass evaluation to handle dependencies:
// 1. First pass: evaluate simple expressions that don't reference other bindings
// 2. Second pass: evaluate complex expressions that may reference first-pass results.
func (e *Evaluator) evalAttrSet(
	expr *types.AttrSetExpr,
	env value.Environment,
) (value.Value, error) {
	// Initialize the result attribute set
	attrs := value.NewAttrs()

	// Determine evaluation environment based on recursion
	evalEnv := env
	if expr.Recursive {
		// Recursive attribute set: create extended environment for self-references
		recEnv := env.Extend()
		evalEnv = recEnv

		// Two-pass evaluation for recursive sets to handle dependencies
		// First pass: evaluate simple expressions (literals, non-referencing expressions)
		// This establishes basic bindings that other expressions can reference
		for _, binding := range expr.Bindings {
			if len(binding.Path) == 1 && isSimpleExpr(binding.Value) {
				// Simple expression - safe to evaluate first
				val, err := e.evalExpr(binding.Value, recEnv)
				if err != nil {
					return nil, err
				}
				// Add to both the result set and the environment
				attrs.Set(binding.Path[0], val)
				recEnv.Set(binding.Path[0], val)
			}
		}

		// Second pass: evaluate complex expressions that may reference first-pass results
		for _, binding := range expr.Bindings {
			if len(binding.Path) == 1 && !isSimpleExpr(binding.Value) {
				// Complex expression - evaluate with access to first-pass bindings
				val, err := e.evalExpr(binding.Value, recEnv)
				if err != nil {
					return nil, err
				}
				attrs.Set(binding.Path[0], val)
			} else if len(binding.Path) > 1 {
				// Nested attribute path: a.b.c = value
				if err := e.setNestedAttr(attrs, binding.Path, binding.Value, recEnv); err != nil {
					return nil, err
				}
			}
		}
	} else {
		// Non-recursive attribute set: straightforward evaluation
		// Bindings cannot reference each other, so order doesn't matter
		for _, binding := range expr.Bindings {
			if len(binding.Path) == 1 {
				// Simple attribute: name = value
				val, err := e.evalExpr(binding.Value, evalEnv)
				if err != nil {
					return nil, err
				}
				attrs.Set(binding.Path[0], val)
			} else {
				if err := e.setNestedAttr(attrs, binding.Path, binding.Value, evalEnv); err != nil {
					return nil, err
				}
			}
		}
	}

	return attrs, nil
}

// setNestedAttr handles nested attribute assignments like a.b.c = value.
// This method navigates through the attribute path, creating intermediate
// attribute sets as needed, and sets the final value at the end of the path.
// It handles conflicts where a path component already exists as a non-set value.
func (e *Evaluator) setNestedAttr(
	attrs *value.Attrs,
	path []string,
	expr types.Expr,
	env value.Environment,
) error {
	if len(path) == 0 {
		// Invalid: empty path should not occur in well-formed ASTs
		return errors.New("empty attribute path")
	}

	// Evaluate the value expression that will be assigned
	val, err := e.evalExpr(expr, env)
	if err != nil {
		return err
	}

	// Navigate through the path, creating intermediate attribute sets as needed
	current := attrs
	for i := range len(path) - 1 {
		key := path[i]
		if existing, ok := current.Get(key); ok {
			// Path component already exists - must be an attribute set
			if nested, ok := existing.(*value.Attrs); ok {
				// Continue navigation through existing attribute set
				current = nested
			} else {
				// Conflict: trying to treat non-set value as attribute set
				return fmt.Errorf("attribute path conflict at %s", key)
			}
		} else {
			// Path component doesn't exist - create new attribute set
			nested := value.NewAttrs()
			current.Set(key, nested)
			current = nested
		}
	}

	// Set the final value at the end of the path
	current.Set(path[len(path)-1], val)

	return nil
}

// isSimpleExpr determines if an expression is "simple" for recursive set evaluation.
// Simple expressions are those that don't reference other bindings within the same
// attribute set. These can be safely evaluated in the first pass of recursive set
// evaluation, providing values that more complex expressions can then reference.
//
// Simple expressions include all literal values that evaluate to themselves.
func isSimpleExpr(expr types.Expr) bool {
	switch expr.(type) {
	// All literal expressions are simple (no variable references)
	case *types.IntExpr, *types.FloatExpr, *types.StringExpr,
		*types.BoolExpr, *types.NullExpr, *types.PathExpr:
		return true
	default:
		// All other expressions may contain variable references
		// This includes: identifiers, functions, operations, control flow
		return false
	}
}

// resolvePath resolves path literals against the evaluator's base directory.
// This ensures that relative paths in Nix expressions are interpreted relative
// to a consistent base directory, typically where the Nix file is located.
// Absolute paths are returned unchanged, while relative paths are joined with
// the base directory to create absolute paths.
func (e *Evaluator) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		// Absolute path - use as-is
		return path
	}

	// Relative path - resolve against base directory
	return filepath.Join(e.baseDir, path)
}
