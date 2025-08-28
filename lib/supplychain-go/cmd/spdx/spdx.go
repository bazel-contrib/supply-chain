package main

import (
	"encoding/json"
	"flag"
	"os"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdxJson "github.com/spdx/tools-golang/json"
	"github.com/spdx/tools-golang/spdx"
)

var (
	outPath    = flag.String("out", "", "")
	configPath = flag.String("config", "", "")
)

func main() {
	flag.Parse()
	var config sbom.GenConfig

	configBytes, err := os.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(configBytes, &config)

	out, err := os.OpenFile(*outPath, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	doc, err := GenerateDocument(config)
	if err != nil {
		panic(err)
	}

	spdxJson.Write(doc, out)
}

func GenerateDocument(config sbom.GenConfig) (*spdx.Document, error) {
	spdxPackages := make([]*spdx.Package, len(config.Deps))

	for i, dep := range config.Deps {
		pkgMetadata, err := supplychain.ReadPackageMetadataFromFile(dep.Metadata)
		if err != nil {
			return nil, err
		}
		spdxPackages[i] = &spdx.Package{
			PackageName: pkgMetadata.GetPURL().Name,
		}
	}

	return &spdx.Document{
		SPDXVersion: "SPDX-3.0",
		Packages:    spdxPackages,
	}, nil
}
