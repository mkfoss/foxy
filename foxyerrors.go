package foxy

// todo: redesign error system, to have own Error interfaceand then separate error types that implement that interface,
// like in FieldError

import "fmt"

type Errorer interface {
	Error() string
	Unwrap() error
	SetWrapped(error) Errorer
	SetContext(context string) Errorer
}

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

func NewError(message string) Errorer {
	return &Error{message: message}
}

func NewErrorf(message string, args ...any) Errorer {
	return NewError(fmt.Sprintf(message, args...))
}

func NewInactiveError() Errorer {
	return &Error{message: "dbf is inactive, could not perform operation"}
}

func NewNavigationError() Errorer {
	return &Error{message: "navigation error"}
}

func NewNavigationBofError() Errorer {
	return &Error{message: "bof"}
}

func NewNavigationEofError() Errorer {
	return &Error{message: "eof"}
}

type FieldError struct {
	*Error
	fld *Field
}

func NewFieldError(field *Field, message string) Errorer {
	return &FieldError{
		fld: field,
		Error: &Error{
			message: message,
		},
	}
}

func NewFieldErrorf(field *Field, message string, args ...any) Errorer {
	return NewFieldError(field, fmt.Sprintf(message, args...))
}
