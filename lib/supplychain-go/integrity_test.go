package supplychain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseChecksums(t *testing.T) {
	sha256 := "sha256-" + strings.Repeat("A", 43) + "="
	got, err := ParseChecksums([]string{sha256, "md5-AAAA", "SHA1-" + strings.Repeat("A", 27)})
	require.NoError(t, err)
	assert.Equal(t, []Checksum{
		{Algorithm: "sha256", Hex: strings.Repeat("00", 32)},
		{Algorithm: "sha1", Hex: strings.Repeat("00", 20)},
	}, got)
}

func TestParseChecksumsMalformed(t *testing.T) {
	for _, sri := range []string{"sha256", "sha256-!!!", "sha256-AAAA"} {
		_, err := ParseChecksums([]string{sri})
		assert.Error(t, err, sri)
	}
}
