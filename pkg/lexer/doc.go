// Package lexer provides lexical analysis functionality for the Nix expression language.
//
// The lexer is the first stage of the Nix interpreter pipeline, responsible for
// converting raw source text into a stream of tokens that can be consumed by the parser.
//
// Key Features:
//
// Token Recognition:
//   - Keywords: if, then, else, let, in, with, assert, or, and, not, rec, inherit
//   - Identifiers: variable names following Nix naming rules
//   - Literals: integers, floats, strings (with escape sequences), paths
//   - Operators: +, -, *, /, ==, !=, <, >, <=, >=, &&, ||, ->, ++, ?, .
//   - Delimiters: (, ), {, }, [, ], ;, :, ,, =
//
// Comment Handling:
//   - Single-line comments starting with '#'
//   - Multi-line comments enclosed in /* */
//   - Comments are skipped during tokenization
//
// Position Tracking:
//   - Accurate line and column information for each token
//   - Essential for meaningful error reporting
//   - Handles both Unix (\n) and Windows (\r\n) line endings
//
// String Processing:
//   - Double-quoted strings with escape sequences
//   - Proper handling of escaped quotes, newlines, etc.
//   - Unicode support through Go's UTF-8 handling
//
// Performance Optimizations:
//   - Single-pass tokenization
//   - Minimal token design for memory efficiency
//   - Lazy computation of token properties
//   - Efficient character-by-character scanning
//
// Error Handling:
//   - Graceful handling of unexpected characters
//   - ILLEGAL tokens for invalid input
//   - Position information preserved for error reporting
//
// The lexer follows the maximal munch principle, consuming the longest possible
// sequence of characters for each token. This ensures correct tokenization of
// multi-character operators like '++', '->', '&&', etc.
//
// Usage Example:
//
//	lexer := lexer.New("let x = 42; in x + 1")
//	for {
//	    token := lexer.NextToken()
//	    if token.Type == lexer.TOKEN_EOF {
//	        break
//	    }
//	    fmt.Printf("%s: %s\n", token.Type, token.Literal)
//	}
package lexer
