package main

import (
	"os"
	"sort"
	"strings"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/spdx/tools-golang/spdx"
)

// licenseCollector gathers the `LicenseRef-` licenses used by packages, which
// SPDX requires to be defined in the document.
type licenseCollector struct {
	byID map[string]*spdx.OtherLicense
}

func newLicenseCollector() *licenseCollector {
	return &licenseCollector{byID: map[string]*spdx.OtherLicense{}}
}

// declared returns the SPDX license expression for the package and records
// the custom licenses it references.
//
// Custom exceptions (`AdditionRef-`) are only referenced from the expression:
// SPDX 2.3 has no document-level definition for them.
func (c *licenseCollector) declared(attribute *supplychain.LicenseAttribute) string {
	if attribute == nil {
		return "NOASSERTION"
	}

	root := attribute.Root()
	for _, kind := range root.Kinds() {
		if !isLicenseRef(kind) {
			continue
		}
		if _, ok := c.byID[kind.Identifier]; ok {
			continue
		}
		c.byID[kind.Identifier] = &spdx.OtherLicense{
			LicenseIdentifier: kind.Identifier,
			LicenseName:       kind.Name,
			ExtractedText:     extractedText(attribute, kind),
		}
	}
	return root.String()
}

func (c *licenseCollector) otherLicenses() []*spdx.OtherLicense {
	licenses := make([]*spdx.OtherLicense, 0, len(c.byID))
	for _, l := range c.byID {
		licenses = append(licenses, l)
	}
	sort.Slice(licenses, func(i, j int) bool {
		return licenses[i].LicenseIdentifier < licenses[j].LicenseIdentifier
	})
	return licenses
}

func isLicenseRef(kind supplychain.LicenseKind) bool {
	return strings.HasPrefix(kind.Identifier, "LicenseRef-")
}

// extractedText is the license text file of the package when the license
// consists of a single kind, otherwise the name of the license.
func extractedText(attribute *supplychain.LicenseAttribute, kind supplychain.LicenseKind) string {
	if attribute.Text != "" && attribute.Root().IsSingle() {
		if text, err := os.ReadFile(attribute.Text); err == nil {
			return string(text)
		}
	}
	return kind.Name
}
