package supplychain

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/label"
)

func TestReadValid(t *testing.T) {
	type ExpectedData struct {
		Label          string
		PURL           packageurl.PackageURL
		AttributeKinds []string
	}

	cases := []struct {
		Name     string
		Input    string
		Expected ExpectedData
	}{
		{
			Name: "no-version",
			Input: `
			{
				"purl": "pkg:github/bazel-contrib/supply-chain",
				"label": "@package_metadata//foo/bar",
				"attributes": {}
			}
			`,
			Expected: ExpectedData{
				Label: "@package_metadata//foo/bar",
				PURL: *packageurl.NewPackageURL(
					/* type= */ "github",
					/* namespace= */ "bazel-contrib",
					/* name= */ "supply-chain",
					/* version= */ "",
					/* qualifiers= */ packageurl.Qualifiers{},
					/* subpath= */ "",
				),
			},
		},
		{
			Name: "version",
			Input: `
			{
				"purl": "pkg:github/bazel-contrib/supply-chain@v0.0.1",
				"label": "@package_metadata//foo/bar",
				"attributes": {}
			}
			`,
			Expected: ExpectedData{
				Label: "@package_metadata//foo/bar",
				PURL: *packageurl.NewPackageURL(
					/* type= */ "github",
					/* namespace= */ "bazel-contrib",
					/* name= */ "supply-chain",
					/* version= */ "v0.0.1",
					/* qualifiers= */ packageurl.Qualifiers{},
					/* subpath= */ "",
				),
			},
		},
		{
			Name: "single-attribute-kind",
			Input: `
			{
				"purl": "pkg:github/bazel-contrib/supply-chain",
				"label": "@package_metadata//foo/bar",
				"attributes": {
				  "foo": "path/to/foo.attributes.json"
				}
			}
			`,
			Expected: ExpectedData{
				Label: "@package_metadata//foo/bar",
				PURL: *packageurl.NewPackageURL(
					/* type= */ "github",
					/* namespace= */ "bazel-contrib",
					/* name= */ "supply-chain",
					/* version= */ "",
					/* qualifiers= */ packageurl.Qualifiers{},
					/* subpath= */ "",
				),
				AttributeKinds: []string{
					"foo",
				},
			},
		},
		{
			Name: "multiple-attribute-kind",
			Input: `
			{
				"purl": "pkg:github/bazel-contrib/supply-chain",
				"label": "@package_metadata//foo/bar",
				"attributes": {
				  "bar": "path/to/bar.attributes.json",
				  "foo": "path/to/foo.attributes.json",
				  "baz": "path/to/foo.attributes.json"
				}
			}
			`,
			Expected: ExpectedData{
				Label: "@package_metadata//foo/bar",
				PURL: *packageurl.NewPackageURL(
					/* type= */ "github",
					/* namespace= */ "bazel-contrib",
					/* name= */ "supply-chain",
					/* version= */ "",
					/* qualifiers= */ packageurl.Qualifiers{},
					/* subpath= */ "",
				),
				AttributeKinds: []string{
					"bar",
					"baz",
					"foo",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(
			fmt.Sprintf("Deserialize %s", c.Name),
			func(t *testing.T) {
				m, err := ReadPackageMetadata(strings.NewReader(c.Input))
				assert.Nil(t, err)
				assert.Equal(t, label.MustParse(c.Expected.Label), m.GetLabel())
				assert.Equal(t, c.Expected.PURL, m.GetPURL())
				assert.ElementsMatch(t, c.Expected.AttributeKinds, m.ListAttributeKinds())
			})
		t.Run(
			fmt.Sprintf("Serialize %s", c.Name),
			func(t *testing.T) {
				m, err := ReadPackageMetadata(strings.NewReader(c.Input))
				assert.Nil(t, err)

				b, err := json.Marshal(m)
				assert.Nil(t, err)
				assert.JSONEq(t, c.Input, string(b))
			})
	}
}

type FakeAttribute struct {
	Name string
}

func TestGetAttribute(t *testing.T) {
	p := &packageMetadata{
		Label: label.MustParse("@package_metadata//foo/bar"),
		PURL: *packageurl.NewPackageURL(
			/* type= */ "github",
			/* namespace= */ "bazel-contrib",
			/* name= */ "supply-chain",
			/* version= */ "HEAD",
			/* qualifiers= */ packageurl.Qualifiers{},
			/* subpath= */ "",
		),
		Attributes: map[string]string{
			"fake":  os.Args[0],
			"other": "/does/not/exist",
		},
	}

	a, err := GetPackageAttribute(
		p,
		PackageAttributeDescriptor[FakeAttribute]{
			Kind: "fake",
			Parser: func(r io.Reader) (*FakeAttribute, error) {
				return &FakeAttribute{
					Name: "foo",
				}, nil
			},
		})
	assert.Nil(t, err)
	assert.Equal(t, &FakeAttribute{Name: "foo"}, a)
}
