package main

import (
	"encoding/json"
	"flag"
	"os"

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

	doc := spdx.Document{
		SPDXVersion: "SPDX-3.0",
	}

	out, err := os.OpenFile(*outPath, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	spdxJson.Write(&doc, out)
}
