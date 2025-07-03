package types

import (
	"fmt"
	"strconv"
	"strings"
)

// Node represents any node in the AST.
// All AST nodes must implement this interface.
type Node interface {
	// String returns a string representation of the node
	String() string

	// Position returns the source position of the node
	Position() SourcePos
}

// SourcePos represents a position in the source code.
type SourcePos struct {
	Line   int // 1-based line number
	Column int // 0-based column number
}

// Expr represents an expression node in the AST.
// All expression types must implement this interface.
type Expr interface {
	Node
	// exprNode is a marker method to ensure only expression types implement this interface
	exprNode()
}

// baseNode provides common functionality for all AST nodes.
type baseNode struct {
	pos SourcePos
}

func (n baseNode) Position() SourcePos { return n.pos }

// WithPos creates a new base node with the given position.
func WithPos(line, column int) baseNode {
	return baseNode{pos: SourcePos{Line: line, Column: column}}
}

// ============================================================================
// Literal Expressions
// ============================================================================

// IntExpr represents an integer literal.
type IntExpr struct {
	baseNode
	Value int64
}

func (e *IntExpr) String() string { return strconv.FormatInt(e.Value, 10) }
func (e *IntExpr) exprNode()      {}

// FloatExpr represents a floating-point literal.
type FloatExpr struct {
	baseNode
	Value float64
}

func (e *FloatExpr) String() string { return fmt.Sprintf("%g", e.Value) }
func (e *FloatExpr) exprNode()      {}

// StringExpr represents a string literal.
type StringExpr struct {
	baseNode
	Value string
	// IsIndented indicates if this was a '' string literal
	IsIndented bool
}

func (e *StringExpr) String() string {
	if e.IsIndented {
		return fmt.Sprintf("''%s''", e.Value)
	}

	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(e.Value, `"`, `\"`))
}
func (e *StringExpr) exprNode() {}

// BoolExpr represents a boolean literal (true/false).
type BoolExpr struct {
	baseNode
	Value bool
}

func (e *BoolExpr) String() string { return strconv.FormatBool(e.Value) }
func (e *BoolExpr) exprNode()      {}

// NullExpr represents the null literal.
type NullExpr struct {
	baseNode
}

func (e *NullExpr) String() string { return "null" }
func (e *NullExpr) exprNode()      {}

// PathExpr represents a path literal.
type PathExpr struct {
	baseNode
	Value string
	// IsAbsolute indicates if the path is absolute (/...) vs relative (./...)
	IsAbsolute bool
}

func (e *PathExpr) String() string { return e.Value }
func (e *PathExpr) exprNode()      {}

// IdentExpr represents an identifier (variable reference).
type IdentExpr struct {
	baseNode
	Name string
}

func (e *IdentExpr) String() string { return e.Name }
func (e *IdentExpr) exprNode()      {}

// ============================================================================
// Compound Expressions
// ============================================================================

// ListExpr represents a list literal [e1 e2 ... en].
type ListExpr struct {
	baseNode
	Elements []Expr
}

func (e *ListExpr) String() string {
	var elems []string
	for _, elem := range e.Elements {
		elems = append(elems, elem.String())
	}

	return fmt.Sprintf("[ %s ]", strings.Join(elems, " "))
}
func (e *ListExpr) exprNode() {}

// AttrSetExpr represents an attribute set { k1 = v1; k2 = v2; ... }.
type AttrSetExpr struct {
	baseNode
	Recursive bool // true for rec { ... }
	Bindings  []AttrBinding
	// Inherits represents inherit statements
	Inherits []InheritClause
}

func (e *AttrSetExpr) String() string {
	var parts []string
	if e.Recursive {
		parts = append(parts, "rec")
	}
	parts = append(parts, "{")

	var bindings []string
	for _, inherit := range e.Inherits {
		bindings = append(bindings, inherit.String())
	}
	for _, binding := range e.Bindings {
		bindings = append(bindings, fmt.Sprintf("%s = %s;",
			strings.Join(binding.Path, "."), binding.Value))
	}

	if len(bindings) > 0 {
		parts = append(parts, strings.Join(bindings, " "))
	}
	parts = append(parts, "}")

	return strings.Join(parts, " ")
}
func (e *AttrSetExpr) exprNode() {}

