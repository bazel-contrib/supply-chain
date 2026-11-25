package supplychain

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
)

// CPEAttributeKind is the kind of the CPE attribute written by the `cpe` rule
// of `@package_metadata//attributes:cpe.bzl`.
const CPEAttributeKind = "build.bazel.attribute.cpe"

// CPEVersion is the version of the CPE specification a CPE name is written in.
type CPEVersion string

const (
	// CPE22 is the URI binding of CPE 2.2, e.g. `cpe:/o:canonical:ubuntu_linux:10.04:-:lts`.
	CPE22 CPEVersion = "2.2"
	// CPE23 is the formatted string binding of CPE 2.3, e.g.
	// `cpe:2.3:o:canonical:ubuntu_linux:10.04:-:lts:*:*:*:*:*`.
	CPE23 CPEVersion = "2.3"
)

// The locator formats of the external repository identifiers cpe22Type and
// cpe23Type, from Annex F of the SPDX 2.3 specification.
var (
	cpe22Pattern = regexp.MustCompile(`^[c][pP][eE]:/[AHOaho]?(:[A-Za-z0-9\._\-~%]*){0,6}$`)

	cpe23Component = `(?:(?:(?:\?*|\*?)(?:[a-zA-Z0-9\-\._]|(?:\\[\\\*\?!"#$%&'\(\)\+,/:;<=>@\[\]\^` + "`" + `\{\|}~]))+(?:\?*|\*?))|[\*\-])`
	cpe23Pattern   = regexp.MustCompile(`^cpe:2\.3:[aho\*\-]` +
		`(?::` + cpe23Component + `){5}` +
		`(?::(?:(?:[a-zA-Z]{2,3}(?:-(?:[a-zA-Z]{2}|[0-9]{3}))?)|[\*\-]))` +
		`(?::` + cpe23Component + `){4}$`)
)

// CPEVersionOf returns the version of the CPE specification the name is
// written in, or an error if it is not a CPE 2.2 URI or a CPE 2.3 formatted
// string.
func CPEVersionOf(cpe string) (CPEVersion, error) {
	switch {
	case cpe22Pattern.MatchString(cpe):
		return CPE22, nil
	case cpe23Pattern.MatchString(cpe):
		return CPE23, nil
	}
	return "", fmt.Errorf("%q is not a CPE 2.2 URI (cpe:/...) or a CPE 2.3 formatted string (cpe:2.3:...)", cpe)
}

// CPEAttribute holds the CPE names of a package, in declaration order.
type CPEAttribute []string

// CPEAttributeDescriptor describes the CPE attribute of a `PackageMetadata`.
var CPEAttributeDescriptor = PackageAttributeDescriptor[CPEAttribute]{
	Kind: CPEAttributeKind,
	Parser: func(r io.Reader) (*CPEAttribute, error) {
		var a CPEAttribute
		if err := json.NewDecoder(r).Decode(&a); err != nil {
			return nil, err
		}
		for _, cpe := range a {
			if _, err := CPEVersionOf(cpe); err != nil {
				return nil, fmt.Errorf("invalid cpe attribute: %w", err)
			}
		}
		return &a, nil
	},
}

// GetCPEAttribute returns the CPE attribute of the `PackageMetadata`, or nil
// if it does not declare one.
func GetCPEAttribute(m PackageMetadata) (*CPEAttribute, error) {
	if !slices.Contains(m.ListAttributeKinds(), CPEAttributeKind) {
		return nil, nil
	}
	return GetPackageAttribute(m, CPEAttributeDescriptor)
}
