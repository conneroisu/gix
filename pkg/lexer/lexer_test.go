package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `let x = 5;
let y = 10;

if x > y then
  "x is greater"
else
  "y is greater"
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "5"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "y"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "10"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_IF, "if"},
		{TOKEN_IDENT, "x"},
		{TOKEN_GT, ">"},
		{TOKEN_IDENT, "y"},
		{TOKEN_THEN, "then"},
		{TOKEN_STRING, "x is greater"},
		{TOKEN_ELSE, "else"},
		{TOKEN_STRING, "y is greater"},
		{TOKEN_EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := "+-*/==!=<><=>=&&||->++?"

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_PLUS, "+"},
		{TOKEN_MINUS, "-"},
		{TOKEN_MULTIPLY, "*"},
		{TOKEN_DIVIDE, "/"},
		{TOKEN_EQ, "=="},
		{TOKEN_NEQ, "!="},
		{TOKEN_LT, "<"},
		{TOKEN_GT, ">"},
		{TOKEN_LTE, "<="},
		{TOKEN_GTE, ">="},
		{TOKEN_AND_OP, "&&"},
		{TOKEN_OR_OP, "||"},
		{TOKEN_IMPL, "->"},
		{TOKEN_CONCAT, "++"},
		{TOKEN_QUESTION, "?"},
		{TOKEN_EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNumbers(t *testing.T) {
	input := "123 3.14 0.5"

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_INT, "123"},
		{TOKEN_FLOAT, "3.14"},
		{TOKEN_FLOAT, "0.5"},
		{TOKEN_EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStrings(t *testing.T) {
	input := `"hello world" "escaped \"quote\""`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_STRING, "hello world"},
		{TOKEN_STRING, "escaped \\\"quote\\\""},
		{TOKEN_EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := "if then else let in with assert or and not rec inherit true false null"

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_IF, "if"},
		{TOKEN_THEN, "then"},
		{TOKEN_ELSE, "else"},
		{TOKEN_LET, "let"},
		{TOKEN_IN, "in"},
		{TOKEN_WITH, "with"},
		{TOKEN_ASSERT, "assert"},
		{TOKEN_OR, "or"},
		{TOKEN_AND, "and"},
		{TOKEN_NOT, "not"},
		{TOKEN_REC, "rec"},
		{TOKEN_INHERIT, "inherit"},
		{TOKEN_IDENT, "true"},  // true is parsed as identifier, converted to bool in parser
		{TOKEN_IDENT, "false"}, // false is parsed as identifier, converted to bool in parser
		{TOKEN_IDENT, "null"},  // null is parsed as identifier, converted to null in parser
		{TOKEN_EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `# This is a comment
let x = 5; # End of line comment
/* Multi-line
   comment */
let y = 10;`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "5"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "y"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "10"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
