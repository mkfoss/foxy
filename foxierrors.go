package foxy

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
