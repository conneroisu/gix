package parser

import (
	"testing"

	"github.com/conneroisu/gix/internal/types"
	"github.com/conneroisu/gix/pkg/lexer"
)

func testIntegerLiteral(t *testing.T, il types.Expr, value int64) bool {
	integ, ok := il.(*types.IntExpr)
	if !ok {
		t.Errorf("il not *types.IntExpr. got=%T", il)

		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)

		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp types.Expr, value string) bool {
	ident, ok := exp.(*types.IdentExpr)
	if !ok {
		t.Errorf("exp not *types.IdentExpr. got=%T", exp)

		return false
	}

	if ident.Name != value {
		t.Errorf("ident.Name not %s. got=%s", value, ident.Name)

		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp types.Expr, value bool) bool {
	bo, ok := exp.(*types.BoolExpr)
	if !ok {
		t.Errorf("exp not *types.BoolExpr. got=%T", exp)

		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)

		return false
	}

	return true
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	l := lexer.New(input)
	p := New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if !testIntegerLiteral(t, program, 5) {
		return
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	p := New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if !testIdentifier(t, program, "foobar") {
		return
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program, err := p.Parse()
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		if !testBooleanLiteral(t, program, tt.expected) {
			return
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator types.UnaryOp
		value    interface{}
	}{
		{"!true", types.OpNot, true},
		{"!false", types.OpNot, false},
		{"-15", types.OpNeg, 15},
		{"-20", types.OpNeg, 20},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program, err := p.Parse()
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		exp, ok := program.(*types.UnaryExpr)
		if !ok {
			t.Fatalf("program not *types.UnaryExpr. got=%T", program)
		}

		if exp.Op != tt.operator {
			t.Fatalf("exp.Op is not %v. got=%v", tt.operator, exp.Op)
		}

		switch v := tt.value.(type) {
		case int:
			if !testIntegerLiteral(t, exp.Expr, int64(v)) {
				return
			}
		case bool:
			if !testBooleanLiteral(t, exp.Expr, v) {
				return
			}
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   types.BinaryOp
		rightValue interface{}
	}{
		{"5 + 5", 5, types.OpAdd, 5},
		{"5 - 5", 5, types.OpSub, 5},
		{"5 * 5", 5, types.OpMul, 5},
		{"5 / 5", 5, types.OpDiv, 5},
		{"5 > 5", 5, types.OpGT, 5},
		{"5 < 5", 5, types.OpLT, 5},
		{"5 == 5", 5, types.OpEq, 5},
		{"5 != 5", 5, types.OpNEq, 5},
		{"true == true", true, types.OpEq, true},
		{"true != false", true, types.OpNEq, false},
		{"false == false", false, types.OpEq, false},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program, err := p.Parse()
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		exp, ok := program.(*types.BinaryExpr)
		if !ok {
			t.Fatalf("program is not *types.BinaryExpr. got=%T", program)
		}

		switch v := tt.leftValue.(type) {
		case int:
			if !testIntegerLiteral(t, exp.Left, int64(v)) {
				return
			}
		case bool:
			if !testBooleanLiteral(t, exp.Left, v) {
				return
			}
		}

		if exp.Op != tt.operator {
			t.Fatalf("exp.Op is not %v. got=%v", tt.operator, exp.Op)
		}

		switch v := tt.rightValue.(type) {
		case int:
			if !testIntegerLiteral(t, exp.Right, int64(v)) {
				return
			}
		case bool:
			if !testBooleanLiteral(t, exp.Right, v) {
				return
			}
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) then x else y`

	l := lexer.New(input)
	p := New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	exp, ok := program.(*types.IfExpr)
	if !ok {
		t.Fatalf("program not *types.IfExpr. got=%T", program)
	}

	if !testInfixExpression(t, exp.Cond, "x", types.OpLT, "y") {
		return
	}

	if !testIdentifier(t, exp.Then, "x") {
		return
	}

	if !testIdentifier(t, exp.Else, "y") {
		return
	}
}

func testInfixExpression(t *testing.T, exp types.Expr, left interface{},
	operator types.BinaryOp, right interface{}) bool {
	opExp, ok := exp.(*types.BinaryExpr)
	if !ok {
		t.Errorf("exp is not *types.BinaryExpr. got=%T(%s)", exp, exp)

		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Op != operator {
		t.Errorf("exp.Op is not %v. got=%v", operator, opExp.Op)

		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp types.Expr, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)

	return false
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `x: x + 2`

	l := lexer.New(input)
	p := New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	function, ok := program.(*types.FunctionExpr)
	if !ok {
		t.Fatalf("program not *types.FunctionExpr. got=%T", program)
	}

	if function.Param != "x" {
		t.Fatalf("function.Param not 'x'. got=%q", function.Param)
	}

	if !testInfixExpression(t, function.Body, "x", types.OpAdd, 2) {
		return
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add 1 2"

	l := lexer.New(input)
	p := New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	// In Nix, function application is left-associative: add 1 2 is parsed as (add 1) 2
	exp, ok := program.(*types.ApplyExpr)
	if !ok {
		t.Fatalf("program not *types.ApplyExpr. got=%T", program)
	}

	// exp.Func should be (add 1)
	innerExp, ok := exp.Func.(*types.ApplyExpr)
	if !ok {
		t.Fatalf("exp.Func not *types.ApplyExpr. got=%T", exp.Func)
	}

	if !testIdentifier(t, innerExp.Func, "add") {
		return
	}

	if !testIntegerLiteral(t, innerExp.Arg, 1) {
		return
	}

	if !testIntegerLiteral(t, exp.Arg, 2) {
		return
	}
}

func TestListLiterals(t *testing.T) {
	input := "[1, 2 3]"

	l := lexer.New(input)
	p := New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	list, ok := program.(*types.ListExpr)
	if !ok {
		t.Fatalf("program not *types.ListExpr. got=%T", program)
	}

	if len(list.Elements) != 3 {
		t.Fatalf("len(list.Elements) not 3. got=%d", len(list.Elements))
	}

	testIntegerLiteral(t, list.Elements[0], 1)
	testIntegerLiteral(t, list.Elements[1], 2)
	testIntegerLiteral(t, list.Elements[2], 3)
}

func TestLetExpressions(t *testing.T) {
	input := `let x = 5; y = 10; in x + y`

	l := lexer.New(input)
	p := New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	letExp, ok := program.(*types.LetExpr)
	if !ok {
		t.Fatalf("program not *types.LetExpr. got=%T", program)
	}

	if len(letExp.Bindings) != 2 {
		t.Fatalf("len(letExp.Bindings) not 2. got=%d", len(letExp.Bindings))
	}

	if letExp.Bindings[0].Name != "x" {
		t.Errorf("letExp.Bindings[0].Name not 'x'. got=%q", letExp.Bindings[0].Name)
	}

	if !testIntegerLiteral(t, letExp.Bindings[0].Value, 5) {
		return
	}

	if letExp.Bindings[1].Name != "y" {
		t.Errorf("letExp.Bindings[1].Name not 'y'. got=%q", letExp.Bindings[1].Name)
	}

	if !testIntegerLiteral(t, letExp.Bindings[1].Value, 10) {
		return
	}

	if !testInfixExpression(t, letExp.Body, "x", types.OpAdd, "y") {
		return
	}
}
