package value

import (
	"fmt"
	"sort"
	"strings"
)

// Type represents the type of a Nix value.
type Type byte

const (
	TypeNull Type = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypePath
	TypeList
	TypeAttrs
	TypeFunction
	TypeBuiltin
)

// Value is the interface all Nix values must implement.
type Value interface {
	Type() Type
	String() string
	Equals(Value) bool
}

// Null represents the null value.
type Null struct{}

func (Null) Type() Type     { return TypeNull }
func (Null) String() string { return "null" }
func (Null) Equals(v Value) bool {
	_, ok := v.(Null)

	return ok
}

// Bool represents a boolean value.
type Bool bool

func (b Bool) Type() Type     { return TypeBool }
func (b Bool) String() string { return fmt.Sprintf("%t", b) }
func (b Bool) Equals(v Value) bool {
	other, ok := v.(Bool)

	return ok && b == other
}

// Int represents an integer value.
type Int int64

func (i Int) Type() Type     { return TypeInt }
func (i Int) String() string { return fmt.Sprintf("%d", i) }
func (i Int) Equals(v Value) bool {
	other, ok := v.(Int)

	return ok && i == other
}

// Float represents a floating-point value.
type Float float64

func (f Float) Type() Type     { return TypeFloat }
func (f Float) String() string { return fmt.Sprintf("%g", f) }
func (f Float) Equals(v Value) bool {
	other, ok := v.(Float)

	return ok && f == other
}

// String represents a string value.
type String string

func (s String) Type() Type     { return TypeString }
func (s String) String() string { return fmt.Sprintf(`"%s"`, string(s)) }
func (s String) Equals(v Value) bool {
	other, ok := v.(String)

	return ok && s == other
}

// Path represents a path value.
type Path string

func (p Path) Type() Type     { return TypePath }
func (p Path) String() string { return string(p) }
func (p Path) Equals(v Value) bool {
	other, ok := v.(Path)

	return ok && p == other
}

// List represents a list value.
type List struct {
	elems []Value
}

// NewList creates a new list from elements.
func NewList(elems ...Value) *List {
	return &List{elems: append([]Value(nil), elems...)}
}

func (l *List) Type() Type { return TypeList }
func (l *List) Len() int   { return len(l.elems) }
func (l *List) Get(i int) Value {
	if i >= 0 && i < len(l.elems) {
		return l.elems[i]
	}

	return Null{}
}
func (l *List) Elements() []Value { return append([]Value(nil), l.elems...) }

func (l *List) String() string {
	parts := make([]string, len(l.elems))
	for i, elem := range l.elems {
		parts[i] = elem.String()
	}

	return fmt.Sprintf("[ %s ]", strings.Join(parts, " "))
}

func (l *List) Equals(v Value) bool {
	other, ok := v.(*List)
	if !ok || len(l.elems) != len(other.elems) {
		return false
	}
	for i, e := range l.elems {
		if !e.Equals(other.elems[i]) {
			return false
		}
	}

	return true
}

// Attrs represents an attribute set.
type Attrs struct {
	attrs map[string]Value
}

// NewAttrs creates a new attribute set.
func NewAttrs() *Attrs {
	return &Attrs{attrs: make(map[string]Value)}
}

// NewAttrsFrom creates an attribute set from a map.
func NewAttrsFrom(m map[string]Value) *Attrs {
	a := NewAttrs()
	for k, v := range m {
		a.Set(k, v)
	}

	return a
}

func (a *Attrs) Type() Type { return TypeAttrs }
func (a *Attrs) Len() int   { return len(a.attrs) }

func (a *Attrs) Get(key string) (Value, bool) {
	v, ok := a.attrs[key]

	return v, ok
}

func (a *Attrs) Set(key string, val Value) {
	a.attrs[key] = val
}

func (a *Attrs) Keys() []string {
	keys := make([]string, 0, len(a.attrs))
	for k := range a.attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func (a *Attrs) String() string {
	if len(a.attrs) == 0 {
		return "{ }"
	}

	keys := a.Keys()
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s = %s;", k, a.attrs[k])
	}

	return fmt.Sprintf("{ %s }", strings.Join(parts, " "))
}

func (a *Attrs) Equals(v Value) bool {
	other, ok := v.(*Attrs)
	if !ok || len(a.attrs) != len(other.attrs) {
		return false
	}
	for k, v := range a.attrs {
		otherV, ok := other.attrs[k]
		if !ok || !v.Equals(otherV) {
			return false
		}
	}

	return true
}

// Function represents a user-defined function.
type Function struct {
	param string
	body  interface{} // AST node
	env   Environment
}

// NewFunction creates a new function.
func NewFunction(param string, body interface{}, env Environment) *Function {
	return &Function{param: param, body: body, env: env}
}

func (f *Function) Type() Type        { return TypeFunction }
func (f *Function) String() string    { return fmt.Sprintf("<LAMBDA %s>", f.param) }
func (f *Function) Equals(Value) bool { return false } // Functions are not comparable
func (f *Function) Param() string     { return f.param }
func (f *Function) Body() interface{} { return f.body }
func (f *Function) Env() Environment  { return f.env }

// Builtin represents a built-in function.
type Builtin struct {
	name string
	fn   func([]Value) (Value, error)
}

// NewBuiltin creates a new builtin function.
func NewBuiltin(name string, fn func([]Value) (Value, error)) *Builtin {
	return &Builtin{name: name, fn: fn}
}

func (b *Builtin) Type() Type     { return TypeBuiltin }
func (b *Builtin) String() string { return fmt.Sprintf("<BUILTIN %s>", b.name) }
func (b *Builtin) Equals(v Value) bool {
	other, ok := v.(*Builtin)

	return ok && b.name == other.name
}
func (b *Builtin) Name() string                      { return b.name }
func (b *Builtin) Apply(args []Value) (Value, error) { return b.fn(args) }

// Environment represents variable bindings.
type Environment interface {
	Get(name string) (Value, bool)
	Set(name string, value Value)
	Extend() Environment
}

// Constructors for convenience.
func MakeNull() Value           { return Null{} }
func MakeBool(b bool) Value     { return Bool(b) }
func MakeInt(i int64) Value     { return Int(i) }
func MakeFloat(f float64) Value { return Float(f) }
func MakeString(s string) Value { return String(s) }
func MakePath(p string) Value   { return Path(p) }
