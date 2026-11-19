package supplychain

import (
	"encoding/json"
	"io"
)

// ComponentTypeAttributeKind is the kind of the component type attribute written by the
// `component_type` rule of `@package_metadata//attributes:component_type.bzl`.
const ComponentTypeAttributeKind = "build.bazel.attribute.component_type"

// ComponentTypeAttributeDescriptor describes the component type attribute of a `PackageMetadata`.
var ComponentTypeAttributeDescriptor = PackageAttributeDescriptor[string]{
	Kind: ComponentTypeAttributeKind,
	Parser: func(r io.Reader) (*string, error) {
		var t string
		if err := json.NewDecoder(r).Decode(&t); err != nil {
			return nil, err
		}
		return &t, nil
	},
}

// GetComponentTypeAttribute returns the component type of the `PackageMetadata`
// such as "container", or "" if it does not declare one.
func GetComponentTypeAttribute(m PackageMetadata) (string, error) {
	for _, kind := range m.ListAttributeKinds() {
		if kind == ComponentTypeAttributeKind {
			t, err := GetPackageAttribute(m, ComponentTypeAttributeDescriptor)
			if err != nil {
				return "", err
			}
			return *t, nil
		}
	}
	return "", nil
}
