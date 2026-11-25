package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

func main() {
	var outPath, graphPath, classificationsPath, format, buildStatusPath, serialNumberPath, validatorPath string
	var embedPatchFiles bool
	flag.StringVar(&outPath, "out", "", "The path to write the generated CycloneDX SBOM.")
	flag.StringVar(&graphPath, "graph", "", "The path to the graph JSON file.")
	flag.StringVar(&classificationsPath, "classifications", "", "The path to the classifications JSON file.")
	flag.StringVar(&format, "format", "json", "The output format of the CycloneDX SBOM (json or xml).")
	flag.StringVar(&buildStatusPath, "created_from_status_file", "", "Path to a Bazel volatile-status.txt file to read a BUILD_TIMESTAMP from for metadata.timestamp. If unset, or the file has no BUILD_TIMESTAMP key, a fixed deterministic timestamp (the Unix epoch) is used instead, matching Bazel's own --nostamp default.")
	flag.StringVar(&serialNumberPath, "serial_number_file", "", "Optional path to a file holding the BOM serialNumber, in urn:uuid form. It should be unique to each generated BOM. If unset, the BOM has no serialNumber.")
	flag.StringVar(&validatorPath, "validator", "", "Optional runfiles path or executable path to the CycloneDX CLI validator.")
	flag.BoolVar(&embedPatchFiles, "embed_patch_files", true, "Embed the content of the patch files in the pedigree of the packages they patch. Disable it to keep local patches out of the report.")
	flag.Parse()

	if outPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --out flag is required")
		os.Exit(1)
	}

	if graphPath == "" || classificationsPath == "" {
		fmt.Fprintln(os.Stderr, "Error: both --graph and --classifications flags are required")
		os.Exit(1)
	}

	graphBytes, err := os.ReadFile(graphPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading graph: %v\n", err)
		os.Exit(1)
	}

	var graph sbom.GraphConfig
	if err := json.Unmarshal(graphBytes, &graph); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing graph: %v\n", err)
		os.Exit(1)
	}

	classBytes, err := os.ReadFile(classificationsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading classifications: %v\n", err)
		os.Exit(1)
	}

	var classifications sbom.Classifications
	if err := json.Unmarshal(classBytes, &classifications); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing classifications: %v\n", err)
		os.Exit(1)
	}

	var opts []Option
	if !embedPatchFiles {
		opts = append(opts, WithoutPatchFiles())
	}
	bom, err := GenerateBOM(graph, classifications, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating BOM: %v\n", err)
		os.Exit(1)
	}

	created, err := sbom.ReadBuildTimestamp(buildStatusPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading build status file: %v\n", err)
		os.Exit(1)
	}
	if bom.Metadata == nil {
		bom.Metadata = &cdx.Metadata{}
	}
	bom.Metadata.Timestamp = created.Format(time.RFC3339)

	serialNumber, err := readSerialNumber(serialNumberPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading serial number: %v\n", err)
		os.Exit(1)
	}
	if serialNumber != "" {
		bom.SerialNumber = serialNumber
	}

	var buf bytes.Buffer
	encoder := cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatJSON)
	switch format {
	case "json":
		encoder = cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatJSON)
	case "xml":
		encoder = cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatXML)
	default:
		fmt.Fprintf(os.Stderr, "Error: '%s' is not a supported format. Use 'json' or 'xml'\n", format)
		os.Exit(1)
	}

	// cyclonedx-go defaults to a newer spec than the validator supports.
	if err := encoder.EncodeVersion(bom, cdx.SpecVersion1_6); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding BOM: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	if validatorPath != "" {
		if err := sbom.RunValidator(validatorPath, "validate", "--input-file", outPath, "--input-format", format, "--input-version", "v1_6", "--fail-on-errors"); err != nil {
			fmt.Fprintf(os.Stderr, "Error validating BOM: %v\n", err)
			os.Exit(1)
		}
	}
}

// Option customizes GenerateBOM.
type Option func(*options)

type options struct {
	omitPatchFiles bool
}

// WithoutPatchFiles leaves the content of patch files out of the BOM. The
// patches are still reported, with their type and the issues they resolve, and
// the path of a patch file stays in the notes of the pedigree.
func WithoutPatchFiles() Option {
	return func(o *options) { o.omitPatchFiles = true }
}

