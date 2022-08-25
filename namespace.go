package epplib

type (
	// Namespace is the enum type for namespaces.
	Namespace int
	// Namespaces is a list of namespaces.
	Namespaces []Namespace
)

// Supported namespaces.
const (
	NamespaceUnknown Namespace = iota
	NamespaceIETFEPP10
	NamespaceW3XSI
	NamespaceIETFHost10
	NamespaceIETFContact10
	NamespaceIETFDomain10
	NamespaceIETFSecDNS10
	NamespaceIETFSecDNS11
	NamespaceIISEpp12
	NamespaceIISRegistryLock10
)

var (
	stringToNamespaceMap = map[string]Namespace{
		"urn:ietf:params:xml:ns:epp-1.0":            NamespaceIETFEPP10,
		"http://www.w3.org/2001/XMLSchema-instance": NamespaceW3XSI,
		"urn:ietf:params:xml:ns:host-1.0":           NamespaceIETFHost10,
		"urn:ietf:params:xml:ns:contact-1.0":        NamespaceIETFContact10,
		"urn:ietf:params:xml:ns:domain-1.0":         NamespaceIETFDomain10,
		"urn:ietf:params:xml:ns:secDNS-1.0":         NamespaceIETFSecDNS10,
		"urn:ietf:params:xml:ns:secDNS-1.1":         NamespaceIETFSecDNS11,
		"urn:se:iis:xml:epp:iis-1.2":                NamespaceIISEpp12,
		"urn:se:iis:xml:epp:registryLock-1.0":       NamespaceIISRegistryLock10,
	}
	namespaceToStringMap = map[Namespace]string{
		NamespaceIETFEPP10:         "urn:ietf:params:xml:ns:epp-1.0",
		NamespaceW3XSI:             "http://www.w3.org/2001/XMLSchema-instance",
		NamespaceIETFHost10:        "urn:ietf:params:xml:ns:host-1.0",
		NamespaceIETFContact10:     "urn:ietf:params:xml:ns:contact-1.0",
		NamespaceIETFDomain10:      "urn:ietf:params:xml:ns:domain-1.0",
		NamespaceIETFSecDNS10:      "urn:ietf:params:xml:ns:secDNS-1.0",
		NamespaceIETFSecDNS11:      "urn:ietf:params:xml:ns:secDNS-1.1",
		NamespaceIISEpp12:          "urn:se:iis:xml:epp:iis-1.2",
		NamespaceIISRegistryLock10: "urn:se:iis:xml:epp:registryLock-1.0",
	}
	objectNamespaceMap = map[Namespace]struct{}{
		NamespaceIETFHost10:    {},
		NamespaceIETFContact10: {},
		NamespaceIETFDomain10:  {},
	}
	extensionNamespaceMap = map[Namespace]struct{}{
		NamespaceIETFSecDNS10:      {},
		NamespaceIETFSecDNS11:      {},
		NamespaceIISEpp12:          {},
		NamespaceIISRegistryLock10: {},
	}
)

// NamespaceFromString return a namespace from a given string.
// If the namespace is not supported NamespaceUnknown will be returned.
func NamespaceFromString(ns string) Namespace {
	ret, ok := stringToNamespaceMap[ns]
	if !ok {
		return NamespaceUnknown
	}

	return ret
}

// String return the namespace as a string.
func (n Namespace) String() string {
	return namespaceToStringMap[n]
}

// IsObjectNamespace can be used to see if a given namespace is of the object type.
func (n Namespace) IsObjectNamespace() bool {
	_, ok := objectNamespaceMap[n]
	return ok
}

// IsExtensionNamespace can be used to see if a given namespace is of the extension type.
func (n Namespace) IsExtensionNamespace() bool {
	_, ok := extensionNamespaceMap[n]
	return ok
}

// HasNamespace check if a given namespace is in the list of namespaces.
func (ns Namespaces) HasNamespace(wantedNs Namespace) bool {
	for _, n := range ns {
		if n == wantedNs {
			return true
		}
	}

	return false
}
