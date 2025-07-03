package parser

import (
	"strconv"

	"github.com/conneroisu/gix/internal/types"
	"github.com/conneroisu/gix/pkg/lexer"
)

// Parser implements a recursive descent parser with Pratt parsing for Nix expressions.
// It transforms a stream of tokens from the lexer into an Abstract Syntax Tree (AST).
// The parser uses lookahead (cur/peek tokens) for disambiguation and precedence handling.
type Parser struct {
	l      *lexer.Lexer // The lexer providing the token stream
	cur    lexer.Token  // Current token being processed
	peek   lexer.Token  // Next token (lookahead for parsing decisions)
	errors *ParseErrors // Accumulated parsing errors for comprehensive reporting
}

// New creates a new parser instance from a lexer.
// The parser is initialized with the first two tokens (cur and peek) to enable
// immediate parsing with proper lookahead. This two-token window is essential
// for distinguishing ambiguous constructs and implementing operator precedence.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: &ParseErrors{}, // Initialize empty error collection
	}
	// Prime the parser by reading the first two tokens
	// This establishes the cur/peek window needed for parsing decisions
	p.advance() // Sets cur to first token, peek to second
	p.advance() // Sets cur to second token, peek to third

	return p
}

// Parse is the main entry point for parsing a complete Nix expression.
// It parses the entire token stream into a single expression AST starting
// with the lowest precedence level. Returns either the parsed AST or
// accumulated parsing errors for comprehensive error reporting.
func (p *Parser) Parse() (types.Expr, error) {
	// Start parsing with lowest precedence to capture the entire expression
	expr := p.parseExpression(precedenceLowest)

	// Check if any errors were encountered during parsing
	if p.errors.HasErrors() {
		// Return accumulated errors for detailed error reporting
		return nil, p.errors
	}

	// Successfully parsed expression
	return expr, nil
}

// Errors returns a slice of error messages from parsing failures.
// This method provides a simple string-based interface for error reporting,
// converting the internal ParseErrors structure to human-readable messages.
// Useful for legacy compatibility and simple error display.
func (p *Parser) Errors() []string {
	// Pre-allocate slice with known capacity for efficiency
	msgs := make([]string, 0, p.errors.Count())
	// Convert each ParseError to its string representation
	for _, err := range p.errors.Errors() {
		msgs = append(msgs, err.Error())
	}

	return msgs
}

// advance shifts the token window forward by one position.
// This maintains the cur/peek lookahead pattern essential for parsing:
// - cur becomes the previous peek token
// - peek becomes the next token from the lexer
// This method is called after successfully consuming a token.
func (p *Parser) advance() {
	// Shift the lookahead window forward
	p.cur = p.peek           // Current token becomes previous peek
	p.peek = p.l.NextToken() // Get next token from lexer
}

// parseExpression implements the core Pratt parsing algorithm for expressions.
// This method handles operator precedence and associativity by:
// 1. Parsing a prefix expression (literals, identifiers, unary ops, etc.)
// 2. Continuously parsing infix operations while precedence allows
// 3. Supporting function application as a special infix operation
//
// The precedence parameter controls how tightly this expression binds,
// enabling proper handling of complex expressions like: a + b * c && d.
func (p *Parser) parseExpression(precedence int) types.Expr {
	// Phase 1: Parse the initial prefix expression (required)
	prefix := p.parsePrefixExpression()
	if prefix == nil {
		// Prefix parsing failed - this is a parsing error
		return nil
	}

	// Phase 2: Parse infix operations while precedence and tokens allow
	// Continue until we hit a statement terminator or lower precedence
	for !p.peekIs(lexer.TOKEN_SEMICOLON) && !p.peekIs(lexer.TOKEN_EOF) {
		// Check precedence: stop if the next operator has lower precedence
		// Special case: function application needs special handling
		if precedence >= p.peekPrecedence() && !p.couldBeArgument() {
			break // Lower precedence - let parent expression handle it
		}

		// Determine the type of infix operation to parse
		if p.isInfixOperator(p.peek.Type) {
			// Standard binary operator (==, +, &&, etc.)
			p.advance()
			prefix = p.parseInfixExpression(prefix)
		} else if p.couldBeArgument() && precedence < precedenceCall {
			// Function application: juxtaposition of expressions
			// e.g., "f x" where f is a function and x is an argument
			p.advance()
			prefix = p.parseFunctionApplication(prefix)
		} else {
			// No valid infix operation - end expression parsing
			break
		}
	}

	return prefix
}

