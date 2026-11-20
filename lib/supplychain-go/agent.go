package supplychain

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	PackageAgentTypeOrganization = "organization"
	PackageAgentTypePerson       = "person"
)

// PackageAgent describes the supplier or originator of a package.
type PackageAgent struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// SupplierAttributeKind is the kind of the supplier attribute.
const SupplierAttributeKind = "build.bazel.attribute.supplier"

// OriginatorAttributeKind is the kind of the originator attribute.
const OriginatorAttributeKind = "build.bazel.attribute.originator"

// SupplierAttributeDescriptor describes the supplier attribute of a package.
var SupplierAttributeDescriptor = PackageAttributeDescriptor[PackageAgent]{
	Kind:   SupplierAttributeKind,
	Parser: parsePackageAgent,
}

// OriginatorAttributeDescriptor describes the originator attribute of a package.
var OriginatorAttributeDescriptor = PackageAttributeDescriptor[PackageAgent]{
	Kind:   OriginatorAttributeKind,
	Parser: parsePackageAgent,
}

func parsePackageAgent(r io.Reader) (*PackageAgent, error) {
	var agent PackageAgent
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&agent); err != nil {
		return nil, err
	}
	if agent.Name == "" {
		return nil, fmt.Errorf("package agent name must not be empty")
	}
	if agent.Type != PackageAgentTypeOrganization && agent.Type != PackageAgentTypePerson {
		return nil, fmt.Errorf("invalid package agent type %q: want organization or person", agent.Type)
	}
	return &agent, nil
}

// GetSupplierAttribute returns the package supplier, or nil if it is not declared.
func GetSupplierAttribute(m PackageMetadata) (*PackageAgent, error) {
	for _, kind := range m.ListAttributeKinds() {
		if kind == SupplierAttributeKind {
			return GetPackageAttribute(m, SupplierAttributeDescriptor)
		}
	}
	return nil, nil
}

// GetOriginatorAttribute returns the package originator, or nil if it is not declared.
func GetOriginatorAttribute(m PackageMetadata) (*PackageAgent, error) {
	for _, kind := range m.ListAttributeKinds() {
		if kind == OriginatorAttributeKind {
			return GetPackageAttribute(m, OriginatorAttributeDescriptor)
		}
	}
	return nil, nil
}
