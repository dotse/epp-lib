package epplib

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// XMLString is a string that will be XML encoded when used.
type XMLString string

func (x XMLString) String() string {
	var buf bytes.Buffer

	if err := xml.EscapeText(&buf, []byte(x)); err != nil {
		// It should be safe to panic here since this method only
		// returns an error if it fails to write which is extremely unlikely...
		panic(err)
	}

	return buf.String()
}

// ParseXMLBool parses an XML value according to the
// XML Schema Part 2: Datatypes 3.2.2 boolean specification.
// Note: does not properly handle the replace and collapse constraints.
func ParseXMLBool(value string) (bool, error) {
	value = strings.TrimSpace(value)

	switch value {
	case "0", "false":
		return false, nil
	case "1", "true":
		return true, nil
	default:
		return false, fmt.Errorf("invalid value: %s", value)
	}
}
