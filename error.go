package epplib

import (
	"fmt"
)

// Value represent the value element in an EPP error.
type Value struct {
	Element string
	Value   string
}

// ExtValue represent the extvalue element in an EPP error.
type ExtValue struct {
	Element   string
	Value     string
	Namespace string
	Reason    string
}

// Error represent the data needed for an EPP error.
type Error struct {
	Code      int
	Message   string
	Values    []Value
	ExtValues []ExtValue
}

// Error implements the error interface.
func (err *Error) Error() string {
	return fmt.Sprintf("%d: %s", err.Code, err.Message)
}

// NewError create a new Error.
func NewError(code int) *Error {
	return &Error{
		Code:    code,
		Message: CodeText(code),
	}
}

// WithExtValues add extvalue data to the error.
func (err *Error) WithExtValues(extValue ...ExtValue) *Error {
	err.ExtValues = append(err.ExtValues, extValue...)
	return err
}

// WithValues add value data to the error.
func (err *Error) WithValues(value ...Value) *Error {
	err.Values = append(err.Values, value...)
	return err
}
