// Package types provides Abstract Syntax Tree (AST) node definitions for the Nix expression language.
//
// This package defines all the expression types that make up the Nix language's syntax tree.
// Each expression type implements the Expr interface and represents a specific language construct.
//
// The AST is designed to be:
//   - Immutable: Nodes don't change after creation
//   - Type-safe: Strong typing prevents many runtime errors
//   - Extensible: Easy to add new expression types
//   - Debuggable: String() methods provide readable representations
//   - Evaluable: Each node can be evaluated to produce a value
//
// Expression Categories:
//
// Literals:
//   - IntExpr: Integer values (42, -10)
//   - FloatExpr: Floating-point values (3.14, -0.5)
//   - StringExpr: String literals ("hello", "world")
//   - BoolExpr: Boolean values (true, false)
//   - NullExpr: The null value
//   - PathExpr: File system paths (/etc/hosts, ./file.txt)
//
// Identifiers and Variables:
//   - IdentExpr: Variable references (x, myVar)
//
// Operators:
//   - BinaryExpr: Binary operations (1 + 2, x && y, a ++ b)
//   - UnaryExpr: Unary operations (!x, -y)
//
// Control Flow:
//   - IfExpr: Conditional expressions (if cond then a else b)
//   - LetExpr: Local bindings (let x = 1; y = 2; in x + y)
//   - WithExpr: Scope introduction (with attrs; expr)
//   - AssertExpr: Assertions (assert condition; expr)
//
// Functions:
//   - FunctionExpr: Function definitions (x: x + 1)
//   - ApplyExpr: Function applications (f x)
//
// Data Structures:
//   - ListExpr: Lists ([1, 2, 3])
//   - AttrSetExpr: Attribute sets ({ x = 1; y = 2; })
//
// Attribute Operations:
//   - SelectExpr: Attribute selection (attrs.x.y)
//   - HasAttrExpr: Attribute existence test (attrs ? x)
//
// The parser builds these AST nodes from tokens, and the evaluator traverses them
// to compute values. Each node type is responsible for validating its own structure
// and providing meaningful string representations for debugging.
//
// All expression nodes implement the Expr interface, which provides:
//   - String() method for debugging and pretty-printing
//   - Type safety through Go's type system
//   - Consistent traversal patterns for evaluation
package types