// parsePrefixExpression handles expressions that begin with a prefix element.
// This includes:
// - Literals (numbers, strings, paths, booleans, null)
// - Identifiers (variables and function parameters)
// - Keywords (if, let, with, assert for control flow)
// - Unary operators (-, ! for negation and logical NOT)
// - Compound expressions ({...}, [...], (...) for grouping)
//
// This is the "nud" (null denotation) function in Pratt parsing terminology.
func (p *Parser) parsePrefixExpression() types.Expr {
	switch p.cur.Type {
	// Literal values - direct value representations
	case lexer.TOKEN_INT:
		return p.parseInt() // Integer literals: 42, -10, 0
	case lexer.TOKEN_FLOAT:
		return p.parseFloat() // Float literals: 3.14, -0.5
	case lexer.TOKEN_STRING:
		return p.parseString() // String literals: "hello", "world"
	case lexer.TOKEN_PATH:
		return p.parsePath() // Path literals: ./file, /absolute
	case lexer.TOKEN_IDENT:
		// Identifiers or function definitions (x, variable, x: x + 1)
		return p.parseIdentifierOrFunction()

	// Control flow keywords - complex expressions that modify evaluation
	case lexer.TOKEN_IF:
		return p.parseIf() // Conditional expressions: if cond then a else b
	case lexer.TOKEN_LET:
		return p.parseLet() // Let bindings: let x = 1; in x + 2
	case lexer.TOKEN_WITH:
		return p.parseWith() // Scope extension: with attrs; expr
	case lexer.TOKEN_ASSERT:
		return p.parseAssert() // Assertions: assert condition; expr

	// Unary prefix operators - operations on single operands
	case lexer.TOKEN_NOT:
		return p.parseUnary(types.OpNot) // Logical negation: !expr
	case lexer.TOKEN_MINUS:
		return p.parseUnary(types.OpNeg) // Arithmetic negation: -expr

	// Compound data structures and grouping
	case lexer.TOKEN_LBRACE:
		return p.parseAttrSet() // Attribute sets: { x = 1; y = 2; }
	case lexer.TOKEN_LBRACKET:
		return p.parseList() // Lists: [1, 2, 3]
	case lexer.TOKEN_LPAREN:
		return p.parseGrouped() // Grouped expressions: (expr)

	default:
		// Unrecognized token - record error and fail gracefully
		p.errors.Addf(p.cur.Line, p.cur.Column,
			"no prefix parse function for %v", p.cur.Type)

		return nil
	}
}

