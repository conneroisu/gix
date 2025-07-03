// Package value provides the runtime value system for the Nix expression language interpreter.
//
// This package defines all value types that can result from evaluating Nix expressions.
// The value system is designed to be immutable, type-safe, and efficient.
//
// Core Design Principles:
//
// Immutability:
//
//	All values are immutable after creation. This enables safe concurrent access,
//	prevents unexpected mutations, and makes reasoning about program behavior easier.
//
// Type Safety:
//
//	Each value type implements the Value interface, providing type checking at runtime.
//	The Type() method allows for safe type discrimination and error reporting.
//
// Equality Semantics:
//
//	All values support structural equality comparison through the Equals() method.
//	This enables correct behavior for operators like == and !=.
//
// String Representation:
//
//	Every value can be converted to a human-readable string representation,
//	which is essential for debugging and the REPL interface.
//
// Value Types:
//
// Primitive Types:
//   - Null: The null value (singleton)
//   - Bool: Boolean values (true, false)
//   - Int: 64-bit signed integers
//   - Float: 64-bit floating-point numbers
//   - String: UTF-8 encoded strings
//   - Path: File system paths
//
// Composite Types:
//   - List: Ordered sequences of values
//   - Attrs: Key-value mappings (attribute sets)
//
// Functional Types:
//   - Function: User-defined functions with closures
//   - Builtin: Built-in functions implemented in Go
//
// The Environment interface provides variable scoping and binding management.
// It supports lexical scoping with proper closure semantics for functions.
//
// Memory Management:
//
//	Values use Go's garbage collector for automatic memory management.
//	Large composite values (lists, attribute sets) use structural sharing
//	where possible to reduce memory overhead.
//
// Performance Considerations:
//   - Primitive values are implemented as simple Go types for efficiency
//   - Composite values use efficient data structures (slices, maps)
//   - String concatenation and list operations are optimized
//   - Equality comparison is short-circuited for performance
//
// Thread Safety:
//
//	All value types are safe for concurrent read access due to immutability.
//	However, Environment implementations may require synchronization for
//	concurrent modifications.
package value
