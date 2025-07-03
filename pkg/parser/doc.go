// Package parser implements a recursive descent parser with Pratt parsing for the Nix expression language.
//
// The parser is the second stage of the Nix interpreter pipeline, transforming a stream
// of tokens from the lexer into a well-formed Abstract Syntax Tree (AST) that can be
// evaluated by the evaluator.
//
// Architecture:
//
// The parser uses a combination of recursive descent and Pratt parsing techniques:
//   - Recursive descent for control structures and complex expressions
//   - Pratt parsing for operators with proper precedence and associativity
//   - Lookahead parsing for disambiguation of syntax elements
//
// Language Support:
//
// The parser supports the complete Nix expression language:
//
// Literals:
//   - Integers: 42, -10, 0
//   - Floats: 3.14, -0.5, 1.0
//   - Strings: "hello", "world with \"quotes\""
//   - Booleans: true, false
//   - Null: null
//   - Paths: ./file.txt, /absolute/path
//
// Operators (with precedence):
//  1. -> (implication - lowest precedence)
//  2. || (logical or)
//  3. && (logical and)
//  4. == != (equality comparison)
//  5. < > <= >= (relational comparison)
//  6. // (attribute set update)
//  7. ++ (list/string concatenation)
//  8. + - (addition/subtraction)
//  9. * / (multiplication/division)
//  10. function application (left-associative)
//  11. . (attribute selection - highest precedence)
//
// Control Flow:
//   - Conditionals: if condition then value else alternative
//   - Let bindings: let x = 1; y = 2; in x + y
//   - With expressions: with attrs; expression
//   - Assertions: assert condition; expression
//
// Functions:
//   - Definitions: x: x + 1
//   - Applications: f x (left-associative)
//   - Currying: f x y is parsed as (f x) y
//
// Data Structures:
//   - Lists: [1, 2, 3] or [1 2 3]
//   - Attribute sets: { x = 1; y = 2; }
//   - Recursive sets: rec { x = 1; y = x + 1; }
//   - Nested attributes: { a.b.c = value; }
//
// Attribute Operations:
//   - Selection: attrs.x.y
//   - Existence test: attrs ? x
//   - Default values: attrs.x or defaultValue
//
// Error Handling:
//
// The parser provides comprehensive error reporting:
//   - Syntax error detection with line/column information
//   - Expected token reporting for missing elements
//   - Multiple error collection for better user experience
//   - Structured error types for programmatic handling
//
// Performance Features:
//   - Single-pass parsing with minimal backtracking
//   - Efficient operator precedence resolution
//   - Memory-efficient AST node construction
//   - Early error detection and reporting
//
// Design Principles:
//   - Fail fast: detect errors as early as possible
//   - Informative errors: provide context for debugging
//   - Extensible: easy to add new language constructs
//   - Maintainable: clear separation of parsing concerns
//
// Usage Example:
//
//	lexer := lexer.New(`let x = 42; in if x > 0 then "positive" else "negative"`)
//	parser := parser.New(lexer)
//	ast, err := parser.Parse()
//	if err != nil {
//	    fmt.Printf("Parse error: %v\n", err)
//	    return
//	}
//	// ast now contains the parsed expression tree
package parser
