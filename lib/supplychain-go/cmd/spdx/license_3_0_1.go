package main

import (
	"fmt"
	"os"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

// declaredLicenses creates the license elements of the packages and the
// hasDeclaredLicense relationships to them. Relationship ids start at
// firstRelationship.
//
// A license on the SPDX license list is a listed license. Any other license is
// a custom license that carries the text of its license file, if it has one. A custom license id can be declared
// with different texts (e.g. different copyright lines); each text gets its own
// element, and the ids after the first are suffixed to keep them unique.
func declaredLicenses(packages []*spdx.Package, licenseOf map[*spdx.Package]*supplychain.LicenseAttribute, firstRelationship int) ([]spdx.AnyElement, []*spdx.Relationship, error) {
	type key struct{ id, text string }
	var elements []spdx.AnyElement
	var relationships []*spdx.Relationship
	byKey := make(map[key]spdx.AnyLicense)
	variants := make(map[string]int)

	for _, pkg := range packages {
		attribute, ok := licenseOf[pkg]
		if !ok {
			continue
		}
		kind := attribute.Root().Kind

		k := key{id: kind.Identifier}
		var text string
		if kind.IsCustom() && attribute.Text != "" {
			content, err := os.ReadFile(attribute.Text)
			if err != nil {
				return nil, nil, fmt.Errorf("reading the license text of %s: %w", pkg.Name, err)
			}
			text = string(content)
		}
		k.text = text
		license, ok := byKey[k]
		if !ok {
			if kind.IsCustom() {
				id := kind.RefID(false)
				if n := variants[id]; n > 0 {
					id = fmt.Sprintf("%s-%d", id, n)
				}
				variants[kind.RefID(false)]++
				name := kind.Name
				if name == "" {
					name = kind.Identifier
				}
				license = &spdx.CustomLicense{ID: id, Name: name, Text: text}
			} else {
				license = &spdx.ListedLicense{ID: "https://spdx.org/licenses/" + kind.Identifier, Name: kind.Identifier}
			}
			byKey[k] = license
			elements = append(elements, license)
		}

		relationships = append(relationships, &spdx.Relationship{
			ID:   fmt.Sprintf("rel-%d", firstRelationship+len(relationships)),
			From: pkg,
			To:   spdx.ElementList{license},
			Type: spdx.RelationshipType_HasDeclaredLicense,
		})
	}
	return elements, relationships, nil
}
