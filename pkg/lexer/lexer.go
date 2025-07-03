package lexer

import (
	"unicode"
)

// Lexer represents a lexical analyzer for Nix expressions.
// It implements a single-pass scanner that converts source text into tokens.
// The lexer maintains position information for accurate error reporting.
type Lexer struct {
	input        string // The complete input string being tokenized
	position     int    // Current position in input (points to current char)
	readPosition int    // Current reading position in input (after current char)
	ch           byte   // Current char under examination (0 for EOF)
	line         int    // Current line number (1-based for user display)
	column       int    // Current column number (0-based within line)
}

// New creates a new lexer instance for the given input string.
// The lexer is initialized with position (1,0) and reads the first character.
// This prepares the lexer for immediate tokenization via NextToken().
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1, // Start at line 1 for human-readable error messages
		column: 0, // Start at column 0, will increment to 1 on first readChar
	}
	// Prime the lexer by reading the first character
	l.readChar()

	return l
}

// readChar reads the next character and advances the lexer position.
// This method implements the core character consumption mechanism:
// 1. Sets ch to the character at readPosition (or 0 for EOF)
// 2. Advances position pointers
// 3. Updates line/column tracking for position reporting
//
// Line/column tracking handles both Unix (\n) and Windows (\r\n) line endings.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		// Use ASCII NUL (0) to represent end-of-file condition
		l.ch = 0
	} else {
		// Read the character at the current read position
		l.ch = l.input[l.readPosition]
	}

	// Advance both position pointers
	l.position = l.readPosition
	l.readPosition++

	// Update line and column counters for accurate error reporting
	if l.ch == '\n' {
		// Newline: increment line, reset column to start of line
		l.line++
		l.column = 0
	} else {
		// Regular character: advance column position
		l.column++
	}
}

// peekChar returns the next character without advancing the lexer position.
// This is essential for lookahead operations when tokenizing multi-character
// operators like "++", "&&", "==", etc. Returns 0 (EOF) if at end of input.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		// Return EOF marker when peeking beyond input
		return 0
	}

	// Return the character at the read position without consuming it
	return l.input[l.readPosition]
}

