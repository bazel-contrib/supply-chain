package supplychain

import (
	"crypto/sha256"
	"encoding/hex"
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
			// only used when there is no expression.
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
		for _, kind := range a.Expression.Kinds() {
			if err := kind.validate(); err != nil {
				return nil, err
			}
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
	// Custom is set by the rule if the identifier is not on the SPDX list.
	Custom bool `json:"custom,omitempty"`
}

// IsCustom returns whether the kind is not part of the SPDX license list.
func (k LicenseKind) IsCustom() bool {
	return k.Custom || strings.HasPrefix(k.Identifier, "LicenseRef-") || strings.HasPrefix(k.Identifier, "AdditionRef-")
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

// ShortHash returns a short, stable hash of the parts, used to make
// synthesized references unique.
func ShortHash(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:4])
}

// validate checks that custom identifiers only use the characters of an SPDX
// idstring (letters, digits, `.` and `-`) after their optional prefix.
func (k LicenseKind) validate() error {
	if !k.IsCustom() {
		return nil
	}
	id := strings.TrimPrefix(strings.TrimPrefix(k.Identifier, "LicenseRef-"), "AdditionRef-")
	valid := id != ""
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-':
		default:
			valid = false
		}
	}
	if !valid {
		return fmt.Errorf("invalid custom license identifier %q: only letters, digits, '.' and '-' are allowed", k.Identifier)
	}
	return nil
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
	// OrLater is set on a leaf written as `<id>+`.
	OrLater bool `json:"or_later,omitempty"`
}

// LicenseAttribute is the license attribute of a package. Its Expression is
// the root of the license AST; a single license is a leaf node.
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
	switch {
	case e.Kind != nil:
		return e.LeafID()
	case e.Operator == LicenseWith:
		return e.Operands[0].String() + " WITH " + e.Exception.RefID(true)
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

// LeafID returns the SPDX identifier of a leaf. SPDX only allows `+` after a
// license of the SPDX list, so the "or later" terms of a custom license get a
// reference of their own.
func (e LicenseExpression) LeafID() string {
	id := e.Kind.RefID(false)
	switch {
	case !e.OrLater:
		return id
	case e.Kind.IsCustom():
		return id + "-or-later"
	}
	return id + "+"
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
