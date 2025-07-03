package value

// Env implements the Environment interface with lexical scoping.
type Env struct {
	bindings map[string]Value
	parent   *Env
}

// NewEnv creates a new empty environment.
func NewEnv() *Env {
	return &Env{
		bindings: make(map[string]Value),
	}
}

// Get looks up a variable in the environment.
func (e *Env) Get(name string) (Value, bool) {
	if val, ok := e.bindings[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}

	return nil, false
}

// Set binds a variable in the current environment.
func (e *Env) Set(name string, value Value) {
	e.bindings[name] = value
}

// Extend creates a new child environment.
func (e *Env) Extend() Environment {
	return &Env{
		bindings: make(map[string]Value),
		parent:   e,
	}
}

// WithBindings creates a new environment with the given bindings.
func (e *Env) WithBindings(bindings map[string]Value) *Env {
	child := e.Extend().(*Env)
	for k, v := range bindings {
		child.Set(k, v)
	}

	return child
}

// Clone creates a shallow copy of the environment.
func (e *Env) Clone() *Env {
	bindings := make(map[string]Value)
	for k, v := range e.bindings {
		bindings[k] = v
	}

	return &Env{
		bindings: bindings,
		parent:   e.parent,
	}
}
