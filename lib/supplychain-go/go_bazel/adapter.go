// Package go_bazel maps Go package metadata emitted by Gazelle to fetched
// source directories in a Bazel output base.
package go_bazel

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/coverage"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/license_resolver"
)

// BuildRequest converts graph metadata into the generic resolver input. It
// recognizes any external Bazel repository carrying a pkg:golang PURL; it is
// not tied to a particular organization's repository prefix.
func BuildRequest(graph coverage.GraphConfig, outputBase string) ([]license_resolver.PackageInput, error) {
	if outputBase == "" {
		return nil, fmt.Errorf("Bazel output base must be non-empty")
	}

	packages := make(map[string]license_resolver.PackageInput)
	for _, node := range graph.Nodes {
		if node.MetadataFile == "" {
			continue
		}
		metadata, err := supplychain.ReadPackageMetadataFromFile(node.MetadataFile)
		if err != nil {
			return nil, fmt.Errorf("read package metadata %q: %w", node.MetadataFile, err)
		}
		purl := metadata.GetPURL()
		if purl.Type != "golang" {
			continue
		}
		repository, ok := externalRepository(metadata.GetLabel().String())
		if !ok {
			continue
		}
		purlString := purl.String()
		packages[purlString] = license_resolver.PackageInput{
			PURL:                purlString,
			Root:                filepath.Join(outputBase, "external", repository),
			SourceMetadataLabel: metadata.GetLabel().String(),
		}
	}

	result := make([]license_resolver.PackageInput, 0, len(packages))
	for _, pkg := range packages {
		result = append(result, pkg)
	}
	// Map iteration is intentionally normalized before producing action input.
	sort.Slice(result, func(i, j int) bool { return result[i].PURL < result[j].PURL })
	return result, nil
}

func externalRepository(label string) (string, bool) {
	if strings.HasPrefix(label, "@@") {
		label = label[2:]
	} else if strings.HasPrefix(label, "@") {
		label = label[1:]
	} else {
		return "", false
	}
	repository, _, found := strings.Cut(label, "//")
	return repository, found && repository != ""
}
