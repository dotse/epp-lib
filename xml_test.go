package epplib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	str := XMLString("&")
	assert.Equal(t, "&amp;", str.String())

	str = "hello"
	assert.Equal(t, "hello", str.String())
}

func TestParseBool(t *testing.T) {
	for _, tc := range []struct {
		name        string
		input       string
		expectError bool
		expectBool  bool
	}{
		{
			name:       "can parse 1",
			input:      "1",
			expectBool: true,
		},
		{
			name:       "can parse 0",
			input:      "0",
			expectBool: false,
		},
		{
			name:       "can parse true",
			input:      "true",
			expectBool: true,
		},
		{
			name:       "can parse false",
			input:      "false",
			expectBool: false,
		},
		{
			name:        "can't parse unknown",
			input:       "uknown",
			expectError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotBool, gotErr := ParseXMLBool(tc.input)

			require.Equal(t, tc.expectError, gotErr != nil)
			assert.Equal(t, tc.expectBool, gotBool)
		})
	}
}