func GenerateBOM(graph sbom.GraphConfig, classifications sbom.Classifications, opts ...Option) (*cdx.BOM, error) {
	var config options
	for _, opt := range opts {
		opt(&config)
	}
	graph = sbom.NormalizeGraph(graph)
	components := make([]cdx.Component, 0)
	labelToBOMRef := make(map[string]string)
	var rootComponent *cdx.Component

	// Helper function to create component from node
	createComponent := func(node *sbom.NodeConfig) (cdx.Component, error) {
		if node.MetadataFile == "" {
			return cdx.Component{}, fmt.Errorf("node %s has no metadata file", node.Label)
		}

		pkgMetadata, err := supplychain.ReadPackageMetadataFromFile(node.MetadataFile)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading metadata file %s: %w", node.MetadataFile, err)
		}

		purl := pkgMetadata.GetPURL()

		fullName := purl.Name
		if purl.Namespace != "" {
			fullName = purl.Namespace + "/" + fullName
		}

		bomRef := purl.String()
		component := cdx.Component{
			BOMRef:     bomRef,
			Type:       cdx.ComponentTypeLibrary,
			Name:       fullName,
			PackageURL: bomRef,
		}

		downloadLocation := ""
		qualifiers := purl.Qualifiers.Map()
		if qualifier := strings.TrimSpace(qualifiers["download_url"]); qualifier != "" {
			downloadLocation = qualifier
		} else if qualifier := strings.TrimSpace(qualifiers["url_download"]); qualifier != "" {
			downloadLocation = qualifier
		}
		if downloadLocation != "" {
			component.ExternalReferences = &[]cdx.ExternalReference{
				{
					Type: cdx.ERTypeDistribution,
					URL:  downloadLocation,
				},
			}
		}

		// Add version if available
		if purl.Version != "" {
			component.Version = purl.Version
		}

		licenseAttribute, err := supplychain.GetLicenseAttribute(pkgMetadata)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading license of %s: %w", node.Label, err)
		}
		component.Licenses = licensesOf(licenseAttribute)

		integrity, err := supplychain.GetIntegrityAttribute(pkgMetadata)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading integrity of %s: %w", node.Label, err)
		}
		if integrity != nil {
			hashes, err := hashesOf(*integrity)
			if err != nil {
				return cdx.Component{}, fmt.Errorf("error reading integrity of %s: %w", node.Label, err)
			}
			component.Hashes = hashes
		}

		cpeAttribute, err := supplychain.GetCPEAttribute(pkgMetadata)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading CPE of %s: %w", node.Label, err)
		}
		if cpeAttribute != nil && len(*cpeAttribute) > 0 {
			// A CycloneDX component has room for a single CPE. It is the first
			// one, and the others are kept as properties.
			component.CPE = (*cpeAttribute)[0]
			for _, cpe := range (*cpeAttribute)[1:] {
				properties := append(valueOrEmpty(component.Properties), cdx.Property{Name: "bazel:cpe", Value: cpe})
				component.Properties = &properties
			}
		}

		patches, err := supplychain.GetPatchAttributes(pkgMetadata)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading patches of %s: %w", node.Label, err)
		}
		pedigree, err := pedigreeOf(patches, !config.omitPatchFiles)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reporting patches of %s: %w", node.Label, err)
		}
		component.Pedigree = pedigree

		labelToBOMRef[node.Label] = bomRef
		return component, nil
	}

	// Handle root component
	if classifications.RootComponent != nil {
		comp, err := createComponent(classifications.RootComponent)
		if err != nil {
			return nil, err
		}
		rootComponent = &comp
	}

	// Add direct and transitive dependencies, skipping any component whose
	// BOMRef was already added (classification should already be unique per
	// component, but a duplicate BOMRef is invalid CycloneDX, so guard here
	// too).
	seenBOMRef := make(map[string]bool)
	addComponent := func(node *sbom.NodeConfig) error {
		comp, err := createComponent(node)
		if err != nil {
			return err
		}
		if seenBOMRef[comp.BOMRef] {
			return nil
		}
		seenBOMRef[comp.BOMRef] = true
		components = append(components, comp)
		return nil
	}

	for i := range classifications.Dependencies.Direct {
		if err := addComponent(&classifications.Dependencies.Direct[i]); err != nil {
			return nil, err
		}
	}

	for i := range classifications.Dependencies.Transitive {
		if err := addComponent(&classifications.Dependencies.Transitive[i]); err != nil {
			return nil, err
		}
	}

	// Build Dependencies from graph edges
	depMap := make(map[string][]string) // parent BOMRef -> []child BOMRefs
	seenDep := make(map[string]map[string]bool)
	for _, edge := range graph.Edges {
		fromRef, fromOk := labelToBOMRef[edge.From]
		toRef, toOk := labelToBOMRef[edge.To]

		if !fromOk || !toOk || fromRef == toRef {
			// A self-reference can appear if two distinct graph nodes
			// resolved to the same component (e.g. duplicate metadata).
			continue
		}
		if seenDep[fromRef] == nil {
			seenDep[fromRef] = make(map[string]bool)
		}
		if seenDep[fromRef][toRef] {
			continue
		}
		seenDep[fromRef][toRef] = true
		depMap[fromRef] = append(depMap[fromRef], toRef)
	}

	// CycloneDX treats a component that is absent from the dependency graph as
	// having unknown dependencies, and one with an empty dependsOn as having none.
	// Every component of the graph was resolved above, so declare the ones without
	// outgoing edges explicitly instead of leaving them out.
	declareLeaf := func(ref string) {
		if _, ok := depMap[ref]; !ok {
			depMap[ref] = []string{}
		}
	}
	if rootComponent != nil {
		declareLeaf(rootComponent.BOMRef)
	}
	for _, component := range components {
		declareLeaf(component.BOMRef)
	}

	// Convert to CycloneDX Dependencies format
	var dependencies *[]cdx.Dependency
	if len(depMap) > 0 {
		deps := make([]cdx.Dependency, 0, len(depMap))
		for parentRef, childRefs := range depMap {
			deps = append(deps, cdx.Dependency{
				Ref:          parentRef,
				Dependencies: &childRefs,
			})
		}
		// Map iteration order is random; keep the output reproducible.
		sort.Slice(deps, func(i, j int) bool { return deps[i].Ref < deps[j].Ref })
		dependencies = &deps
	}

	bom := cdx.NewBOM()
	bom.Version = 1

	if len(components) > 0 {
		bom.Components = &components
	}

	bom.Dependencies = dependencies

	// Add metadata with tool information and root component
	metadata := &cdx.Metadata{
		Tools: &cdx.ToolsChoice{
			Components: &[]cdx.Component{
				{
					Type: cdx.ComponentTypeApplication,
					Name: "Bazel Supply Chain Tools CycloneDX generator",
				},
			},
		},
	}

	// Set the root component in metadata if we have one
	if rootComponent != nil {
		metadata.Component = rootComponent
	}

	bom.Metadata = metadata

	return bom, nil
}

