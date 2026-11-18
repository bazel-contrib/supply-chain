package supplychain

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

// IntegrityAttributeKind is the kind of the integrity attribute written by the
// `integrity` rule of `@package_metadata//attributes:integrity.bzl`.
const IntegrityAttributeKind = "build.bazel.attribute.integrity"

// Checksum is a digest of a package or of a file.
type Checksum struct {
	// Algorithm is one of "sha1", "sha256", "sha384" and "sha512".
	Algorithm string
	// Hex is the digest, hex-encoded.
	Hex string
}

var checksumSizes = map[string]int{
	"sha1":   20,
	"sha256": 32,
	"sha384": 48,
	"sha512": 64,
}

// ParseChecksums returns the checksums of Subresource Integrity (SRI) strings
// (`<algorithm>-<base64 digest>`). Algorithms that are not supported are
// dropped. A malformed entry is an error.
func ParseChecksums(sris []string) ([]Checksum, error) {
	var checksums []Checksum
	for _, sri := range sris {
		algorithm, digest, ok := strings.Cut(sri, "-")
		if !ok {
			return nil, fmt.Errorf("integrity %q is not of the form <algorithm>-<base64 digest>", sri)
		}
		algorithm = strings.ToLower(algorithm)
		size, ok := checksumSizes[algorithm]
		if !ok {
			continue
		}
		// The SRI spec (https://www.w3.org/TR/sri/, section 3.5) reserves
		// `?<options>` after the digest but defines no options, and consumers
		// must ignore the ones they do not recognize.
		digest, _, _ = strings.Cut(digest, "?")
		// Accept both the standard and the URL-safe base64 alphabets.
		digest = strings.NewReplacer("-", "+", "_", "/").Replace(strings.TrimRight(digest, "="))
		raw, err := base64.RawStdEncoding.DecodeString(digest)
		if err != nil {
			return nil, fmt.Errorf("integrity %q: %w", sri, err)
		}
		if len(raw) != size {
			return nil, fmt.Errorf("integrity %q: a %s digest has %d bytes, got %d", sri, algorithm, size, len(raw))
		}
		checksums = append(checksums, Checksum{Algorithm: algorithm, Hex: hex.EncodeToString(raw)})
	}
	return checksums, nil
}

// IntegrityAttribute holds the Subresource Integrity (SRI) strings of a package.
type IntegrityAttribute []string

// Checksums returns the checksums of the attribute. See ParseChecksums.
func (a IntegrityAttribute) Checksums() ([]Checksum, error) {
	return ParseChecksums(a)
}

// IntegrityAttributeDescriptor describes the integrity attribute of a
// `PackageMetadata`.
var IntegrityAttributeDescriptor = PackageAttributeDescriptor[IntegrityAttribute]{
	Kind: IntegrityAttributeKind,
	Parser: func(r io.Reader) (*IntegrityAttribute, error) {
		var a IntegrityAttribute
		if err := json.NewDecoder(r).Decode(&a); err != nil {
			return nil, err
		}
		return &a, nil
	},
}

// GetIntegrityAttribute returns the integrity attribute of the
// `PackageMetadata`, or nil if it does not declare one.
func GetIntegrityAttribute(m PackageMetadata) (*IntegrityAttribute, error) {
	if !slices.Contains(m.ListAttributeKinds(), IntegrityAttributeKind) {
		return nil, nil
	}
	return GetPackageAttribute(m, IntegrityAttributeDescriptor)
}
