package parser

import (
	"fmt"
	"strings"
)

// ParseError represents a parsing error with location information.
type ParseError struct {
	Message string
	Line    int
	Column  int
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// ParseErrors is a collection of parse errors.
type ParseErrors struct {
	errors []ParseError
}

// Add adds a new parse error.
func (p *ParseErrors) Add(msg string, line, column int) {
	p.errors = append(p.errors, ParseError{
		Message: msg,
		Line:    line,
		Column:  column,
	})
}

// Addf adds a new parse error with formatting.
func (p *ParseErrors) Addf(line, column int, format string, args ...interface{}) {
	p.Add(fmt.Sprintf(format, args...), line, column)
}

// HasErrors returns true if there are any errors.
func (p *ParseErrors) HasErrors() bool {
	return len(p.errors) > 0
}

// Count returns the number of errors.
func (p *ParseErrors) Count() int {
	return len(p.errors)
}

// Errors returns all errors as a slice.
func (p *ParseErrors) Errors() []ParseError {
	return p.errors
}

// Error implements the error interface.
func (p *ParseErrors) Error() string {
	if len(p.errors) == 0 {
		return "no errors"
	}
	if len(p.errors) == 1 {
		return p.errors[0].Error()
	}

	var msgs []string
	for _, err := range p.errors {
		msgs = append(msgs, err.Error())
	}

	return fmt.Sprintf("%d parse errors:\n%s", len(p.errors), strings.Join(msgs, "\n"))
}

// First returns the first error, or nil if there are no errors.
func (p *ParseErrors) First() error {
	if len(p.errors) == 0 {
		return nil
	}

	return p.errors[0]
}
