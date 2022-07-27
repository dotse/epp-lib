package epplib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	b := NewXMLPathBuilder().
		Add("epp", "urn:ietf:params:xml:ns:epp-1.0").
		Add("command", "random:namespace").
		Add("check", "urn:ietf:params:xml:ns:contact-1.0")

	assert.Equal(t, "/epp[namespace-uri()='urn:ietf:params:xml:ns:epp-1.0']/command[namespace-uri()='random:namespace']/check[namespace-uri()='urn:ietf:params:xml:ns:contact-1.0']", string(b))

	b2 := NewXMLPathBuilder().AddOrphan("name", "urn:ietf:params:xml:ns:contact-1.0")

	assert.Equal(t, "name[namespace-uri()='urn:ietf:params:xml:ns:contact-1.0']", b2.String())

	b3 := NewXMLPathBuilder().
		Add("//command", "random:namespace").
		Add("check", "urn:ietf:params:xml:ns:contact-1.0")

	assert.Equal(t, "//command[namespace-uri()='random:namespace']/check[namespace-uri()='urn:ietf:params:xml:ns:contact-1.0']", string(b3))
}
