// Package eval provides the expression evaluator for the Nix expression language interpreter.
//
// The evaluator is the final stage of the Nix interpreter pipeline, taking Abstract
// Syntax Trees (ASTs) from the parser and computing their runtime values. It implements
// the complete Nix evaluation semantics including lazy evaluation, lexical scoping,
// and built-in functions.
//
// Architecture:
//
// The evaluator uses a tree-walking approach with the following key components:
//   - Evaluator: Main evaluation engine with environment management
//   - Environment: Lexical scoping and variable binding system
//   - Value System: Runtime representation of all Nix values
//   - Built-in Functions: Standard library implementations
//
// The design follows domain-driven principles with clear separation of concerns:
//   - evaluator.go: Core evaluation logic and AST traversal
//   - operators.go: Binary and unary operator implementations
//   - control_flow.go: Control flow constructs (if, let, with, assert)
//   - functions.go: Function application and closure handling
//   - builtins.go: Built-in function library
//
// Evaluation Strategy:
//
// The evaluator implements eager evaluation with lazy semantics where appropriate:
//   - Function arguments are evaluated when passed (eager)
//   - Let bindings are evaluated when accessed (lazy-ish)
//   - Attribute sets support recursive references
//   - Short-circuit evaluation for logical operators
//
// Supported Language Features:
//
// All major Nix language constructs are supported:
//   - Literals: integers, floats, strings, booleans, null, paths
//   - Operators: arithmetic, comparison, logical, concatenation
//   - Control flow: if-then-else, let-in, with, assert
//   - Functions: definitions, applications, closures
//   - Data structures: lists, attribute sets (recursive and non-recursive)
//   - Built-ins: comprehensive standard library
//   - Derivations: Nix store integration
//
// Built-in Functions:
//
// The evaluator provides 25+ built-in functions including:
//   - Type checking: isNull, isBool, isInt, isFloat, isString, isList, isAttrs, isFunction
//   - Conversions: toString
//   - List operations: length, head, tail, elem
//   - Attribute operations: attrNames, attrValues, hasAttr, getAttr
//   - Math: add, sub, mul, div
//   - System: derivation
//
// Error Handling:
//
// Comprehensive error reporting with:
//   - Type errors with expected vs actual types
//   - Undefined variable errors
//   - Runtime errors (division by zero, etc.)
//   - Function arity mismatches
//   - Attribute access errors
//
// Performance Features:
//   - Tree-walking evaluation for simplicity and correctness
//   - Immutable values for safety
//   - Structural sharing for memory efficiency
//   - Short-circuit evaluation for logical operators
//
// Usage Example:
//
//	lexer := lexer.New(`let x = 42; f = y: x + y; in f 8`)
//	parser := parser.New(lexer)
//	ast, err := parser.Parse()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	evaluator := eval.New(".")
//	result, err := evaluator.Eval(ast)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println(result.String()) // Output: 50
package eval
