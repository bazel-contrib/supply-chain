package supplychain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseLicense(t *testing.T, s string) *LicenseAttribute {
	t.Helper()
	a, err := LicenseAttributeDescriptor.Parser(strings.NewReader(s))
	require.NoError(t, err)
	return a
}

func TestSingleLicenseIsLeafNode(t *testing.T) {
	root := parseLicense(t, `{"expression":{"kind":{"identifier":"MIT","name":"MIT License"}},"label":"//:l"}`).Root()
	assert.True(t, root.IsSingle())
	assert.Equal(t, "MIT", root.String())
}

func TestLicenseAttributeParserLegacyKind(t *testing.T) {
	a := parseLicense(t, `{"kind":{"identifier":"MIT","name":"MIT License"},"label":"//:l","text":"LICENSE"}`)
	assert.Equal(t, "MIT", a.Root().String())
	assert.Equal(t, "LICENSE", a.Text)
}

func TestLicenseAttributeRequiresExpression(t *testing.T) {
	_, err := LicenseAttributeDescriptor.Parser(strings.NewReader(`{"label":"//:l"}`))
	assert.Error(t, err)
}

func TestLicenseKindIsCustom(t *testing.T) {
	for _, tc := range []struct {
		json string
		want bool
	}{
		{`{"identifier":"MIT","name":"MIT License"}`, false},
		{`{"identifier":"acme","name":"ACME","custom":true}`, true},
		{`{"identifier":"LicenseRef-acme","name":"ACME"}`, true},
	} {
		root := parseLicense(t, `{"expression":{"kind":`+tc.json+`},"label":"//:l"}`).Root()
		assert.Equal(t, tc.want, root.Kind.IsCustom(), tc.json)
	}
}

func TestLicenseKindRefID(t *testing.T) {
	assert.Equal(t, "MIT", LicenseKind{Identifier: "MIT"}.RefID(false))
	assert.Equal(t, "LicenseRef-acme", LicenseKind{Identifier: "LicenseRef-acme"}.RefID(false))
	assert.Equal(t, "LicenseRef-acme", LicenseKind{Identifier: "acme", Custom: true}.RefID(false))
	assert.Equal(t, "AdditionRef-acme", LicenseKind{Identifier: "acme", Custom: true}.RefID(true))
}
