package epplib

import (
	"fmt"
	"strings"
)

// XMLPathBuilder is a string with xml path functions.
type XMLPathBuilder string

// NewXMLPathBuilder get a new xml path builder.
func NewXMLPathBuilder() XMLPathBuilder {
	return ""
}

// AddOrphan add an orphaned (no parent - no "/" prefix) tag to the path.
func (b XMLPathBuilder) AddOrphan(tag, namespace string) XMLPathBuilder {
	if namespace != "" {
		tag = fmt.Sprintf("%s[namespace-uri()='%s']", tag, namespace)
	}

	return XMLPathBuilder(string(b) + tag)
}

// Add a tag to the path. If no "/" is present in the beginning of the tag it will be added.
func (b XMLPathBuilder) Add(tag, namespace string) XMLPathBuilder {
	if !strings.HasPrefix(tag, "/") {
		tag = fmt.Sprintf("/%s", tag)
	}

	if namespace != "" {
		tag = fmt.Sprintf("%s[namespace-uri()='%s']", tag, namespace)
	}

	return XMLPathBuilder(string(b) + tag)
}

// String get the path as a string.
func (b XMLPathBuilder) String() string {
	return string(b)
}
