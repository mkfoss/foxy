package foxy

// todo: redesign error system, to have own Error interfaceand then separate error types that implement that interface,
// like in FieldError

import "fmt"

// Errorer is the interface that all foxy errors implement.
// It provides methods for wrapping other errors and adding context.
type Errorer interface {
	Error() string
	Unwrap() error
	SetWrapped(error) Errorer
	SetContext(context string) Errorer
}

// Error is a basic foxy error implementation.
type Error struct {
	context string
	message string
	wrapped error
}

func (fer *Error) Error() string {
	if fer.context != "" {
		return fer.context + ": " + fer.message
	}
	return fer.message
}

func (fer *Error) Unwrap() error {
	return fer.wrapped
}

func (fer *Error) SetWrapped(err error) Errorer {
	fer.wrapped = err
	return fer
}

func (fer *Error) SetContext(context string) Errorer {
	fer.context = context
	return fer
}

// NewError creates a new foxy error with the given message.
func NewError(message string) Errorer {
	return &Error{message: message}
}

// NewErrorf creates a new foxy error with a formatted message.
func NewErrorf(message string, args ...any) Errorer {
	return NewError(fmt.Sprintf(message, args...))
}

// NewInactiveError returns an error indicating the DBF is not open.
func NewInactiveError() Errorer {
	return &Error{message: "dbf is inactive, could not perform operation"}
}

// NewNavigationError returns a general navigation error.
func NewNavigationError() Errorer {
	return &Error{message: "navigation error"}
}

// NewNavigationBofError returns an error indicating the beginning of the file has been reached.
func NewNavigationBofError() Errorer {
	return &Error{message: "bof"}
}

// NewNavigationEofError returns an error indicating the end of the file has been reached.
func NewNavigationEofError() Errorer {
	return &Error{message: "eof"}
}

// FieldError is an error associated with a specific field.
type FieldError struct {
	werr *Error
	fld  *Field
}

func (fe *FieldError) Error() string {

	return fe.werr.Error()
}

func (fe *FieldError) Unwrap() error {

	return fe.werr.Unwrap()
}

func (fe *FieldError) SetWrapped(err error) Errorer {

	return fe.werr.SetWrapped(err)
}

func (fe *FieldError) SetContext(context string) Errorer {

	return fe.werr.SetContext(context)
}

func NewFieldError(field *Field, message string) Errorer {
	return &FieldError{
		fld: field,
		werr: &Error{
			message: message,
		},
	}
}

func NewFieldErrorf(field *Field, message string, args ...any) Errorer {

	return NewFieldError(field, fmt.Sprintf(message, args...))
}