// parseInfixExpression handles binary operators and special infix operations.
// This is the "led" (left denotation) function in Pratt parsing terminology.
// It takes the left operand and parses the right operand according to the
// operator's precedence and associativity rules.
//
// The function handles all binary operators plus special operations like
// attribute selection (.) and existence testing (?).
func (p *Parser) parseInfixExpression(left types.Expr) types.Expr {
	switch p.cur.Type {
	// Arithmetic operators - mathematical operations on numbers
	case lexer.TOKEN_PLUS:
		return p.parseBinary(left, types.OpAdd) // Addition: a + b
	case lexer.TOKEN_MINUS:
		return p.parseBinary(left, types.OpSub) // Subtraction: a - b
	case lexer.TOKEN_MULTIPLY:
		return p.parseBinary(left, types.OpMul) // Multiplication: a * b
	case lexer.TOKEN_DIVIDE:
		return p.parseBinary(left, types.OpDiv) // Division: a / b

	// Concatenation operator - joining sequences
	case lexer.TOKEN_CONCAT:
		return p.parseBinary(left, types.OpConcat) // List/string concat: a ++ b

	// Comparison operators - relational comparisons
	case lexer.TOKEN_EQ:
		return p.parseBinary(left, types.OpEq) // Equality: a == b
	case lexer.TOKEN_NEQ:
		return p.parseBinary(left, types.OpNEq) // Inequality: a != b
	case lexer.TOKEN_LT:
		return p.parseBinary(left, types.OpLT) // Less than: a < b
	case lexer.TOKEN_GT:
		return p.parseBinary(left, types.OpGT) // Greater than: a > b
	case lexer.TOKEN_LTE:
		return p.parseBinary(left, types.OpLTE) // Less/equal: a <= b
	case lexer.TOKEN_GTE:
		return p.parseBinary(left, types.OpGTE) // Greater/equal: a >= b

	// Logical operators - boolean operations with short-circuit evaluation
	case lexer.TOKEN_AND:
		return p.parseBinary(left, types.OpAnd) // Logical AND: a && b
	case lexer.TOKEN_OR_OP:
		return p.parseBinary(left, types.OpOr) // Logical OR: a || b
	case lexer.TOKEN_IMPL:
		return p.parseBinary(left, types.OpImpl) // Implication: a -> b

	// Special attribute operations - Nix-specific operators
	case lexer.TOKEN_DOT:
		return p.parseSelect(left) // Attribute selection: obj.attr
	case lexer.TOKEN_QUESTION:
		return p.parseHasAttr(left) // Existence test: obj ? attr
	case lexer.TOKEN_OR:
		return p.parseOrDefault(left) // Default value: obj.attr or default

	default:
		// Unrecognized infix operator - record error
		p.errors.Addf(p.cur.Line, p.cur.Column,
			"no infix parse function for %v", p.cur.Type)

		return nil
	}
}

// parseInt parses integer literals from token text to AST nodes.
// Converts the string representation ("42", "-10") to a 64-bit signed integer.
// Reports parsing errors with precise location information for user feedback.
func (p *Parser) parseInt() types.Expr {
	// Convert string literal to integer value
	val, err := strconv.ParseInt(p.cur.Literal, 10, 64)
	if err != nil {
		// Integer parsing failed - likely overflow or invalid format
		p.errors.Addf(p.cur.Line, p.cur.Column,
			"could not parse %q as integer", p.cur.Literal)

		return nil
	}

	// Create integer expression AST node
	return &types.IntExpr{Value: val}
}

// parseFloat parses floating-point literals from token text to AST nodes.
// Converts string representations ("3.14", "-0.5") to 64-bit floating-point values.
// Handles standard decimal notation with proper error reporting.
func (p *Parser) parseFloat() types.Expr {
	// Convert string literal to float value
	val, err := strconv.ParseFloat(p.cur.Literal, 64)
	if err != nil {
		// Float parsing failed - likely invalid format or overflow
		p.errors.Addf(p.cur.Line, p.cur.Column,
			"could not parse %q as float", p.cur.Literal)

		return nil
	}

	// Create floating-point expression AST node
	return &types.FloatExpr{Value: val}
}

// parseString creates string literal AST nodes from token text.
// The lexer has already processed escape sequences and removed quotes,
// so we can directly use the literal value from the token.
func (p *Parser) parseString() types.Expr {
	// String literal is ready to use (lexer handled escapes and quotes)
	return &types.StringExpr{Value: p.cur.Literal}
}

// parsePath creates path literal AST nodes from token text.
// Path literals represent file system paths and are used for imports
// and file references. Examples: ./file.nix, /etc/nixos/configuration.nix.
func (p *Parser) parsePath() types.Expr {
	// Path literal is ready to use as-is from lexer
	return &types.PathExpr{Value: p.cur.Literal}
}

