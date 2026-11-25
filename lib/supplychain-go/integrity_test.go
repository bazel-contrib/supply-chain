package supplychain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseChecksums(t *testing.T) {
	sha256 := "sha256-" + strings.Repeat("A", 43) + "="
	got, err := ParseChecksums([]string{sha256, "md5-AAAA", "SHA384-" + strings.Repeat("A", 64)})
	require.NoError(t, err)
	assert.Equal(t, []Checksum{
		{Algorithm: "sha256", Hex: strings.Repeat("00", 32)},
		{Algorithm: "sha384", Hex: strings.Repeat("00", 48)},
	}, got)
}

func TestParseChecksumsMalformed(t *testing.T) {
	for _, sri := range []string{
		"sha256", "sha256-!!!", "sha256-AAAA",
		// A length of 1 modulo 4 is impossible in base64.
		"sha384-" + strings.Repeat("A", 65),
		// At most two padding characters, completing a multiple of 4.
		"sha256-" + strings.Repeat("A", 43) + "====",
		"sha256-" + strings.Repeat("A", 42) + "=",
	} {
		_, err := ParseChecksums([]string{sri})
		assert.Error(t, err, sri)
	}
}
