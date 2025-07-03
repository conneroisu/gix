package lexer

import (
	"fmt"
)

// TokenType represents the classification of lexical tokens in the Nix language.
// Each token type corresponds to a specific syntactic element that the parser
// can recognize and process. The types are grouped logically for maintainability.
type TokenType int

// Token type constants define all possible lexical elements in Nix expressions.
// The iota enumeration ensures unique integer values for each token type.
// Grouping helps organize related tokens and makes the grammar easier to understand.
const (
	// Special tokens for lexical analysis control.
	TOKEN_EOF     = iota // End of file marker
	TOKEN_ILLEGAL        // Invalid/unrecognized character sequences

	// Literal value tokens - represent data directly in source code.
	TOKEN_INT    // Integer literals (42, -10, 0)
	TOKEN_FLOAT  // Floating-point literals (3.14, -0.5, 1.0)
	TOKEN_STRING // String literals ("hello", "world")
	TOKEN_PATH   // Path literals (./file, /absolute/path)
	TOKEN_IDENT  // Identifiers and variable names

	// Reserved keywords - language control structures and built-in concepts.
	TOKEN_IF      // "if" conditional expression start
	TOKEN_THEN    // "then" conditional true branch
	TOKEN_ELSE    // "else" conditional false branch
	TOKEN_LET     // "let" variable binding start
	TOKEN_IN      // "in" let expression body start
	TOKEN_WITH    // "with" scope extension
	TOKEN_ASSERT  // "assert" assertion statement
	TOKEN_OR      // "or" alternative value operator
	TOKEN_AND     // "and" logical conjunction (as keyword)
	TOKEN_NOT     // "not" logical negation (as keyword)
	TOKEN_REC     // "rec" recursive attribute set modifier
	TOKEN_INHERIT // "inherit" attribute inheritance

	// Operators - symbols that perform operations on values
	// Assignment and binding.
	TOKEN_ASSIGN // "=" assignment operator

	// Arithmetic operators.
	TOKEN_PLUS     // "+" addition
	TOKEN_MINUS    // "-" subtraction/negation
	TOKEN_MULTIPLY // "*" multiplication
	TOKEN_DIVIDE   // "/" division

	// Comparison operators.
	TOKEN_EQ  // "==" equality
	TOKEN_NEQ // "!=" inequality
	TOKEN_LT  // "<" less than
	TOKEN_GT  // ">" greater than
	TOKEN_LTE // "<=" less than or equal
	TOKEN_GTE // ">=" greater than or equal

	// Logical operators.
	TOKEN_AND_OP // "&&" logical AND
	TOKEN_OR_OP  // "||" logical OR
	TOKEN_IMPL   // "->" logical implication

	// Specialized operators.
	TOKEN_CONCAT   // "++" list/string concatenation
	TOKEN_QUESTION // "?" attribute existence test
	TOKEN_DOT      // "." attribute access

	// Delimiters - structural punctuation for grouping and separation.
	TOKEN_SEMICOLON // ";" statement separator
	TOKEN_COLON     // ":" function parameter separator
	TOKEN_COMMA     // "," element separator

	// Grouping delimiters.
	TOKEN_LPAREN   // "(" left parenthesis
	TOKEN_RPAREN   // ")" right parenthesis
	TOKEN_LBRACE   // "{" left brace (attribute sets)
	TOKEN_RBRACE   // "}" right brace
	TOKEN_LBRACKET // "[" left bracket (lists)
	TOKEN_RBRACKET // "]" right bracket
)

// Token represents a complete lexical unit from the Nix source code.
// Each token contains its classification, the actual text from the source,
// and position information for accurate error reporting and debugging.
type Token struct {
	Type    TokenType // The classification of this token (what kind it is)
	Literal string    // The actual text from source ("42", "hello", "+", etc.)
	Line    int       // Line number in source (1-based for human readability)
	Column  int       // Column position in line (0-based within line)
}

