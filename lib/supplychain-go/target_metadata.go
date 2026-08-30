package supplychain

import (
	"encoding/json"
	"io"
	"os"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/label"
)

// TargetMetadata provides metadata about a (configured) Bazel target.
//
// This represents a single node in the target graph. It can be used to recover
// the full graph of dependencies starting from one or multiple targets (e.g.,
// from the shippable artifact).
type TargetMetadata interface {
	// TargetMetadataPrivate acts as marker to prevent other packages to implement the interface.
	//
	// See also https://medium.com/@johnsiilver/writing-an-interface-that-only-sub-packages-can-implement-fe36e7511449
	targetMetadataPrivate()

	// GetLabel returns the [Label](https://bazel.build/rules/lib/builtins/Label) of the target this `TargetMetadata` is for.
	GetLabel() label.Label

	// GetPackageMetadata returns a slice of `PackageMetadata` directly attached to the target this `TargetMetadata` is for.
	GetPackageMetadata() ([]PackageMetadata, error)

	// GetDependencies returns a slice of direct dependencies of the target this `TargetMetadata` is for.
	//
	// This includes
	//   - tool dependencies like compilers that were used to produce the target (e.g., toolchain dependencies),
	//   - runtime dependencies that are linked into the target (e.g., dependencies from the `deps` attribute of the target),
	//   - dynamic dependencies that are not linked into the target but needed at runtime (e.g., system-wide shared libraries), and
	//   - bundled dependencies that are included in the target (e.g., for archives or containers that bundle other targets).
	GetDependencies() ([]Dependency, error)
}

// ReadTargetMetadata deserializes `TargetMetadata` from the provided reader.
func ReadTargetMetadata(r io.Reader) (TargetMetadata, error) {
	var metadata targetMetadata
	if err := json.NewDecoder(r).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// ReadTargetMetadataFromFile deserializes `TargetMetadata` from a file with the provided path.
func ReadTargetMetadataFromFile(path string) (TargetMetadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadTargetMetadata(f)
}

// WriteTargetMetadata serializes `TargetMetadata` to the provided reader.
func WriteTargetMetadata(r io.Writer, metadata TargetMetadata) error {
	return json.NewEncoder(r).Encode(metadata)
}

// WriteTargetMetadataToFile serializes `TargetMetadata` into a file with the provided path.
func WriteTargetMetadataToFile(path string, metadata TargetMetadata) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return WriteTargetMetadata(f, metadata)
}

type targetMetadata struct {
	Label           label.Label
	PackageMetadata []string
	Dependencies    []rawDependency
}

/*
 * TargetMetadata implementation
 */
var _ TargetMetadata = (*targetMetadata)(nil)

func (t *targetMetadata) targetMetadataPrivate() {
	// Nothing to do.
}

func (t *targetMetadata) GetLabel() label.Label {
	return t.Label
}

func (t *targetMetadata) GetPackageMetadata() ([]PackageMetadata, error) {
	packageMetadata := make([]PackageMetadata, 0, len(t.PackageMetadata))
	for _, path := range t.PackageMetadata {
		p, err := ReadPackageMetadataFromFile(path)
		if err != nil {
			return nil, err
		}

		packageMetadata = append(packageMetadata, p)
	}

	return packageMetadata, nil
}

func (t *targetMetadata) GetDependencies() ([]Dependency, error) {
	dependencies := make([]Dependency, len(t.Dependencies))
	for i, rawDependency := range t.Dependencies {
		targetMetadata, err := ReadTargetMetadataFromFile(rawDependency.Path)
		if err != nil {
			return nil, err
		}

		dependencies[i] = &dependency{
			Scope:          dependencyScope(rawDependency.Scope),
			TargetMetadata: targetMetadata,
		}
	}

	return dependencies, nil
}

/*
 * JSON implementation.
 */
var _ json.Marshaler = (*targetMetadata)(nil)
var _ json.Unmarshaler = (*targetMetadata)(nil)

type rawTargetMetadata struct {
	Label           string          `json:"label"`
	PackageMetadata []string        `json:"package_metadata"`
	Dependencies    []rawDependency `json:"dependencies"`
}

func (t *targetMetadata) UnmarshalJSON(data []byte) error {
	var rawMetadata rawTargetMetadata
	if err := json.Unmarshal(data, &rawMetadata); err != nil {
		return err
	}

	l, err := label.Parse(rawMetadata.Label)
	if err != nil {
		return err
	}
	t.Label = l
	t.PackageMetadata = rawMetadata.PackageMetadata
	t.Dependencies = rawMetadata.Dependencies

	return nil
}

func (p *targetMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(&rawTargetMetadata{
		Label:           p.Label.String(),
		PackageMetadata: p.PackageMetadata,
		Dependencies:    p.Dependencies,
	})
}
