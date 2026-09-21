package util

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureRandomString(t *testing.T) {
	assert.Equal(t, "", SecureRandomString(0))

	re := regexp.MustCompile(`^[0-9A-Za-z]{43}$`)
	seen := map[string]struct{}{}
	for range 100 {
		s := SecureRandomString(43)
		assert.Regexp(t, re, s)
		seen[s] = struct{}{}
	}
	assert.Len(t, seen, 100)
}
