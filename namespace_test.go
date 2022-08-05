package epplib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespace(t *testing.T) {
	t.Parallel()

	for ns, s := range namespaceToStringMap {
		assert.Equal(t, ns, stringToNamespaceMap[s])
		assert.Equal(t, s, ns.String())
	}

	for s, ns := range stringToNamespaceMap {
		assert.Equal(t, s, namespaceToStringMap[ns])
		assert.Equal(t, ns, FromString(s))
	}

	assert.Equal(t, "", NamespaceUnknown.String())
	assert.Equal(t, NamespaceUnknown, FromString("unknown namespace"))
}

func TestNamespace_Types(t *testing.T) {
	for _, tc := range []struct {
		name          string
		namespaces    Namespaces
		isObjectNs    bool
		isExtensionNs bool
	}{
		{
			name: "recognize object namespaces",
			namespaces: Namespaces{
				NamespaceIETFHost10,
				NamespaceIETFContact10,
				NamespaceIETFDomain10,
			},
			isObjectNs: true,
		},
		{
			name: "recognize extension namespaces",
			namespaces: Namespaces{
				NamespaceIETFSecDNS10,
				NamespaceIETFSecDNS11,
				NamespaceIISEpp12,
				NamespaceIISRegistryLock10,
			},
			isExtensionNs: true,
		},
		{
			name: "recognize extensions with no special type",
			namespaces: Namespaces{
				NamespaceW3XSI,
				NamespaceIETFEPP10,
				NamespaceUnknown,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, ns := range tc.namespaces {
				assert.Equal(t, tc.isObjectNs, ns.IsObjectNamespace())
				assert.Equal(t, tc.isExtensionNs, ns.IsExtensionNamespace())
			}
		})
	}
}

func TestNamespaces(t *testing.T) {
	ns := Namespaces{NamespaceIETFHost10, NamespaceIETFDomain10}
	assert.True(t, ns.HasNamespace(NamespaceIETFDomain10))
	assert.False(t, ns.HasNamespace(NamespaceIETFContact10))
}
