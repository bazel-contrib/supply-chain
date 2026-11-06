package main

import (
	"fmt"
	"os"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	spdx "github.com/spdx/tools-golang/spdx"
)

// declaredLicenses23 sets the declared license of the packages and returns the
// licenses that are not on the SPDX list, which SPDX 2.3 defines in the
// document as extracted licensing information.
//
// A custom license id can be declared with different texts (e.g. different
// copyright lines); each text gets its own entry, and the ids after the first
// are suffixed to keep them unique.
func declaredLicenses23(packages []*spdx.Package, licenseOf map[*spdx.Package]*supplychain.LicenseAttribute) ([]*spdx.OtherLicense, error) {
	type key struct{ id, text string }
	var others []*spdx.OtherLicense
	idOf := make(map[key]string)
	variants := make(map[string]int)

	for _, pkg := range packages {
		attribute, ok := licenseOf[pkg]
		if !ok {
			continue
		}
		kind := attribute.Root().Kind
		if !kind.IsCustom() {
			pkg.PackageLicenseDeclared = kind.Identifier
			continue
		}

		text := "NOASSERTION"
		if attribute.Text != "" {
			content, err := os.ReadFile(attribute.Text)
			if err != nil {
				return nil, fmt.Errorf("reading the license text of %s: %w", pkg.PackageName, err)
			}
			text = string(content)
		}
		refID := kind.RefID(false)
		k := key{id: refID, text: text}
		id, ok := idOf[k]
		if !ok {
			id = refID
			if n := variants[refID]; n > 0 {
				id = fmt.Sprintf("%s-%d", refID, n)
			}
			variants[refID]++
			idOf[k] = id
			name := kind.Name
			if name == "" {
				name = kind.Identifier
			}
			others = append(others, &spdx.OtherLicense{LicenseIdentifier: id, ExtractedText: text, LicenseName: name})
		}
		pkg.PackageLicenseDeclared = id
	}
	return others, nil
}