// tokenNames provides human-readable string representations for each token type.
// Used primarily for debugging, error messages, and development tools.
// Each token type maps to a descriptive name that clearly identifies its purpose.
var tokenNames = map[TokenType]string{
	TOKEN_EOF:       "EOF",
	TOKEN_ILLEGAL:   "ILLEGAL",
	TOKEN_INT:       "INT",
	TOKEN_FLOAT:     "FLOAT",
	TOKEN_STRING:    "STRING",
	TOKEN_PATH:      "PATH",
	TOKEN_IDENT:     "IDENT",
	TOKEN_IF:        "IF",
	TOKEN_THEN:      "THEN",
	TOKEN_ELSE:      "ELSE",
	TOKEN_LET:       "LET",
	TOKEN_IN:        "IN",
	TOKEN_WITH:      "WITH",
	TOKEN_ASSERT:    "ASSERT",
	TOKEN_OR:        "OR",
	TOKEN_AND:       "AND",
	TOKEN_NOT:       "NOT",
	TOKEN_REC:       "REC",
	TOKEN_INHERIT:   "INHERIT",
	TOKEN_ASSIGN:    "ASSIGN",
	TOKEN_PLUS:      "PLUS",
	TOKEN_MINUS:     "MINUS",
	TOKEN_MULTIPLY:  "MULTIPLY",
	TOKEN_DIVIDE:    "DIVIDE",
	TOKEN_EQ:        "EQ",
	TOKEN_NEQ:       "NEQ",
	TOKEN_LT:        "LT",
	TOKEN_GT:        "GT",
	TOKEN_LTE:       "LTE",
	TOKEN_GTE:       "GTE",
	TOKEN_AND_OP:    "AND_OP",
	TOKEN_OR_OP:     "OR_OP",
	TOKEN_IMPL:      "IMPL",
	TOKEN_CONCAT:    "CONCAT",
	TOKEN_QUESTION:  "QUESTION",
	TOKEN_DOT:       "DOT",
	TOKEN_SEMICOLON: "SEMICOLON",
	TOKEN_COLON:     "COLON",
	TOKEN_COMMA:     "COMMA",
	TOKEN_LPAREN:    "LPAREN",
	TOKEN_RPAREN:    "RPAREN",
	TOKEN_LBRACE:    "LBRACE",
	TOKEN_RBRACE:    "RBRACE",
	TOKEN_LBRACKET:  "LBRACKET",
	TOKEN_RBRACKET:  "RBRACKET",
}

// String returns a human-readable string representation of the token type.
// This implements the Stringer interface, making token types easy to print
// and debug. For unknown token types, returns a numeric representation.
func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	// Fallback for any token types not in the map
	return fmt.Sprintf("TokenType(%d)", int(t))
}

// keywords maps reserved word strings to their corresponding token types.
// This lookup table enables quick keyword recognition during tokenization.
// Any identifier matching these strings becomes a keyword token instead of
// a regular identifier, preventing their use as variable names.
var keywords = map[string]TokenType{
	"if":      TOKEN_IF,
	"then":    TOKEN_THEN,
	"else":    TOKEN_ELSE,
	"let":     TOKEN_LET,
	"in":      TOKEN_IN,
	"with":    TOKEN_WITH,
	"assert":  TOKEN_ASSERT,
	"or":      TOKEN_OR,
	"and":     TOKEN_AND,
	"not":     TOKEN_NOT,
	"rec":     TOKEN_REC,
	"inherit": TOKEN_INHERIT,
}

// LookupIdent determines whether a given identifier string is a reserved keyword.
// This function is called during tokenization to distinguish between:
// - Reserved keywords (if, let, then, etc.) -> specific keyword token
// - User identifiers (variable names, function names) -> TOKEN_IDENT
//
// This separation is crucial for proper parsing and semantic analysis.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		// This identifier is a reserved keyword
		return tok
	}
	// Regular user-defined identifier
	return TOKEN_IDENT
}

// isLetter determines if a character can start or continue an identifier.
// In Nix, identifiers can contain:
// - Lowercase letters (a-z)
// - Uppercase letters (A-Z)
// - Underscores (_)
// This follows standard identifier conventions used by most programming languages.
func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

// isDigit determines if a character is a decimal digit (0-9).
// Used for:
// - Numeric literal recognition
// - Continuation of identifiers (after first character)
// - Decimal point detection in float literals.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// isPathChar determines if a character is valid within a Nix path literal.
// Path literals can contain:
// - Letters and digits (standard identifier characters)
// - Forward slashes (/) for directory separation
// - Dots (.) for current/parent directory and file extensions
// - Hyphens (-) for kebab-case naming
// - Underscores (_) for snake_case naming
//
// This allows paths like: ./file.txt, /usr/bin/program, ../parent-dir.
func isPathChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '/' || ch == '.' || ch == '-' || ch == '_'
}
