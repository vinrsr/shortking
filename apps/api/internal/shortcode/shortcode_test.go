package shortcode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_ProducesCodeOfExpectedLengthAndAlphabet(t *testing.T) {
	code, err := Generate()
	require.NoError(t, err)
	assert.Len(t, code, generatedLen)
	for _, r := range code {
		assert.Truef(t, strings.ContainsRune(alphabet, r), "unexpected character %q in generated code %q", r, code)
	}
}

func TestGenerate_ProducesDifferentCodesAcrossCalls(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 20; i++ {
		code, err := Generate()
		require.NoError(t, err)
		assert.False(t, seen[code], "Generate produced a duplicate code %q within %d calls", code, i+1)
		seen[code] = true
	}
}

func TestValidateAlias(t *testing.T) {
	cases := []struct {
		name    string
		alias   string
		wantErr bool
	}{
		{"valid alias", "my-link_1", false},
		{"minimum length", strings.Repeat("a", MinAliasLen), false},
		{"maximum length", strings.Repeat("a", MaxAliasLen), false},
		{"too short", strings.Repeat("a", MinAliasLen-1), true},
		{"too long", strings.Repeat("a", MaxAliasLen+1), true},
		{"contains space", "my link", true},
		{"contains special char", "my/link", true},
		{"reserved word", "admin", true},
		{"reserved word dashboard", "dashboard", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAlias(tc.alias)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
