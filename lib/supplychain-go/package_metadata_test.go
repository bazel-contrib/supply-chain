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
)

func TestReadValid(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected packageMetadata
	}{
		{
			Name: "simple",
			Input: `
			{
				"purl": "pkg:github/bazel-contrib/supply-chain@HEAD",
				"label": "@package_metadata//foo/bar",
				"attributes": {}
			}
			`,
			Expected: packageMetadata{
				Label: "@package_metadata//foo/bar",
				PURL: *packageurl.NewPackageURL(
					/* type= */ "github",
					/* namespace= */ "bazel-contrib",
					/* name= */ "supply-chain",
					/* version= */ "HEAD",
					/* qualifiers= */ packageurl.Qualifiers{},
					/* subpath= */ "",
				),
				Attributes: map[string]string{},
			},
		},
	}

	for _, c := range cases {
		t.Run(
			fmt.Sprintf("Deserialize %s", c.Name),
			func(t *testing.T) {
				m, err := ReadPackageMetadata(strings.NewReader(c.Input))
				assert.Nil(t, err)
				assert.Equal(t, &c.Expected, m)
			})
		t.Run(
			fmt.Sprintf("Serialize %s", c.Name),
			func(t *testing.T) {
				b, err := json.Marshal(&c.Expected)
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
		Label: "@package_metadata//foo/bar",
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
