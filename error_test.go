package epplib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := NewError(StatusActionPending)
	assert.Equal(t, "1001: Command completed successfully; action pending", err.Error())

	err = err.WithExtValues(ExtValue{
		Element:   "name",
		Value:     "test.se",
		Namespace: NamespaceIETFDomain10.String(),
		Reason:    "random error",
	}, ExtValue{
		Element:   "registrant",
		Value:     "ABC123",
		Namespace: NamespaceIETFDomain10.String(),
		Reason:    "not found",
	})

	assert.Len(t, err.ExtValues, 2)
	assert.Len(t, err.Values, 0)

	err = err.WithValues(Value{
		Element: "element",
		Value:   "value",
	})

	assert.Len(t, err.ExtValues, 2)
	assert.Len(t, err.Values, 1)
}
