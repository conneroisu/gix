package eval

import (
	"testing"

	"github.com/conneroisu/gix/internal/value"
	"github.com/conneroisu/gix/pkg/lexer"
	"github.com/conneroisu/gix/pkg/parser"
)

func testEval(input string) value.Value {
	l := lexer.New(input)
	p := parser.New(l)
	program, _ := p.Parse()
	e := New(".")
	result, _ := e.Eval(program)

	return result
}

func testIntegerObject(t *testing.T, obj value.Value, expected int64) bool {
	result, ok := obj.(value.Int)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)

		return false
	}
	if int64(result) != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result, expected)

		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj value.Value, expected bool) bool {
	result, ok := obj.(value.Bool)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)

		return false
	}
	if bool(result) != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result, expected)

		return false
	}

	return true
}

func testNullObject(t *testing.T, obj value.Value) bool {
	_, ok := obj.(value.Null)
	if !ok {
		t.Errorf("object is not Null. got=%T (%+v)", obj, obj)

		return false
	}

	return true
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) then 10 else 20", 10},
		{"if (false) then 10 else 20", 20},
		{"if (1) then 10 else 20", 10},
		{"if (1 < 2) then 10 else 20", 10},
		{"if (1 > 2) then 10 else 20", 20},
		{"if (1 > 2) then 10 else null", nil},
		{"if (false) then 10 else null", nil},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; in a", 5},
		{"let a = 5 * 5; in a", 25},
		{"let a = 5; b = a; in b", 5},
		{"let a = 5; b = a; c = a + b + 5; in c", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "x: x + 2"

	evaluated := testEval(input)
	fn, ok := evaluated.(*value.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if fn.Param() != "x" {
		t.Fatalf("function has wrong parameter. want='x', got=%q", fn.Param())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"(x: x) 5", 5},
		{"(x: x * x) 5", 25},
		{"(x: x + 6) 5", 11},
		{"(x: y: x + y) 5 10", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`length([])`, 0},
		{`length([1, 2, 3])`, 3},
		{`length("hello")`, 5},
		{`length("")`, 0},
		{`head([1, 2, 3])`, 1},
		{`tail([1, 2, 3])`, []int{2, 3}},
		{`toString(123)`, "123"},
		{`toString(true)`, "true"},
		{`isInt(5)`, true},
		{`isInt("hello")`, false},
		{`isList([1, 2])`, true},
		{`isList(5)`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case bool:
			testBooleanObject(t, evaluated, expected)
		case string:
			str, ok := evaluated.(value.String)
			if !ok {
				t.Errorf("object is not String. got=%T (%+v)", evaluated, evaluated)

				continue
			}
			if string(str) != expected {
				t.Errorf("String has wrong value. got=%q", str)
			}
		case []int:
			list, ok := evaluated.(*value.List)
			if !ok {
				t.Errorf("object is not List. got=%T (%+v)", evaluated, evaluated)

				continue
			}
			if list.Len() != len(expected) {
				t.Errorf("wrong number of list elements. want=%d, got=%d",
					len(expected), list.Len())

				continue
			}
			for i, elem := range expected {
				testIntegerObject(t, list.Get(i), int64(elem))
			}
		}
	}
}

func TestListLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*value.List)
	if !ok {
		t.Fatalf("object is not List. got=%T (%+v)", evaluated, evaluated)
	}

	if result.Len() != 3 {
		t.Fatalf("list has wrong num of elements. got=%d", result.Len())
	}

	testIntegerObject(t, result.Get(0), 1)
	testIntegerObject(t, result.Get(1), 4)
	testIntegerObject(t, result.Get(2), 6)
}

func TestAttributeSets(t *testing.T) {
	input := `{ foo = 5; bar = 10; }`

	evaluated := testEval(input)
	attrs, ok := evaluated.(*value.Attrs)
	if !ok {
		t.Fatalf("object is not Attrs. got=%T (%+v)", evaluated, evaluated)
	}

	expectedPairs := map[string]int64{
		"foo": 5,
		"bar": 10,
	}

	for expectedKey, expectedVal := range expectedPairs {
		val, ok := attrs.Get(expectedKey)
		if !ok {
			t.Errorf("no pair for given key. key=%q", expectedKey)

			continue
		}

		testIntegerObject(t, val, expectedVal)
	}
}

func TestDerivationBuiltin(t *testing.T) {
	input := `derivation { name = "hello"; builder = "/bin/sh"; }`

	evaluated := testEval(input)
	attrs, ok := evaluated.(*value.Attrs)
	if !ok {
		t.Fatalf("object is not Attrs. got=%T (%+v)", evaluated, evaluated)
	}

	// Check basic attributes
	nameVal, ok := attrs.Get("name")
	if !ok {
		t.Error("derivation missing 'name' attribute")
	} else {
		name, ok := nameVal.(value.String)
		if !ok || string(name) != "hello" {
			t.Errorf("derivation name wrong. got=%v", nameVal)
		}
	}

	builderVal, ok := attrs.Get("builder")
	if !ok {
		t.Error("derivation missing 'builder' attribute")
	} else {
		builder, ok := builderVal.(value.String)
		if !ok || string(builder) != "/bin/sh" {
			t.Errorf("derivation builder wrong. got=%v", builderVal)
		}
	}

	// Check that outputs exist
	outVal, ok := attrs.Get("out")
	if !ok {
		t.Error("derivation missing 'out' attribute")
	} else {
		_, ok := outVal.(value.String)
		if !ok {
			t.Errorf("derivation out not string. got=%T", outVal)
		}
	}
}