// parseIdentifierOrFunction handles identifiers that might be special values or functions.
// This method disambiguates between:
// - Boolean literals (true, false)
// - Null literal (null)
// - Function definitions (param: body)
// - Regular variable references (name)
//
// The disambiguation uses lookahead to detect function syntax (identifier : expression).
func (p *Parser) parseIdentifierOrFunction() types.Expr {
	// Check for special literal identifiers that represent built-in values
	switch p.cur.Literal {
	case "true":
		// Boolean true literal
		return &types.BoolExpr{Value: true}
	case "false":
		// Boolean false literal
		return &types.BoolExpr{Value: false}
	case "null":
		// Null literal
		return &types.NullExpr{}
	}

	// Check if this is a function definition using lookahead
	if p.peekIs(lexer.TOKEN_COLON) {
		// "identifier :" pattern indicates function definition
		return p.parseFunction()
	}

	// Regular identifier (variable reference)
	return &types.IdentExpr{Name: p.cur.Literal}
}

// Helper methods for token inspection and parser state management.

// curIs checks if the current token matches the specified type.
// Used for token-based parsing decisions without consuming the token.
func (p *Parser) curIs(t lexer.TokenType) bool {
	return p.cur.Type == t
}

// peekIs checks if the next token (lookahead) matches the specified type.
// Essential for disambiguation and determining parsing paths without commitment.
func (p *Parser) peekIs(t lexer.TokenType) bool {
	return p.peek.Type == t
}

// expectPeek verifies that the next token matches the expected type and consumes it.
// This is a common pattern in recursive descent parsing for mandatory token sequences.
// If the expected token is found, advances to it and returns true.
// If not found, records a parsing error with precise location and returns false.
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekIs(t) {
		// Expected token found - consume it and continue
		p.advance()

		return true
	}
	// Expected token not found - record detailed error
	p.errors.Addf(p.peek.Line, p.peek.Column,
		"expected next token to be %v, got %v", t, p.peek.Type)

	return false
}

// peekPrecedence returns the precedence level of the next token.
// Used in Pratt parsing to determine when to stop parsing the current expression
// and let a higher-precedence operation take over. Returns the lowest precedence
// for tokens that aren't operators, allowing them to terminate expressions.
func (p *Parser) peekPrecedence() int {
	if prec, ok := precedenceMap[p.peek.Type]; ok {
		// Token has defined precedence (it's an operator)
		return prec
	}
	// Non-operator token gets lowest precedence
	return precedenceLowest
}

// curPrecedence returns the precedence level of the current token.
// Less commonly used than peekPrecedence, but necessary for certain
// parsing decisions where we need to know the precedence of the operator
// we're currently processing.
func (p *Parser) curPrecedence() int {
	if prec, ok := precedenceMap[p.cur.Type]; ok {
		// Current token has defined precedence
		return prec
	}
	// Non-operator token gets lowest precedence
	return precedenceLowest
}

// isInfixOperator determines if a token type represents a binary/infix operator.
// This check is used to distinguish between:
// - Infix operators that need special parsing (=, +, &&, etc.)
// - Other tokens that might appear between expressions
// Any token with defined precedence is considered an infix operator.
func (p *Parser) isInfixOperator(t lexer.TokenType) bool {
	// Operators have entries in the precedence map
	_, ok := precedenceMap[t]

	return ok
}

// couldBeArgument determines if the next token could start a function argument.
// This is essential for parsing function application (f x) vs other binary operations.
// Function application in Nix is implicit (no parentheses required), so we need
// to distinguish "f x" (application) from "f + x" (addition).
//
// Returns true for tokens that can begin expressions suitable as function arguments.
func (p *Parser) couldBeArgument() bool {
	switch p.peek.Type {
	// Literal values that can be function arguments
	case lexer.TOKEN_INT, lexer.TOKEN_FLOAT, lexer.TOKEN_STRING, lexer.TOKEN_PATH,
		// Identifiers and compound expressions
		lexer.TOKEN_IDENT, lexer.TOKEN_LBRACE, lexer.TOKEN_LBRACKET, lexer.TOKEN_LPAREN,
		// Unary operators and control flow (can start expressions)
		lexer.TOKEN_NOT, lexer.TOKEN_MINUS, lexer.TOKEN_IF, lexer.TOKEN_LET,
		lexer.TOKEN_WITH, lexer.TOKEN_ASSERT:
		return true
	default:
		// Operators, delimiters, EOF, etc. cannot start arguments
		return false
	}
}