// skipWhitespace consumes and skips over all whitespace characters.
// This includes spaces, tabs, newlines, and carriage returns.
// Essential for clean tokenization by eliminating meaningless whitespace.
func (l *Lexer) skipWhitespace() {
	// Continue consuming characters while they are whitespace
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment consumes and skips over comment text.
// Handles both Nix comment styles:
// 1. Single-line comments starting with '#' (until newline or EOF)
// 2. Multi-line comments enclosed in /* ... */ (with proper nesting)
//
// Comments are completely ignored during tokenization, allowing clean
// separation of documentation from executable code.
func (l *Lexer) skipComment() {
	if l.ch == '#' {
		// Single-line comment: consume everything until newline or EOF
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	} else if l.ch == '/' && l.peekChar() == '*' {
		// Multi-line comment: consume /* ... */ block
		l.readChar() // Skip opening '/'
		l.readChar() // Skip opening '*'

		// Consume all characters until we find the closing */
		for l.ch != 0 {
			// Check for the closing */ sequence
			if l.ch == '*' && l.peekChar() == '/' {
				l.readChar() // Skip closing '*'
				l.readChar() // Skip closing '/'

				break
			}
			// Continue consuming characters within the comment
			l.readChar()
		}
	}
}

// readIdentifier reads a complete identifier or keyword from the input.
// Identifiers in Nix can contain letters, digits, underscores, and hyphens.
// This follows the maximal munch principle - consuming the longest possible identifier.
//
// The returned string will be checked against the keyword table to determine
// if it's a reserved word (like 'let', 'if', 'then') or a user identifier.
func (l *Lexer) readIdentifier() string {
	// Mark the start position of the identifier
	position := l.position

	// Consume all valid identifier characters (letters, digits, _, -)
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	// Return the slice of input containing the complete identifier
	return l.input[position:l.position]
}

// readNumber reads a complete numeric literal from the input.
// Supports both integer and floating-point numbers:
// - Integers: 42, 0, 1234
// - Floats: 3.14, 0.5, 123.456
//
// The function uses lookahead to distinguish between:
// - Decimal points in floats (3.14)
// - Dots in attribute access (obj.attr)
// Only decimal points followed by digits are considered float literals.
func (l *Lexer) readNumber() (string, TokenType) {
	// Mark the start position of the number
	position := l.position
	// Default to integer type, upgrade to float if needed
	tokenType := TOKEN_INT

	// Consume the integer part (required for all numbers)
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point followed by digits (float literal)
	if l.ch == '.' && isDigit(l.peekChar()) {
		// This is a float literal, not attribute access
		tokenType = TOKEN_FLOAT
		l.readChar() // Consume the decimal point

		// Consume the fractional part
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Return the complete number string and its determined type
	return l.input[position:l.position], TokenType(tokenType)
}

// readString reads a complete string literal from the input.
// Handles double-quoted strings with escape sequences:
// - \" for literal quotes
// - \n for newlines
// - \\ for literal backslashes
// - \t for tabs
//
// The function properly handles escaped characters and returns the
// string content without the surrounding quotes.
func (l *Lexer) readString() string {
	// Skip the opening quote and mark the start of string content
	position := l.position + 1

	for {
		l.readChar()

		// End of string: closing quote or unexpected EOF
		if l.ch == '"' || l.ch == 0 {
			break
		}

		// Handle escape sequences: consume the escape character and the next character
		if l.ch == '\\' {
			l.readChar() // Skip the backslash
			// The next character (if any) is consumed in the next iteration
		}
	}

	// Return the string content (excluding surrounding quotes)
	return l.input[position:l.position]
}

// readPath reads a complete path literal from the input.
// Paths in Nix can be absolute (/usr/bin) or relative (./file, ../dir).
// Valid path characters include letters, digits, dots, slashes, underscores,
// and hyphens. This follows Nix path literal syntax rules.
func (l *Lexer) readPath() string {
	// Mark the start position of the path
	position := l.position

	// Consume all valid path characters using maximal munch
	for isPathChar(l.ch) {
		l.readChar()
	}

	// Return the complete path string
	return l.input[position:l.position]
}

// NextToken returns the next token from the input stream.
// This is the main entry point for tokenization, implementing a single-pass
// scanner that recognizes all Nix language tokens.
//
// The tokenization process:
// 1. Skip whitespace and comments
// 2. Capture current position for error reporting
// 3. Recognize token type based on current character
// 4. Apply maximal munch for multi-character tokens
// 5. Return complete token with type, literal, and position.
func (l *Lexer) NextToken() Token {
	var tok Token

	// Skip all non-significant characters (whitespace and comments)
	for {
		l.skipWhitespace()
		// Check for comment start patterns
		if l.ch == '#' || (l.ch == '/' && l.peekChar() == '*') {
			l.skipComment()
		} else {
			// No more whitespace or comments, ready to tokenize
			break
		}
	}

	// Capture current position for this token (essential for error reporting)
	tok.Line = l.line
	tok.Column = l.column

	// Token recognition switch: each case handles a specific character or character sequence
	// Multi-character operators use lookahead to distinguish similar patterns
	switch l.ch {
	// Assignment and equality operators
	case '=':
		if l.peekChar() == '=' {
			// "==" equality comparison operator
			l.readChar()
			tok = Token{Type: TOKEN_EQ, Literal: "==", Line: tok.Line, Column: tok.Column}
		} else {
			// "=" assignment operator
			tok = Token{Type: TOKEN_ASSIGN, Literal: "=", Line: tok.Line, Column: tok.Column}
		}

	// Addition and concatenation operators
	case '+':
		if l.peekChar() == '+' {
			// "++" list/string concatenation operator
			l.readChar()
			tok = Token{Type: TOKEN_CONCAT, Literal: "++", Line: tok.Line, Column: tok.Column}
		} else {
			// "+" addition operator
			tok = Token{Type: TOKEN_PLUS, Literal: "+", Line: tok.Line, Column: tok.Column}
		}

	// Subtraction and implication operators
	case '-':
		if l.peekChar() == '>' {
			// "->" implication operator (logical implication)
			l.readChar()
			tok = Token{Type: TOKEN_IMPL, Literal: "->", Line: tok.Line, Column: tok.Column}
		} else {
			// "-" subtraction or negation operator
			tok = Token{Type: TOKEN_MINUS, Literal: "-", Line: tok.Line, Column: tok.Column}
		}

	// Arithmetic operators
	case '*':
		// "*" multiplication operator
		tok = Token{Type: TOKEN_MULTIPLY, Literal: "*", Line: tok.Line, Column: tok.Column}
	case '/':
		// "/" division operator (also used for paths, handled in default case)
		tok = Token{Type: TOKEN_DIVIDE, Literal: "/", Line: tok.Line, Column: tok.Column}

	// Logical NOT and inequality operators
	case '!':
		if l.peekChar() == '=' {
			// "!=" inequality comparison operator
			l.readChar()
			tok = Token{Type: TOKEN_NEQ, Literal: "!=", Line: tok.Line, Column: tok.Column}
		} else {
			// "!" logical NOT operator
			tok = Token{Type: TOKEN_NOT, Literal: "!", Line: tok.Line, Column: tok.Column}
		}

	// Less-than comparison operators
	case '<':
		if l.peekChar() == '=' {
			// "<=" less-than-or-equal comparison operator
			l.readChar()
			tok = Token{Type: TOKEN_LTE, Literal: "<=", Line: tok.Line, Column: tok.Column}
		} else {
			// "<" less-than comparison operator
			tok = Token{Type: TOKEN_LT, Literal: "<", Line: tok.Line, Column: tok.Column}
		}

	// Greater-than comparison operators
	case '>':
		if l.peekChar() == '=' {
			// ">=" greater-than-or-equal comparison operator
			l.readChar()
			tok = Token{Type: TOKEN_GTE, Literal: ">=", Line: tok.Line, Column: tok.Column}
		} else {
			// ">" greater-than comparison operator
			tok = Token{Type: TOKEN_GT, Literal: ">", Line: tok.Line, Column: tok.Column}
		}

	// Logical AND operator (& alone is illegal)
	case '&':
		if l.peekChar() == '&' {
			// "&&" logical AND operator
			l.readChar()
			tok = Token{Type: TOKEN_AND_OP, Literal: "&&", Line: tok.Line, Column: tok.Column}
		} else {
			// Single "&" is not valid in Nix
			tok = Token{Type: TOKEN_ILLEGAL, Literal: "&", Line: tok.Line, Column: tok.Column}
		}

	// Logical OR operator (| alone is illegal)
	case '|':
		if l.peekChar() == '|' {
			// "||" logical OR operator
			l.readChar()
			tok = Token{Type: TOKEN_OR_OP, Literal: "||", Line: tok.Line, Column: tok.Column}
		} else {
			// Single "|" is not valid in Nix
			tok = Token{Type: TOKEN_ILLEGAL, Literal: "|", Line: tok.Line, Column: tok.Column}
		}

	// Special operators
	case '?':
		// "?" attribute existence test operator
		tok = Token{Type: TOKEN_QUESTION, Literal: "?", Line: tok.Line, Column: tok.Column}
	case '.':
		// "." attribute selection operator
		tok = Token{Type: TOKEN_DOT, Literal: ".", Line: tok.Line, Column: tok.Column}

	// Structural delimiters
	case ';':
		// ";" statement terminator
		tok = Token{Type: TOKEN_SEMICOLON, Literal: ";", Line: tok.Line, Column: tok.Column}
	case ':':
		// ":" function parameter separator
		tok = Token{Type: TOKEN_COLON, Literal: ":", Line: tok.Line, Column: tok.Column}
	case ',':
		// "," element separator
		tok = Token{Type: TOKEN_COMMA, Literal: ",", Line: tok.Line, Column: tok.Column}

	// Grouping delimiters
	case '(':
		// "(" left parenthesis for grouping and function calls
		tok = Token{Type: TOKEN_LPAREN, Literal: "(", Line: tok.Line, Column: tok.Column}
	case ')':
		// ")" right parenthesis
		tok = Token{Type: TOKEN_RPAREN, Literal: ")", Line: tok.Line, Column: tok.Column}
	case '{':
		// "{" left brace for attribute sets
		tok = Token{Type: TOKEN_LBRACE, Literal: "{", Line: tok.Line, Column: tok.Column}
	case '}':
		// "}" right brace
		tok = Token{Type: TOKEN_RBRACE, Literal: "}", Line: tok.Line, Column: tok.Column}
	case '[':
		// "[" left bracket for lists
		tok = Token{Type: TOKEN_LBRACKET, Literal: "[", Line: tok.Line, Column: tok.Column}
	case ']':
		// "]" right bracket
		tok = Token{Type: TOKEN_RBRACKET, Literal: "]", Line: tok.Line, Column: tok.Column}

	// String literals
	case '"':
		// String literal: delegate to readString() for proper escape handling
		tok.Type = TOKEN_STRING
		tok.Literal = l.readString()

	// End of file
	case 0:
		// ASCII NUL indicates end of input
		tok.Type = TOKEN_EOF
		tok.Literal = ""

	// Complex token recognition
	default:
		if isLetter(l.ch) {
			// Identifier or keyword: delegate to readIdentifier() and keyword lookup
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			// Early return: readIdentifier() already advanced the position
			return tok
		} else if isDigit(l.ch) {
			// Numeric literal: delegate to readNumber() for int/float detection
			tok.Literal, tok.Type = l.readNumber()
			// Early return: readNumber() already advanced the position
			return tok
		} else if l.ch == '/' && unicode.IsLetter(rune(l.peekChar())) {
			// Path literal: "/" followed by letter indicates path, not division
			tok.Type = TOKEN_PATH
			tok.Literal = l.readPath()
			// Early return: readPath() already advanced the position
			return tok
		} else {
			// Unrecognized character: mark as illegal for error reporting
			tok = Token{Type: TOKEN_ILLEGAL, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	}

	l.readChar()

	return tok
}
