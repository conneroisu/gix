package eval

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/conneroisu/gix/internal/value"
	"github.com/conneroisu/gix/pkg/derivation"
)

// registerBuiltins populates the evaluator with all standard Nix built-in functions.
// This creates the standard library that's available in all Nix expressions,
// including type checking, data manipulation, and system functions.
//
// The built-ins are organized into categories:
// - Constants: true, false, null
// - Type checking: isNull, isBool, isInt, isFloat, isString, isList, isAttrs, isFunction
// - Conversion: toString
// - List operations: length, head, tail, elem
// - Attribute operations: attrNames, attrValues, hasAttr, getAttr
// - Math functions: add, sub, mul, div
// - System functions: derivation.
func (e *Evaluator) registerBuiltins() {
	// Built-in constants - fundamental values available in all expressions
	e.builtins["true"] = value.Bool(true)   // Boolean true constant
	e.builtins["false"] = value.Bool(false) // Boolean false constant
	e.builtins["null"] = value.Null{}       // Null constant

	// Type checking functions - runtime type inspection
	// Nix examples: isNull null → true, isBool "hello" → false
	e.registerBuiltin("isNull", 1, builtinIsNull)         // isNull value → bool
	e.registerBuiltin("isBool", 1, builtinIsBool)         // isBool value → bool
	e.registerBuiltin("isInt", 1, builtinIsInt)           // isInt value → bool
	e.registerBuiltin("isFloat", 1, builtinIsFloat)       // isFloat value → bool
	e.registerBuiltin("isString", 1, builtinIsString)     // isString value → bool
	e.registerBuiltin("isList", 1, builtinIsList)         // isList value → bool
	e.registerBuiltin("isAttrs", 1, builtinIsAttrs)       // isAttrs value → bool
	e.registerBuiltin("isFunction", 1, builtinIsFunction) // isFunction value → bool

	// Conversion functions - type transformations
	// Nix example: toString 42 → "42"
	e.registerBuiltin("toString", 1, builtinToString) // toString value → string

	// List operations - working with sequences
	// Nix examples: length [1 2 3] → 3, head [1 2 3] → 1
	e.registerBuiltin("length", 1, builtinLength) // length list|string|attrs → int
	e.registerBuiltin("head", 1, builtinHead)     // head list → value
	e.registerBuiltin("tail", 1, builtinTail)     // tail list → list
	e.registerBuiltin("elem", 2, builtinElem)     // elem value list → bool

	// Attribute set operations - working with key-value mappings
	// Nix examples: attrNames {x=1; y=2;} → ["x" "y"]
	e.registerBuiltin("attrNames", 1, builtinAttrNames)   // attrNames attrs → list
	e.registerBuiltin("attrValues", 1, builtinAttrValues) // attrValues attrs → list
	e.registerBuiltin("hasAttr", 2, builtinHasAttr)       // hasAttr name attrs → bool
	e.registerBuiltin("getAttr", 2, builtinGetAttr)       // getAttr name attrs → value

	// Mathematical functions - arithmetic operations
	// Nix examples: add 1 2 → 3, mul 3 4 → 12
	e.registerBuiltin("add", 2, builtinAdd) // add a b → number
	e.registerBuiltin("sub", 2, builtinSub) // sub a b → number
	e.registerBuiltin("mul", 2, builtinMul) // mul a b → number
	e.registerBuiltin("div", 2, builtinDiv) // div a b → number

	// System functions - Nix-specific operations
	// Nix example: derivation {name="hello"; builder="/bin/sh"; args=["-c" "echo hello"];}
	e.registerBuiltin("derivation", 1, builtinDerivation) // derivation attrs → attrs
}

// registerBuiltin wraps a built-in function implementation with arity checking.
// This ensures that built-in functions receive the correct number of arguments,
// providing clear error messages when called incorrectly.
//
// Parameters:
// - name: The function name as it appears in Nix expressions
// - arity: Expected number of arguments (e.g., 1 for unary, 2 for binary)
// - fn: The Go implementation function
//
// Example usage in Go:
//
//	e.registerBuiltin("add", 2, builtinAdd)  // Registers add as 2-argument function
func (e *Evaluator) registerBuiltin(
	name string,
	arity int,
	fn func([]value.Value) (value.Value, error),
) {
	// Create wrapper that validates argument count before calling implementation
	wrapped := func(args []value.Value) (value.Value, error) {
		if len(args) != arity {
			// Provide clear error message for incorrect argument count
			return nil, fmt.Errorf("%s expects %d argument(s), got %d", name, arity, len(args))
		}

		// Argument count is correct - delegate to actual implementation
		return fn(args)
	}
	// Register the wrapped function in the built-ins registry
	e.builtins[name] = value.NewBuiltin(name, wrapped)
}

