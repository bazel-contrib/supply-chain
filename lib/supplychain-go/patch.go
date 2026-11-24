package supplychain

import (
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
		// Fail on a malformed digest here, so that every format rejects it.
		if _, err := a.Checksums(); err != nil {
			return nil, err
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

// WarnIfNoSupportedIntegrity writes a warning to w if none of the integrity
// entries of the patch uses a supported algorithm, so that the checksum missing
// from a report does not go unnoticed. owner names what the patch applies to.
// A malformed entry is not reported here: parsing the attribute already fails
// on it.
func (p *PatchAttribute) WarnIfNoSupportedIntegrity(w io.Writer, owner string) {
	checksums, err := p.Checksums()
	if err != nil || len(checksums) > 0 {
		return
	}
	algorithms := make([]string, 0, len(checksumSizes))
	for algorithm := range checksumSizes {
		algorithms = append(algorithms, algorithm)
	}
	slices.Sort(algorithms)
	fmt.Fprintf(w, "warning: none of the integrity entries of patch %s of %s uses a supported algorithm (%s), so no checksum is reported for it\n",
		p.Name(), owner, strings.Join(algorithms, ", "))
}

// Checksums returns the checksums of the patch from its integrity. See
// ParseChecksums.
func (p *PatchAttribute) Checksums() ([]Checksum, error) {
	return ParseChecksums(p.Integrity)
}
