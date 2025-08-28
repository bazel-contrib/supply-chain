package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
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

	fmt.Println(config)
}
