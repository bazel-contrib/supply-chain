package supplychain

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"slices"
	"strings"
)

// PatchAttributeKind is the kind of the patch attribute written by the `patch`
// rule of `@package_metadata//attributes:patch.bzl`.
const PatchAttributeKind = "build.bazel.attribute.patch"

// PatchType is the nature of a patch. The values are the ones of the CycloneDX
// `pedigree.patches[].type`.
type PatchType string

const (
	PatchUnofficial PatchType = "unofficial"
	PatchMonkey     PatchType = "monkey"
	PatchBackport   PatchType = "backport"
	PatchCherryPick PatchType = "cherry-pick"
)

// PatchAttribute is a patch applied to a package. A package can declare
// several of them.
type PatchAttribute struct {
	// Description tells what the patch does.
	Description string `json:"description"`
	// Integrity holds the checksums of the patch as Subresource Integrity (SRI)
	// strings (`<algorithm>-<base64 digest>`). Use Checksums to read them.
	Integrity []string `json:"integrity"`
	// Type is the nature of the patch.
	Type PatchType `json:"type"`
	// URL is where the patch can be retrieved from. Exactly one of URL and
	// File is set.
	URL string `json:"url,omitempty"`
	// File is the path of the patch file, relative to the root of the repository
	// that holds it. Exactly one of URL and File is set.
	File string `json:"file,omitempty"`
	// FileExecPath is where to read the content of File from. It is set if and
	// only if File is.
	FileExecPath string `json:"file_exec_path,omitempty"`
	// Resolves holds identifiers (e.g., CVE ids, issue URLs) of the issues the
	// patch resolves.
	Resolves []string `json:"resolves,omitempty"`
}

// PatchAttributeDescriptor describes the patch attribute of a `PackageMetadata`.
var PatchAttributeDescriptor = PackageAttributeDescriptor[PatchAttribute]{
	Kind: PatchAttributeKind,
	Parser: func(r io.Reader) (*PatchAttribute, error) {
		var a PatchAttribute
		if err := json.NewDecoder(r).Decode(&a); err != nil {
			return nil, err
		}
		if len(a.Integrity) == 0 {
			return nil, fmt.Errorf("patch attribute must have an integrity")
		}
		if !slices.Contains([]PatchType{PatchUnofficial, PatchMonkey, PatchBackport, PatchCherryPick}, a.Type) {
			return nil, fmt.Errorf("patch attribute has an unknown type %q", a.Type)
		}
		if (a.URL == "") == (a.File == "") {
			return nil, fmt.Errorf("patch attribute must have exactly one of url and file")
		}
		if (a.File == "") != (a.FileExecPath == "") {
			return nil, fmt.Errorf("patch attribute must have file_exec_path if and only if it has file")
		}
		return &a, nil
	},
}

// GetPatchAttributes returns the patch attributes of the `PackageMetadata` in
// declaration order, or nil if it does not declare any.
func GetPatchAttributes(m PackageMetadata) ([]*PatchAttribute, error) {
	if !slices.Contains(m.ListAttributeKinds(), PatchAttributeKind) {
		return nil, nil
	}
	return GetPackageAttributes(m, PatchAttributeDescriptor)
}

// FileComment describes where the patch file is, or returns "" for a patch that
// is not a file. It is for tools whose format has no field for the path.
func (p *PatchAttribute) FileComment() string {
	if p.File == "" {
		return ""
	}
	return "Patch file " + p.File + " in the repository that declares the package."
}

// Name returns a name for the patch: the last segment of its URL or file path.
func (p *PatchAttribute) Name() string {
	name := ""
	if p.URL != "" {
		if u, err := url.Parse(p.URL); err == nil {
			name = path.Base(u.Path)
		}
	} else {
		name = path.Base(p.File)
	}
	if name == "." || name == "/" {
		return "patch"
	}
	return name
}

// Checksum is the digest of a patch.
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

// Checksums returns the checksums of the patch from its integrity. Algorithms
// that are not supported are dropped. A malformed entry is an error.
func (p *PatchAttribute) Checksums() ([]Checksum, error) {
	var checksums []Checksum
	for _, sri := range p.Integrity {
		algorithm, digest, ok := strings.Cut(sri, "-")
		if !ok {
			return nil, fmt.Errorf("integrity %q is not of the form <algorithm>-<base64 digest>", sri)
		}
		size, ok := checksumSizes[strings.ToLower(algorithm)]
		if !ok {
			continue
		}
		// An SRI string can carry options after the digest.
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
		checksums = append(checksums, Checksum{
			Algorithm: strings.ToLower(algorithm),
			Hex:       hex.EncodeToString(raw),
		})
	}
	return checksums, nil
}
