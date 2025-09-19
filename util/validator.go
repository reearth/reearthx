package util

import (
	"errors"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{3,30}[a-zA-Z0-9]$`)

type TempName struct {
	Name string `validate:"required,min=5,max=32,printascii"`
}

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

	var tempName TempName
	tempName.Name = name

	validate := validator.New()
	if err := validate.Struct(&tempName); err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			return false
		}
		return false
	}

	return nameRegex.MatchString(name)
}
