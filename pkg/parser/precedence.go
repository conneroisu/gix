package parser

import "github.com/conneroisu/gix/pkg/lexer"

// Operator precedence levels.
const (
	precedenceLowest  = iota
	precedenceImpl    // ->
	precedenceOr      // ||
	precedenceAnd     // &&
	precedenceEquals  // == !=
	precedenceCompare // < > <= >=
	precedenceUpdate  // //
	precedenceConcat  // ++
	precedenceSum     // + -
	precedenceProduct // * /
	precedenceCall    // function application
	precedenceSelect  // . attribute selection
)

// precedenceMap maps token types to their precedence.
var precedenceMap = map[lexer.TokenType]int{
	lexer.TOKEN_IMPL:     precedenceImpl,
	lexer.TOKEN_OR_OP:    precedenceOr,
	lexer.TOKEN_AND:      precedenceAnd,
	lexer.TOKEN_EQ:       precedenceEquals,
	lexer.TOKEN_NEQ:      precedenceEquals,
	lexer.TOKEN_LT:       precedenceCompare,
	lexer.TOKEN_GT:       precedenceCompare,
	lexer.TOKEN_LTE:      precedenceCompare,
	lexer.TOKEN_GTE:      precedenceCompare,
	lexer.TOKEN_CONCAT:   precedenceConcat,
	lexer.TOKEN_PLUS:     precedenceSum,
	lexer.TOKEN_MINUS:    precedenceSum,
	lexer.TOKEN_MULTIPLY: precedenceProduct,
	lexer.TOKEN_DIVIDE:   precedenceProduct,
	lexer.TOKEN_DOT:      precedenceSelect,
}
