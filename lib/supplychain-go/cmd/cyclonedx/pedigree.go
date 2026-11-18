package main

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
)

// pedigreeOf converts the patches of a package to a CycloneDX pedigree.
//
// CycloneDX has no description or checksum on a patch. The checksums are not
// reported; the content of a patch file is attached to its diff, and the
// descriptions go to the notes of the pedigree, the place for commentary about
// it, each prefixed by the name of its patch. The path of a patch file goes to
// the notes too.
func pedigreeOf(patches []*supplychain.PatchAttribute) (*cdx.Pedigree, error) {
	if len(patches) == 0 {
		return nil, nil
	}

	var notes []string
	cdxPatches := make([]cdx.Patch, 0, len(patches))
	for _, p := range patches {
		patch := cdx.Patch{Type: cdx.PatchType(p.Type)}
		if p.URL != "" {
			patch.Diff = &cdx.Diff{URL: p.URL}
		} else {
			content, err := os.ReadFile(p.FileExecPath)
			if err != nil {
				return nil, fmt.Errorf("reading patch file %s: %w", p.File, err)
			}
			// Base64 keeps a patch that is not valid UTF-8 intact.
			patch.Diff = &cdx.Diff{Text: &cdx.AttachedText{
				Content:     base64.StdEncoding.EncodeToString(content),
				ContentType: "text/x-patch",
				Encoding:    "base64",
			}}
		}
		if len(p.Resolves) > 0 {
			issues := make([]cdx.Issue, 0, len(p.Resolves))
			for _, id := range p.Resolves {
				issues = append(issues, issueOf(id))
			}
			patch.Resolves = &issues
		}
		cdxPatches = append(cdxPatches, patch)
		if note := strings.TrimSpace(p.Description + " " + p.FileComment()); note != "" {
			notes = append(notes, p.Name()+": "+note)
		}
	}
	return &cdx.Pedigree{
		Patches: &cdxPatches,
		Notes:   strings.Join(notes, "\n"),
	}, nil
}

// issueOf converts an identifier of a resolved issue to a CycloneDX issue.
// The kind of an issue is not declared, so vulnerability identifiers (CVE and
// GitHub advisory ids) are reported as security issues and anything else as a
// defect.
func issueOf(id string) cdx.Issue {
	issue := cdx.Issue{ID: id, Type: cdx.IssueTypeDefect}
	if upper := strings.ToUpper(id); strings.HasPrefix(upper, "CVE-") || strings.HasPrefix(upper, "GHSA-") {
		issue.Type = cdx.IssueTypeSecurity
	}
	if u, err := url.Parse(id); err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https") {
		issue.References = &[]string{id}
	}
	return issue
}
