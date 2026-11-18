package supplychain

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"github.com/package-url/packageurl-go"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/label"
)

// PackageMetadata provides metadata about a Bazel package.
type PackageMetadata interface {
	// packageMetadataPrivate acts as marker to prevent other packages to implement the interface.
	//
	// See also https://medium.com/@johnsiilver/writing-an-interface-that-only-sub-packages-can-implement-fe36e7511449
	packageMetadataPrivate()

	// GetLabel returns the [Label](https://bazel.build/rules/lib/builtins/Label) of the `package_metadata` target declaring this `PackageMetadata`.
	GetLabel() label.Label

	// GetPURL returns the [package-url](https://github.com/package-url/purl-spec/blob/main/PURL-SPECIFICATION.rst) this `PackageMetadata` if for.
	GetPURL() packageurl.PackageURL

	// ListAttributeKinds returns a slice of kinds of `PackageAttributeDescriptor`s included in this `PackageMetadata`.
	ListAttributeKinds() []string
}

// ReadPackageMetadata deserializes `PackageMetadata` from the provided reader.
func ReadPackageMetadata(r io.Reader) (PackageMetadata, error) {
	var metadata packageMetadata
	if err := json.NewDecoder(r).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// ReadPackageMetadataFromFile deserializes `PackageMetadata` from a file with the provided path.
func ReadPackageMetadataFromFile(path string) (PackageMetadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadPackageMetadata(f)
}

// WritePackageMetadata serializes `PackageMetadata` to the provided reader.
func WritePackageMetadata(r io.Writer, metadata PackageMetadata) error {
	return json.NewEncoder(r).Encode(metadata)
}

// WritePackageMetadataToFile serializes `PackageMetadata` into a file with the provided path.
func WritePackageMetadataToFile(path string, metadata PackageMetadata) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return WritePackageMetadata(f, metadata)
}

// GetPackageAttribute returns the provided attribute from the `PackageMetadata`.
func GetPackageAttribute[T any](m PackageMetadata, d PackageAttributeDescriptor[T]) (*T, error) {
	p := m.(*packageMetadata)

	paths, ok := p.Attributes[d.Kind]
	if !ok {
		return nil, fmt.Errorf("package %q does not have attribute %q", p.PURL, d.Kind)
	}
	if len(paths) != 1 {
		return nil, fmt.Errorf("package %q has %d attributes of kind %q, use GetPackageAttributes", p.PURL, len(paths), d.Kind)
	}

	return parseAttributeFile(paths[0], d)
}

// GetPackageAttributes returns all attributes of the provided kind from the
// `PackageMetadata`, in declaration order. This is meant for attribute kinds
// that allow being declared multiple times.
func GetPackageAttributes[T any](m PackageMetadata, d PackageAttributeDescriptor[T]) ([]*T, error) {
	p := m.(*packageMetadata)

	paths, ok := p.Attributes[d.Kind]
	if !ok {
		return nil, fmt.Errorf("package %q does not have attribute %q", p.PURL, d.Kind)
	}

	result := make([]*T, 0, len(paths))
	for _, path := range paths {
		a, err := parseAttributeFile(path, d)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

func parseAttributeFile[T any](path string, d PackageAttributeDescriptor[T]) (*T, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return d.Parser(r)
}

type packageMetadata struct {
	Label      label.Label
	PURL       packageurl.PackageURL
	// Attributes maps the kind to the paths of its attribute files. Kinds
	// that do not allow multiple attributes have exactly one path.
	Attributes map[string][]string
}

/*
 * PackageMetadata implementation
 */
var _ PackageMetadata = (*packageMetadata)(nil)

func (p *packageMetadata) packageMetadataPrivate() {
	// Nothing to do.
}

func (p *packageMetadata) GetLabel() label.Label {
	return p.Label
}

func (p *packageMetadata) GetPURL() packageurl.PackageURL {
	return p.PURL
}

func (p *packageMetadata) ListAttributeKinds() []string {
	return slices.Sorted(maps.Keys(p.Attributes))
}

/*
 * JSON implementation.
 */
var _ json.Marshaler = (*packageMetadata)(nil)
var _ json.Unmarshaler = (*packageMetadata)(nil)

type rawPackageMetadata struct {
	Label string `json:"label"`
	PURL  string `json:"purl"`
	// Attributes values are either a single path or, for kinds allowing
	// multiple attributes, an array of paths.
	Attributes map[string]json.RawMessage `json:"attributes"`
}

func (p *packageMetadata) UnmarshalJSON(data []byte) error {
	var rawMetadata rawPackageMetadata
	if err := json.Unmarshal(data, &rawMetadata); err != nil {
		return err
	}

	l, err := label.Parse(rawMetadata.Label)
	if err != nil {
		return err
	}
	p.Label = l

	purl, err := packageurl.FromString(rawMetadata.PURL)
	if err != nil {
		return err
	}
	p.PURL = purl

	p.Attributes = make(map[string][]string, len(rawMetadata.Attributes))
	for kind, raw := range rawMetadata.Attributes {
		var single string
		if err := json.Unmarshal(raw, &single); err == nil {
			p.Attributes[kind] = []string{single}
			continue
		}
		var multiple []string
		if err := json.Unmarshal(raw, &multiple); err != nil {
			return fmt.Errorf("attribute %q must be a string or an array of strings: %w", kind, err)
		}
		p.Attributes[kind] = multiple
	}

	return nil
}

func (p *packageMetadata) MarshalJSON() ([]byte, error) {
	attributes := make(map[string]json.RawMessage, len(p.Attributes))
	for kind, paths := range p.Attributes {
		var v any = paths
		if len(paths) == 1 {
			v = paths[0]
		}
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		attributes[kind] = raw
	}

	return json.Marshal(&rawPackageMetadata{
		Label:      p.Label.String(),
		PURL:       p.PURL.String(),
		Attributes: attributes,
	})
}