// =============================================================================
// Type Checking Built-ins
// =============================================================================
// These functions allow runtime type inspection of Nix values.
// All return boolean values indicating whether the argument matches the type.

// builtinIsNull checks if a value is null.
// Nix usage: isNull null → true, isNull 42 → false
// Go implementation: checks if value implements value.Null interface.
func builtinIsNull(args []value.Value) (value.Value, error) {
	// Use Go type assertion to check if the value is null
	_, isNull := args[0].(value.Null)

	return value.Bool(isNull), nil
}

// builtinIsBool checks if a value is a boolean.
// Nix usage: isBool true → true, isBool "hello" → false
// Go implementation: checks if value implements value.Bool interface.
func builtinIsBool(args []value.Value) (value.Value, error) {
	// Use Go type assertion to check if the value is a boolean
	_, isBool := args[0].(value.Bool)

	return value.Bool(isBool), nil
}

// builtinIsInt checks if a value is an integer.
// Nix usage: isInt 42 → true, isInt 3.14 → false
// Go implementation: checks if value implements value.Int interface.
func builtinIsInt(args []value.Value) (value.Value, error) {
	// Use Go type assertion to check if the value is an integer
	_, isInt := args[0].(value.Int)

	return value.Bool(isInt), nil
}

// builtinIsFloat checks if a value is a floating-point number.
// Nix usage: isFloat 3.14 → true, isFloat 42 → false
// Go implementation: checks if value implements value.Float interface.
func builtinIsFloat(args []value.Value) (value.Value, error) {
	// Use Go type assertion to check if the value is a float
	_, isFloat := args[0].(value.Float)

	return value.Bool(isFloat), nil
}

// builtinIsString checks if a value is a string.
// Nix usage: isString "hello" → true, isString 42 → false
// Go implementation: checks if value implements value.String interface.
func builtinIsString(args []value.Value) (value.Value, error) {
	// Use Go type assertion to check if the value is a string
	_, isString := args[0].(value.String)

	return value.Bool(isString), nil
}

// builtinIsList checks if a value is a list.
// Nix usage: isList [1 2 3] → true, isList {x=1;} → false
// Go implementation: checks if value is a pointer to value.List.
func builtinIsList(args []value.Value) (value.Value, error) {
	// Use Go type assertion to check if the value is a list
	_, isList := args[0].(*value.List)

	return value.Bool(isList), nil
}

// builtinIsAttrs checks if a value is an attribute set.
// Nix usage: isAttrs {x=1; y=2;} → true, isAttrs [1 2 3] → false
// Go implementation: checks if value is a pointer to value.Attrs.
func builtinIsAttrs(args []value.Value) (value.Value, error) {
	// Use Go type assertion to check if the value is an attribute set
	_, isAttrs := args[0].(*value.Attrs)

	return value.Bool(isAttrs), nil
}

// builtinIsFunction checks if a value is a function (user-defined or built-in).
// Nix usage: isFunction (x: x + 1) → true, isFunction length → true, isFunction 42 → false
// Go implementation: checks if value is either *value.Function or *value.Builtin.
func builtinIsFunction(args []value.Value) (value.Value, error) {
	// Check if value is either a user function or built-in function
	switch args[0].(type) {
	case *value.Function, *value.Builtin:
		return value.Bool(true), nil
	default:
		return value.Bool(false), nil
	}
}

// =============================================================================
// Conversion Built-ins
// =============================================================================
// These functions convert values between different types.

// builtinToString converts various value types to their string representations.
// Nix usage: toString 42 → "42", toString true → "true", toString 3.14 → "3.14"
// Go implementation: uses type switch and Go's standard conversion functions.
func builtinToString(args []value.Value) (value.Value, error) {
	// Convert value to string based on its type
	switch v := args[0].(type) {
	case value.String:
		// Already a string - return as-is
		return v, nil
	case value.Int:
		// Convert integer to decimal string representation
		return value.String(strconv.FormatInt(int64(v), 10)), nil
	case value.Float:
		// Convert float to string with automatic precision
		return value.String(strconv.FormatFloat(float64(v), 'f', -1, 64)), nil
	case value.Bool:
		// Convert boolean to "true" or "false"
		if v {
			return value.String("true"), nil
		}

		return value.String("false"), nil
	case value.Null:
		return value.String("null"), nil
	case value.Path:
		return value.String(v), nil
	default:
		return nil, fmt.Errorf("cannot convert %v to string", v.Type())
	}
}

