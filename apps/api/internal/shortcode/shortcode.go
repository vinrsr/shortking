package shortcode

import (
	"crypto/rand"
	"fmt"
	"regexp"
)

const (
	alphabet       = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	generatedLen   = 7
	MinAliasLen    = 3
	MaxAliasLen    = 32
	MaxGenAttempts = 5
)

var (
	aliasPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	reservedWords = map[string]bool{
		"api": true, "dashboard": true, "login": true, "signup": true,
		"logout": true, "health": true, "www": true, "admin": true,
		"static": true, "assets": true, "favicon.ico": true,
	}
)

// Generate returns a random base62 short code of the default length,
// using crypto/rand so codes aren't sequential/guessable.
func Generate() (string, error) {
	b := make([]byte, generatedLen)
	buf := make([]byte, generatedLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, v := range buf {
		b[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(b), nil
}

// ValidateAlias checks a user-supplied custom alias for format and
// reserved-word collisions. It does not check uniqueness against the
// database, callers should rely on the DB unique constraint for that,
// since check-then-insert has a race under concurrency.
func ValidateAlias(alias string) error {
	if len(alias) < MinAliasLen || len(alias) > MaxAliasLen {
		return fmt.Errorf("alias must be between %d and %d characters", MinAliasLen, MaxAliasLen)
	}
	if !aliasPattern.MatchString(alias) {
		return fmt.Errorf("alias may only contain letters, numbers, hyphens, and underscores")
	}
	if reservedWords[alias] {
		return fmt.Errorf("alias %q is reserved", alias)
	}
	return nil
}
