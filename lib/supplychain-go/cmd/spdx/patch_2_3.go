package main

import (
	"fmt"
	"strings"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	spdx2 "github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
)

// patchPackages23 creates a package for every patch of the packages, and a
// PATCH_APPLIED relationship from it to the package it patches.
//
// A patch is a package of the purpose FILE rather than a file: SPDX 2.3
// requires a SHA1 checksum on a file, which the integrity of a patch may not
// have. SPDX 2.3 has no equivalent of the type of a patch or of the issues it
// resolves, so they are not reported.
func patchPackages23(packages []*spdx2.Package, patchesOf map[*spdx2.Package][]*supplychain.PatchAttribute) ([]*spdx2.Package, []*spdx2.Relationship, error) {
	var patchPackages []*spdx2.Package
	var relationships []*spdx2.Relationship
	for _, pkg := range packages {
		for _, patch := range patchesOf[pkg] {
			checksums, err := patch.Checksums()
			if err != nil {
				return nil, nil, fmt.Errorf("reading the checksums of patch %s of %s: %w", patch.Name(), pkg.PackageName, err)
			}

			downloadLocation := "NOASSERTION"
			if patch.URL != "" {
				downloadLocation = patch.URL
			}
			patchPackage := &spdx2.Package{
				PackageSPDXIdentifier:   common.ElementID(fmt.Sprintf("patch-%d", len(patchPackages))),
				PackageName:             patch.Name(),
				PackageDescription:      patch.Description,
				PackageComment:          patch.FileComment(),
				PackageDownloadLocation: downloadLocation,
				PackageLicenseConcluded: "NOASSERTION",
				PackageLicenseDeclared:  "NOASSERTION",
				PrimaryPackagePurpose:   "FILE",
			}
			for _, checksum := range checksums {
				patchPackage.PackageChecksums = append(patchPackage.PackageChecksums, common.Checksum{
					Algorithm: common.ChecksumAlgorithm(strings.ToUpper(checksum.Algorithm)),
					Value:     checksum.Hex,
				})
			}
			patchPackages = append(patchPackages, patchPackage)
			relationships = append(relationships, &spdx2.Relationship{
				RefA:         common.MakeDocElementID("", string(patchPackage.PackageSPDXIdentifier)),
				RefB:         common.MakeDocElementID("", string(pkg.PackageSPDXIdentifier)),
				Relationship: "PATCH_APPLIED",
			})
		}
	}
	return patchPackages, relationships, nil
}