// List operations.
func builtinLength(args []value.Value) (value.Value, error) {
	switch v := args[0].(type) {
	case *value.List:
		return value.Int(v.Len()), nil
	case value.String:
		return value.Int(len(v)), nil
	case *value.Attrs:
		return value.Int(v.Len()), nil
	default:
		return nil, fmt.Errorf("length expects a list, string, or attrset, got %v", v.Type())
	}
}

func builtinHead(args []value.Value) (value.Value, error) {
	list, ok := args[0].(*value.List)
	if !ok {
		return nil, fmt.Errorf("head expects a list, got %v", args[0].Type())
	}

	if list.Len() == 0 {
		return nil, errors.New("head called on empty list")
	}

	return list.Get(0), nil
}

func builtinTail(args []value.Value) (value.Value, error) {
	list, ok := args[0].(*value.List)
	if !ok {
		return nil, fmt.Errorf("tail expects a list, got %v", args[0].Type())
	}

	if list.Len() == 0 {
		return nil, errors.New("tail called on empty list")
	}

	elements := list.Elements()

	return value.NewList(elements[1:]...), nil
}

func builtinElem(args []value.Value) (value.Value, error) {
	elem := args[0]
	list, ok := args[1].(*value.List)
	if !ok {
		return nil, fmt.Errorf("elem expects a list as second argument, got %v", args[1].Type())
	}

	for _, e := range list.Elements() {
		if elem.Equals(e) {
			return value.Bool(true), nil
		}
	}

	return value.Bool(false), nil
}

// Attribute set operations.
func builtinAttrNames(args []value.Value) (value.Value, error) {
	attrs, ok := args[0].(*value.Attrs)
	if !ok {
		return nil, fmt.Errorf("attrNames expects an attribute set, got %v", args[0].Type())
	}

	keys := attrs.Keys()
	names := make([]value.Value, len(keys))
	for i, k := range keys {
		names[i] = value.String(k)
	}

	return value.NewList(names...), nil
}

func builtinAttrValues(args []value.Value) (value.Value, error) {
	attrs, ok := args[0].(*value.Attrs)
	if !ok {
		return nil, fmt.Errorf("attrValues expects an attribute set, got %v", args[0].Type())
	}

	keys := attrs.Keys()
	values := make([]value.Value, len(keys))
	for i, k := range keys {
		v, _ := attrs.Get(k)
		values[i] = v
	}

	return value.NewList(values...), nil
}

func builtinHasAttr(args []value.Value) (value.Value, error) {
	name, ok := args[0].(value.String)
	if !ok {
		return nil, fmt.Errorf("hasAttr expects a string as first argument, got %v", args[0].Type())
	}

	attrs, ok := args[1].(*value.Attrs)
	if !ok {
		return nil, fmt.Errorf(
			"hasAttr expects an attribute set as second argument, got %v",
			args[1].Type(),
		)
	}

	_, exists := attrs.Get(string(name))

	return value.Bool(exists), nil
}

func builtinGetAttr(args []value.Value) (value.Value, error) {
	name, ok := args[0].(value.String)
	if !ok {
		return nil, fmt.Errorf("getAttr expects a string as first argument, got %v", args[0].Type())
	}

	attrs, ok := args[1].(*value.Attrs)
	if !ok {
		return nil, fmt.Errorf(
			"getAttr expects an attribute set as second argument, got %v",
			args[1].Type(),
		)
	}

	val, exists := attrs.Get(string(name))
	if !exists {
		return nil, fmt.Errorf("attribute '%s' not found", name)
	}

	return val, nil
}

// Math functions.
func builtinAdd(args []value.Value) (value.Value, error) {
	return evalAdd(args[0], args[1])
}

func builtinSub(args []value.Value) (value.Value, error) {
	return evalSub(args[0], args[1])
}

func builtinMul(args []value.Value) (value.Value, error) {
	return evalMul(args[0], args[1])
}

func builtinDiv(args []value.Value) (value.Value, error) {
	return evalDiv(args[0], args[1])
}

// Derivation functions.
func builtinDerivation(args []value.Value) (value.Value, error) {
	attrs, ok := args[0].(*value.Attrs)
	if !ok {
		return nil, fmt.Errorf("derivation expects an attribute set, got %v", args[0].Type())
	}

	// Create derivation
	drv, err := derivation.FromAttrs(attrs)
	if err != nil {
		return nil, fmt.Errorf("invalid derivation: %v", err)
	}

	// Return as attribute set
	return drv.ToAttrs(), nil
}
