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
		var a LicenseAttribute
		if err := json.NewDecoder(r).Decode(&a); err != nil {
			return nil, err
		}
		if (a.Kind == nil) == (a.Expression == nil) {
			return nil, fmt.Errorf("license attribute must have exactly one of kind and expression")
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

// LicenseKind identifies a license or a license exception.
type LicenseKind struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

// IsCustom returns whether the kind is not part of the SPDX license list.
func (k LicenseKind) IsCustom() bool {
	return strings.HasPrefix(k.Identifier, "LicenseRef-") || strings.HasPrefix(k.Identifier, "AdditionRef-")
}

// LicenseOperator is the operator of a compound node of a `LicenseExpression`.
type LicenseOperator string

const (
	LicenseAnd  LicenseOperator = "AND"
	LicenseOr   LicenseOperator = "OR"
	LicenseWith LicenseOperator = "WITH"
)

// LicenseExpression is a node of a parsed SPDX license expression.
//
// A leaf has a Kind. A compound node has an Operator and Operands; `WITH`
// nodes have exactly one operand (the licensed leaf) and an Exception.
type LicenseExpression struct {
	Kind      *LicenseKind        `json:"kind,omitempty"`
	Operator  LicenseOperator     `json:"operator,omitempty"`
	Operands  []LicenseExpression `json:"operands,omitempty"`
	Exception *LicenseKind        `json:"exception,omitempty"`
}

// LicenseAttribute is the license attribute of a package: either a single
// Kind, or an Expression.
type LicenseAttribute struct {
	Kind       *LicenseKind       `json:"kind,omitempty"`
	Expression *LicenseExpression `json:"expression,omitempty"`
	Label      string             `json:"label"`
	Text       string             `json:"text,omitempty"`
}

// Root returns the license of the attribute as an expression, regardless of
// whether a single kind or an expression was declared.
func (a *LicenseAttribute) Root() LicenseExpression {
	if a.Expression != nil {
		return *a.Expression
	}
	return LicenseExpression{Kind: a.Kind}
}

// String renders the expression as an SPDX license expression.
func (e LicenseExpression) String() string {
	switch {
	case e.Kind != nil:
		return e.Kind.Identifier
	case e.Operator == LicenseWith:
		return e.Operands[0].String() + " WITH " + e.Exception.Identifier
	default:
		parts := make([]string, len(e.Operands))
		for i, o := range e.Operands {
			parts[i] = o.String()
			if (o.Operator == LicenseAnd || o.Operator == LicenseOr) && o.Operator != e.Operator {
				parts[i] = "(" + parts[i] + ")"
			}
		}
		return strings.Join(parts, " "+string(e.Operator)+" ")
	}
}

// IsSingle returns whether the expression is a single license without any
// operator.
func (e LicenseExpression) IsSingle() bool {
	return e.Kind != nil
}

// Kinds returns all license kinds and exceptions referenced by the
// expression, in order of appearance and without duplicates.
func (e LicenseExpression) Kinds() []LicenseKind {
	var kinds []LicenseKind
	seen := map[string]bool{}
	var walk func(LicenseExpression)
	walk = func(n LicenseExpression) {
		if n.Kind != nil && !seen[n.Kind.Identifier] {
			seen[n.Kind.Identifier] = true
			kinds = append(kinds, *n.Kind)
		}
		for _, o := range n.Operands {
			walk(o)
		}
		if n.Exception != nil && !seen[n.Exception.Identifier] {
			seen[n.Exception.Identifier] = true
			kinds = append(kinds, *n.Exception)
		}
	}
	walk(e)
	return kinds
}