// licensesOf converts the license attribute of a package to CycloneDX: a
// license identified by its SPDX identifier, or by its name if it is custom.
// Anything else is an SPDX license expression.
func licensesOf(attribute *supplychain.LicenseAttribute) *cdx.Licenses {
	if attribute == nil {
		return nil
	}

	root := attribute.Root()
	if !root.IsSingle() || root.OrLater {
		return &cdx.Licenses{{Expression: root.String()}}
	}
	kind := root.Kind
	license := cdx.License{ID: kind.Identifier}
	if kind.IsCustom() {
		license = cdx.License{Name: kind.Name}
		if license.Name == "" {
			license.Name = kind.Identifier
		}
	}
	return &cdx.Licenses{{License: &license}}
}

var hashAlgorithms = map[string]cdx.HashAlgorithm{
	"sha256": cdx.HashAlgoSHA256,
	"sha384": cdx.HashAlgoSHA384,
	"sha512": cdx.HashAlgoSHA512,
}

// hashesOf converts integrity entries to CycloneDX hashes, or nil if none of
// the algorithms is supported.
func hashesOf(integrity supplychain.IntegrityAttribute) (*[]cdx.Hash, error) {
	checksums, err := integrity.Checksums()
	if err != nil {
		return nil, err
	}
	var hashes []cdx.Hash
	for _, checksum := range checksums {
		hashes = append(hashes, cdx.Hash{Algorithm: hashAlgorithms[checksum.Algorithm], Value: checksum.Hex})
	}
	if len(hashes) == 0 {
		return nil, nil
	}
	return &hashes, nil
}

// serialNumberPattern is the one of the CycloneDX schema, which only accepts
// lowercase hexadecimal digits, as RFC 4122 asks for.
var serialNumberPattern = regexp.MustCompile(`^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// readSerialNumber reads a CycloneDX BOM serialNumber from the file at path.
// An empty path returns an empty serial number, meaning "not provided".
func readSerialNumber(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	serialNumber := strings.TrimSpace(string(data))
	if !serialNumberPattern.MatchString(serialNumber) {
		return "", fmt.Errorf("%s must contain a serial number in lowercase urn:uuid form, got %q", path, serialNumber)
	}
	return serialNumber, nil
}

func valueOrEmpty[T any](list *[]T) []T {
	if list == nil {
		return nil
	}
	return *list
}
