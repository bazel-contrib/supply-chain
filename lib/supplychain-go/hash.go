package supplychain

import (
	"encoding/json"
	"io"
)

// Hash is a single content hash of a package: an algorithm name (CycloneDX
// `hash-alg` spelling, e.g. "SHA-256") and its hex digest.
type Hash struct {
	Alg     string `json:"alg"`
	Content string `json:"content"`
}

// HashAttribute describes the "build.bazel.attribute.hash" package attribute.
// Its payload is a JSON array of Hash objects.
var HashAttribute = PackageAttributeDescriptor[[]Hash]{
	Kind:   "build.bazel.attribute.hash",
	Parser: parseHashes,
}

func parseHashes(r io.Reader) (*[]Hash, error) {
	var hashes []Hash
	if err := json.NewDecoder(r).Decode(&hashes); err != nil {
		return nil, err
	}
	return &hashes, nil
}
