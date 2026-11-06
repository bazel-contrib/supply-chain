package supplychain

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// LicenseAttributeKind is the kind of the license attribute written by the
// `license` rule of `@package_metadata//licenses/rules:license.bzl`.
const LicenseAttributeKind = "build.bazel.attribute.license"

// LicenseAttributeDescriptor describes the license attribute of a `PackageMetadata`.
var LicenseAttributeDescriptor = PackageAttributeDescriptor[LicenseAttribute]{
	Kind: LicenseAttributeKind,
	Parser: func(r io.Reader) (*LicenseAttribute, error) {
		var raw struct {
			LicenseAttribute
			// Kind is the deprecated representation of a single license,
			// written by versions of the rule that predate `expression`.
			Kind *LicenseKind `json:"kind"`
		}
		if err := json.NewDecoder(r).Decode(&raw); err != nil {
			return nil, err
		}
		a := raw.LicenseAttribute
		switch {
		case a.Expression != nil:
		case raw.Kind != nil:
			a.Expression = &LicenseExpression{Kind: raw.Kind}
		default:
			return nil, fmt.Errorf("license attribute must have an expression")
		}
		if err := a.Expression.validate(); err != nil {
			return nil, err
		}
		return &a, nil
	},
}

// GetLicenseAttribute returns the license attribute of the `PackageMetadata`,
// or nil if it does not declare one.
func GetLicenseAttribute(m PackageMetadata) (*LicenseAttribute, error) {
	for _, kind := range m.ListAttributeKinds() {
		if kind == LicenseAttributeKind {
			return GetPackageAttribute(m, LicenseAttributeDescriptor)
		}
	}
	return nil, nil
}

// LicenseKind identifies a license.
type LicenseKind struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	// Custom is set by the rule if the identifier is not on the SPDX list.
	Custom bool `json:"custom,omitempty"`
}

// IsCustom returns whether the kind is not part of the SPDX license list.
// Reports of rules that predate `custom` can only be told by the `LicenseRef-`
// prefix.
func (k LicenseKind) IsCustom() bool {
	return k.Custom || strings.HasPrefix(k.Identifier, "LicenseRef-")
}

// RefID returns the identifier to use in an SPDX expression: the identifier
// itself for SPDX kinds and for kinds already named `LicenseRef-`/`AdditionRef-`,
// otherwise the identifier with the reference prefix.
func (k LicenseKind) RefID(exception bool) string {
	if !k.IsCustom() || strings.HasPrefix(k.Identifier, "LicenseRef-") || strings.HasPrefix(k.Identifier, "AdditionRef-") {
		return k.Identifier
	}
	if exception {
		return "AdditionRef-" + k.Identifier
	}
	return "LicenseRef-" + k.Identifier
}

// LicenseExpression is a node of the license expression tree. For now, the
// only kind of node is a leaf, which has a Kind.
type LicenseExpression struct {
	Kind *LicenseKind `json:"kind,omitempty"`
}

func (e LicenseExpression) validate() error {
	if e.Kind == nil {
		return fmt.Errorf("invalid license expression: a node needs a license kind")
	}
	return nil
}

// LicenseAttribute is the license attribute of a package. Its Expression is
// the root of the license expression tree; a single license is a leaf node.
type LicenseAttribute struct {
	Expression *LicenseExpression `json:"expression"`
	Label      string             `json:"label"`
	Text       string             `json:"text,omitempty"`
}

// Root returns the root node of the license expression.
func (a *LicenseAttribute) Root() LicenseExpression {
	return *a.Expression
}

// String renders the expression as an SPDX license expression.
func (e LicenseExpression) String() string {
	return e.Kind.Identifier
}

// IsSingle returns whether the expression is a single license without any
// operator.
func (e LicenseExpression) IsSingle() bool {
	return e.Kind != nil
}