// AttrBinding represents a single binding in an attribute set.
type AttrBinding struct {
	Path  []string // Attribute path (e.g., ["a", "b"] for a.b = ...)
	Value Expr     // The value expression
}

// InheritClause represents an inherit statement.
type InheritClause struct {
	From  Expr     // nil for plain inherit, otherwise the source set
	Attrs []string // Attribute names to inherit
}

func (i InheritClause) String() string {
	if i.From == nil {
		return fmt.Sprintf("inherit %s;", strings.Join(i.Attrs, " "))
	}

	return fmt.Sprintf("inherit (%s) %s;", i.From, strings.Join(i.Attrs, " "))
}

// ============================================================================
// Operators
// ============================================================================

// BinaryOp represents a binary operator.
type BinaryOp int

const (
	OpAdd    BinaryOp = iota // +
	OpSub                    // -
	OpMul                    // *
	OpDiv                    // /
	OpConcat                 // ++
	OpEq                     // ==
	OpNEq                    // !=
	OpLT                     // <
	OpGT                     // >
	OpLTE                    // <=
	OpGTE                    // >=
	OpAnd                    // &&
	OpOr                     // ||
	OpImpl                   // ->
	OpUpdate                 // //
)

// String returns the string representation of the operator.
func (op BinaryOp) String() string {
	ops := []string{
		"+", "-", "*", "/", "++",
		"==", "!=", "<", ">", "<=", ">=",
		"&&", "||", "->", "//",
	}
	if int(op) < len(ops) {
		return ops[op]
	}

	return fmt.Sprintf("BinaryOp(%d)", op)
}

// BinaryExpr represents a binary operation.
type BinaryExpr struct {
	baseNode
	Left  Expr
	Op    BinaryOp
	Right Expr
}

func (e *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Op, e.Right)
}
func (e *BinaryExpr) exprNode() {}

// UnaryOp represents a unary operator.
type UnaryOp int

const (
	OpNot UnaryOp = iota // !
	OpNeg                // - (negation)
)

func (op UnaryOp) String() string {
	switch op {
	case OpNot:
		return "!"
	case OpNeg:
		return "-"
	default:
		return fmt.Sprintf("UnaryOp(%d)", op)
	}
}

// UnaryExpr represents a unary operation.
type UnaryExpr struct {
	baseNode
	Op   UnaryOp
	Expr Expr
}

func (e *UnaryExpr) String() string {
	return fmt.Sprintf("(%s%s)", e.Op, e.Expr)
}
func (e *UnaryExpr) exprNode() {}

// ============================================================================
// Control Flow Expressions
// ============================================================================

// IfExpr represents an if-then-else expression.
type IfExpr struct {
	baseNode
	Cond Expr // Condition
	Then Expr // Then branch
	Else Expr // Else branch
}

func (e *IfExpr) String() string {
	return fmt.Sprintf("if %s then %s else %s", e.Cond, e.Then, e.Else)
}
func (e *IfExpr) exprNode() {}

// LetExpr represents a let expression.
type LetExpr struct {
	baseNode
	Bindings []Binding // Variable bindings
	Body     Expr      // Body expression
}

func (e *LetExpr) String() string {
	var bindings []string
	for _, b := range e.Bindings {
		bindings = append(bindings, fmt.Sprintf("%s = %s;", b.Name, b.Value))
	}

	return fmt.Sprintf("let %s in %s", strings.Join(bindings, " "), e.Body)
}
func (e *LetExpr) exprNode() {}

// Binding represents a single binding in a let expression.
type Binding struct {
	Name  string
	Value Expr
}

// WithExpr represents a with expression.
type WithExpr struct {
	baseNode
	Expr Expr // The expression providing the scope
	Body Expr // The body where the scope is available
}

