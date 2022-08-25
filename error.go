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

// EppError represent the data needed for an EPP error.
type EppError struct {
	Code      int
	Message   string
	Values    []Value
	ExtValues []ExtValue
}

// Error implements the error interface.
func (err *EppError) Error() string {
	return fmt.Sprintf("%d: %s", err.Code, err.Message)
}

// NewError create a new Error.
func NewError(code int) *EppError {
	return &EppError{
		Code:    code,
		Message: StatusText(code),
	}
}

// WithExtValues add extvalue data to the error.
func (err *EppError) WithExtValues(extValue ...ExtValue) *EppError {
	err.ExtValues = append(err.ExtValues, extValue...)
	return err
}

// WithValues add value data to the error.
func (err *EppError) WithValues(value ...Value) *EppError {
	err.Values = append(err.Values, value...)
	return err
}
