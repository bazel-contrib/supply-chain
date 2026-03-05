package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

func main() {
	var outPath, configPath, format string
	flag.StringVar(&outPath, "out", "", "The path to write the generated CycloneDX SBOM.")
	flag.StringVar(&configPath, "config", "", "The path to the SBOM generation configuration file.")
	flag.StringVar(&format, "format", "json", "The output format of the CycloneDX SBOM (json or xml).")
	flag.Parse()

	if configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --config flag is required")
		os.Exit(1)
	}

	if outPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --out flag is required")
		os.Exit(1)
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		os.Exit(1)
	}

	var config sbom.GenConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config file: %v\n", err)
		os.Exit(1)
	}

	out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening output file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	bom, err := GenerateBOM(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating BOM: %v\n", err)
		os.Exit(1)
	}

	encoder := cdx.NewBOMEncoder(out, cdx.BOMFileFormatJSON)
	switch format {
	case "json":
		encoder = cdx.NewBOMEncoder(out, cdx.BOMFileFormatJSON)
	case "xml":
		encoder = cdx.NewBOMEncoder(out, cdx.BOMFileFormatXML)
	default:
		fmt.Fprintf(os.Stderr, "Error: '%s' is not a supported format. Use 'json' or 'xml'\n", format)
		os.Exit(1)
	}

	if err := encoder.Encode(bom); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding BOM: %v\n", err)
		os.Exit(1)
	}
}

func GenerateBOM(config sbom.GenConfig) (*cdx.BOM, error) {
	components := make([]cdx.Component, 0, len(config.Deps))

	for _, dep := range config.Deps {
		pkgMetadata, err := supplychain.ReadPackageMetadataFromFile(dep.Metadata)
		if err != nil {
			return nil, fmt.Errorf("error reading metadata file %s: %w", dep.Metadata, err)
		}

		purl := pkgMetadata.GetPURL()
		component := cdx.Component{
			BOMRef:     purl.String(),
			Type:       cdx.ComponentTypeLibrary,
			Name:       purl.Name,
			PackageURL: purl.String(),
		}

		// Add version if available
		if purl.Version != "" {
			component.Version = purl.Version
		}

		components = append(components, component)
	}

	bom := cdx.NewBOM()
	bom.Version = 1

	if len(components) > 0 {
		bom.Components = &components
	}

	// Add metadata with tool information
	bom.Metadata = &cdx.Metadata{
		Tools: &cdx.ToolsChoice{
			Components: &[]cdx.Component{
				{
					Type: cdx.ComponentTypeApplication,
					Name: "Bazel Supply Chain Tools CycloneDX generator",
				},
			},
		},
	}

	return bom, nil
}