func (e *WithExpr) String() string {
	return fmt.Sprintf("with %s; %s", e.Expr, e.Body)
}
func (e *WithExpr) exprNode() {}

// AssertExpr represents an assert expression.
type AssertExpr struct {
	baseNode
	Cond Expr // Assertion condition
	Body Expr // Body to evaluate if assertion passes
}

func (e *AssertExpr) String() string {
	return fmt.Sprintf("assert %s; %s", e.Cond, e.Body)
}
func (e *AssertExpr) exprNode() {}

// ============================================================================
// Function-related Expressions
// ============================================================================

// FunctionExpr represents a function definition.
type FunctionExpr struct {
	baseNode
	Param   string   // Parameter name (for simple functions)
	Pattern *Pattern // Parameter pattern (for pattern matching)
	Body    Expr     // Function body
}

func (e *FunctionExpr) String() string {
	if e.Pattern != nil {
		return fmt.Sprintf("%s: %s", e.Pattern, e.Body)
	}

	return fmt.Sprintf("%s: %s", e.Param, e.Body)
}
func (e *FunctionExpr) exprNode() {}

// Pattern represents a function parameter pattern.
type Pattern struct {
	Type     PatternType
	Name     string   // For IdentPattern or the binding name in AttrSetPattern
	Attrs    []string // For AttrSetPattern
	Ellipsis bool     // For AttrSetPattern with ...
}

type PatternType int

const (
	IdentPattern   PatternType = iota // Simple identifier pattern
	AttrSetPattern                    // Attribute set pattern { x, y, ... }
)

func (p *Pattern) String() string {
	switch p.Type {
	case IdentPattern:
		return p.Name
	case AttrSetPattern:
		parts := append([]string{}, p.Attrs...)
		if p.Ellipsis {
			parts = append(parts, "...")
		}
		pattern := fmt.Sprintf("{ %s }", strings.Join(parts, ", "))
		if p.Name != "" {
			return fmt.Sprintf("%s @ %s", pattern, p.Name)
		}

		return pattern
	default:
		return fmt.Sprintf("Pattern(%d)", p.Type)
	}
}

// ApplyExpr represents function application.
type ApplyExpr struct {
	baseNode
	Func Expr // Function to apply
	Arg  Expr // Argument
}

func (e *ApplyExpr) String() string {
	return fmt.Sprintf("(%s %s)", e.Func, e.Arg)
}
func (e *ApplyExpr) exprNode() {}

// ============================================================================
// Attribute Access Expressions
// ============================================================================

// SelectExpr represents attribute selection (e.attrpath).
type SelectExpr struct {
	baseNode
	Expr     Expr     // Expression to select from
	AttrPath []string // Attribute path
	Default  Expr     // Default value (for 'or' expressions)
}

func (e *SelectExpr) String() string {
	s := fmt.Sprintf("%s.%s", e.Expr, strings.Join(e.AttrPath, "."))
	if e.Default != nil {
		s += fmt.Sprintf(" or %s", e.Default)
	}

	return s
}
func (e *SelectExpr) exprNode() {}

// HasAttrExpr represents attribute existence test (e ? attrpath).
type HasAttrExpr struct {
	baseNode
	Expr     Expr     // Expression to test
	AttrPath []string // Attribute path to check
}

func (e *HasAttrExpr) String() string {
	return fmt.Sprintf("%s ? %s", e.Expr, strings.Join(e.AttrPath, "."))
}
func (e *HasAttrExpr) exprNode() {}

// ============================================================================
// Special Expressions
// ============================================================================

// InheritExpr represents an inherit expression (for internal use)
// This is typically transformed into attribute bindings during parsing.
type InheritExpr struct {
	baseNode
	From  Expr     // Source expression (nil for plain inherit)
	Attrs []string // Attribute names to inherit
}

func (e *InheritExpr) String() string {
	if e.From == nil {
		return "inherit " + strings.Join(e.Attrs, " ")
	}

	return fmt.Sprintf("inherit (%s) %s", e.From, strings.Join(e.Attrs, " "))
}
func (e *InheritExpr) exprNode() {}
