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

func TestLicenseExpressionString(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "single kind",
			Input:    `{"expression":{"kind":{"identifier":"MIT","name":"MIT License"}},"label":"//:l"}`,
			Expected: "MIT",
		},
		{
			Name: "or",
			Input: `{"expression":{"operator":"OR","operands":[
				{"kind":{"identifier":"Apache-2.0","name":"a"}},
				{"kind":{"identifier":"MIT","name":"m"}}]},"label":"//:l"}`,
			Expected: "Apache-2.0 OR MIT",
		},
		{
			Name: "nested with exception",
			Input: `{"expression":{"operator":"AND","operands":[
				{"operator":"OR","operands":[
					{"kind":{"identifier":"MIT","name":"m"}},
					{"kind":{"identifier":"LicenseRef-acme","name":"c"}}]},
				{"operator":"WITH","operands":[{"kind":{"identifier":"GPL-2.0-only","name":"g"}}],
				 "exception":{"identifier":"Classpath-exception-2.0","name":"c"}}]},"label":"//:l"}`,
			Expected: "(MIT OR LicenseRef-acme) AND GPL-2.0-only WITH Classpath-exception-2.0",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			assert.Equal(t, c.Expected, parseLicense(t, c.Input).Root().String())
		})
	}
}

func TestLicenseExpressionKinds(t *testing.T) {
	a := parseLicense(t, `{"expression":{"operator":"OR","operands":[
		{"kind":{"identifier":"MIT","name":"m"}},
		{"operator":"WITH","operands":[{"kind":{"identifier":"MIT","name":"m"}}],
		 "exception":{"identifier":"AdditionRef-x","name":"x"}}]},"label":"//:l"}`)

	kinds := a.Root().Kinds()
	require.Len(t, kinds, 2)
	assert.Equal(t, "MIT", kinds[0].Identifier)
	assert.False(t, kinds[0].IsCustom())
	assert.Equal(t, "AdditionRef-x", kinds[1].Identifier)
	assert.True(t, kinds[1].IsCustom())
}

func TestLicenseAttributeRequiresExpression(t *testing.T) {
	_, err := LicenseAttributeDescriptor.Parser(strings.NewReader(`{"label":"//:l"}`))
	assert.Error(t, err)
}

func TestSingleLicenseIsLeafNode(t *testing.T) {
	root := parseLicense(t, `{"expression":{"kind":{"identifier":"MIT","name":"m"}},"label":"//:l"}`).Root()
	assert.True(t, root.IsSingle())
	assert.Empty(t, root.Operands)
}

func TestLicenseAttributeParserLegacyKind(t *testing.T) {
	a, err := LicenseAttributeDescriptor.Parser(strings.NewReader(
		`{"kind":{"identifier":"MIT","name":"MIT License"},"label":"//:l"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Root().String(); got != "MIT" || !a.Root().IsSingle() {
		t.Errorf("legacy kind decoded as %q", got)
	}

	// The expression takes precedence over the deprecated kind.
	a, err = LicenseAttributeDescriptor.Parser(strings.NewReader(
		`{"kind":null,"expression":{"operator":"OR","operands":[{"kind":{"identifier":"MIT","name":"M"}},{"kind":{"identifier":"ISC","name":"I"}}]},"label":"//:l"}`))
	if err != nil || a.Root().String() != "MIT OR ISC" {
		t.Errorf("expression not preferred: %v %v", a, err)
	}

	for _, input := range []string{
		`{"label":"//:l"}`,
	} {
		if _, err := LicenseAttributeDescriptor.Parser(strings.NewReader(input)); err == nil {
			t.Errorf("expected an error for %s", input)
		}
	}
}

func TestLicenseExpressionOrLaterAndCustomString(t *testing.T) {
	gpl := LicenseExpression{Kind: &LicenseKind{Identifier: "GPL-3.0", Name: "GPL"}, OrLater: true}
	if got := gpl.String(); got != "GPL-3.0+" {
		t.Errorf("got %q", got)
	}
	custom := LicenseExpression{Kind: &LicenseKind{Identifier: "generic-restricted", Custom: true}}
	if got := custom.String(); got != "LicenseRef-generic-restricted" {
		t.Errorf("got %q", got)
	}
	if !custom.Kind.IsCustom() || gpl.Kind.IsCustom() {
		t.Error("unexpected IsCustom")
	}
}

func TestInvalidCustomIdentifierIsRejected(t *testing.T) {
	for _, id := range []string{"acme_internal", "LicenseRef-a b", "LicenseRef-"} {
		_, err := LicenseAttributeDescriptor.Parser(strings.NewReader(`{"kind":{"identifier":"` + id + `","name":"n","custom":true},"label":"x"}`))
		if err == nil {
			t.Errorf("expected an error for %q", id)
		}
	}
}

func TestLicenseCustomOrLaterHasOwnReference(t *testing.T) {
	e := LicenseExpression{Kind: &LicenseKind{Identifier: "LicenseRef-acme", Custom: true}, OrLater: true}
	if got := e.String(); got != "LicenseRef-acme-or-later" {
		t.Errorf("got %q", got)
	}
}
