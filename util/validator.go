package util

import (
	"regexp"
	"strings"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{3,30}[a-zA-Z0-9]$`)

// IsSafePathName
// Compatible with Auth0's restricted character set (subset only: lowercase letters, numbers, and hyphens)
// Regex explanation:
// ^                 // Start of string
// [a-z0-9]          // First character must be a lowercase letter or digit
// (?:               // Start non-capturing group:
//
//	[a-z0-9-]{5,32}  // Allow 5 to 32 lowercase letters, digits, or hyphens
//	[a-z0-9]         // Final character must be a lowercase letter or digit
//
// )?                // Group is optional to allow 1-character usernames
// $                 // End of string
//
// Notes:
// - Prevents leading/trailing hyphens
// - Does not allow special characters beyond a-z, A-Z, 0-9, hyphens
// - Safe for use in subdomains and URL path segments
// - Does NOT allow consecutive hyphens; add extra logic in Go if needed
func IsSafePathName(name string) bool {
	name = strings.TrimSpace(name)
	char := "-"

	if strings.Contains(name, char+char) {
		return false
	}

	return nameRegex.MatchString(name)
}
