package supplychain

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/package-url/packageurl-go"
)

// PackageMetadata provides metadata about a Bazel package.
type PackageMetadata interface {
	// packageMetadataPrivate acts as marker to prevent other packages to implement the interface.
	//
	// See also https://medium.com/@johnsiilver/writing-an-interface-that-only-sub-packages-can-implement-fe36e7511449
	packageMetadataPrivate()

	// GetPURL returns the [package-url](https://github.com/package-url/purl-spec/blob/main/PURL-SPECIFICATION.rst) this `PackageMetadata` if for.
	GetPURL() packageurl.PackageURL
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

// GetPackageAttribute returns the provided attribute from the `PackageMetadata`.
func GetPackageAttribute[T any](m PackageMetadata, d PackageAttributeDescriptor[T]) (*T, error) {
	p := m.(*packageMetadata)

	a, ok := p.Attributes[d.Kind]
	if !ok {
		return nil, fmt.Errorf("package %q does not have attribute %q", p.PURL, d.Kind)
	}

	r, err := os.Open(a)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return d.Parser(r)
}

type packageMetadata struct {
	Label      string
	PURL       packageurl.PackageURL
	Attributes map[string]string
}

/*
 * PackageMetadata implementation
 */
var _ PackageMetadata = (*packageMetadata)(nil)

func (p *packageMetadata) packageMetadataPrivate() {
	// Nothing to do.
}

func (p *packageMetadata) GetPURL() packageurl.PackageURL {
	return p.PURL
}

/*
 * JSON implementation.
 */
var _ json.Marshaler = (*packageMetadata)(nil)
var _ json.Unmarshaler = (*packageMetadata)(nil)

type rawPackageMetadata struct {
	Label      string            `json:"label"`
	PURL       string            `json:"purl"`
	Attributes map[string]string `json:"attributes"`
}

func (p *packageMetadata) UnmarshalJSON(data []byte) error {
	var rawMetadata rawPackageMetadata
	if err := json.Unmarshal(data, &rawMetadata); err != nil {
		return err
	}

	p.Label = rawMetadata.Label

	purl, err := packageurl.FromString(rawMetadata.PURL)
	if err != nil {
		return err
	}
	p.PURL = purl

	p.Attributes = rawMetadata.Attributes
	if p.Attributes == nil {
		p.Attributes = make(map[string]string)
	}

	return nil
}

func (p *packageMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(&rawPackageMetadata{
		Label:      p.Label,
		PURL:       p.PURL.String(),
		Attributes: p.Attributes,
	})
}
