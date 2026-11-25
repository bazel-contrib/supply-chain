package main

import (
	"fmt"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

// patchFiles creates a file of the purpose patch for every patch of the
// packages, and a patchedBy relationship from the package to it. Relationship
// ids start at firstRelationship.
//
// SPDX 3.0.1 has no equivalent of the type of a patch or of the issues it
// resolves, so they are not reported.
func patchFiles(packages []*spdx.Package, patchesOf map[*spdx.Package][]*supplychain.PatchAttribute, firstRelationship int) ([]*spdx.File, []*spdx.Relationship, error) {
	var files []*spdx.File
	var relationships []*spdx.Relationship
	for _, pkg := range packages {
		for _, patch := range patchesOf[pkg] {
			checksums, err := patch.Checksums()
			if err != nil {
				return nil, nil, fmt.Errorf("reading the checksums of patch %s of %s: %w", patch.Name(), pkg.Name, err)
			}

			patch.WarnIfNoSupportedIntegrity(warnings, pkg.Name)

			file := &spdx.File{
				ID:             fmt.Sprintf("patch-%d", len(files)),
				Name:           patch.Name(),
				Description:    patch.Description,
				Comment:        patch.FileComment(),
				PrimaryPurpose: spdx.SoftwarePurpose_Patch,
			}
			if patch.URL != "" {
				file.ExternalRefs = spdx.ExternalRefList{&spdx.ExternalRef{
					Type:     spdx.ExternalRefType_AltDownloadLocation,
					Locators: []string{patch.URL},
				}}
			}
			for _, checksum := range checksums {
				file.VerifiedUsing = append(file.VerifiedUsing, &spdx.Hash{
					Algorithm: hashAlgorithms[checksum.Algorithm],
					Value:     checksum.Hex,
				})
			}
			files = append(files, file)
			relationships = append(relationships, &spdx.Relationship{
				ID:   fmt.Sprintf("rel-%d", firstRelationship+len(relationships)),
				From: pkg,
				To:   spdx.ElementList{file},
				Type: spdx.RelationshipType_PatchedBy,
			})
		}
	}
	return files, relationships, nil
}
