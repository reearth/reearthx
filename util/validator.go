package util

import (
	"regexp"
	"strings"
)

// IsValidName
// Compatible with Auth0's restricted character set (subset only: lowercase letters, numbers, hyphens, underscores, at (@), and dots (.))
// Regex explanation:
// ^                 // Start of string
// [a-z0-9]          // First character must be a lowercase letter or digit
// (?:               // Start non-capturing group:
//
//	[a-z0-9-]{1,61}  // Allow 1 to 61 lowercase letters, digits, or hyphens
//	[a-z0-9]         // Final character must be a lowercase letter or digit
//
// )?                // Group is optional to allow 1-character usernames
// $                 // End of string
//
// Notes:
// - Prevents leading/trailing hyphens
// - Does not allow special characters beyond a-z, 0-9, hyphens, underscores, at (@), and dots (.)
// - Safe for use in subdomains and URL path segments
// - Does NOT allow consecutive hyphens; add extra logic in Go if needed
func IsValidName(name string) bool {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	nameRegex := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-_@.]{0,61}[a-z0-9])?$`)
	chars := []string{"-", "_", ".", "@"}

	for _, c := range chars {
		if strings.Contains(name, c+c) {
			return false
		}
	}

	// Check if it's an email address (not allowed)
	if strings.Contains(name, "@") && strings.Contains(name, ".") {
		return false
	}

	return nameRegex.MatchString(name)
}
