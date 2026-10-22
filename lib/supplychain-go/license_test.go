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
			Input:    `{"kind":{"identifier":"MIT","name":"MIT License"},"label":"//:l"}`,
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

func TestLicenseAttributeRequiresExactlyOne(t *testing.T) {
	for _, s := range []string{
		`{"label":"//:l"}`,
		`{"kind":{"identifier":"MIT","name":"m"},"expression":{"kind":{"identifier":"MIT","name":"m"}},"label":"//:l"}`,
	} {
		_, err := LicenseAttributeDescriptor.Parser(strings.NewReader(s))
		assert.Error(t, err)
	}
}
