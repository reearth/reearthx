package util

import (
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"valid-name", true},
		{"valid_name", true},
		{"valid@name", true},
		{"valid.name", true},
		{"invalid--name", false},
		{"invalid__name", false},
		{"invalid..name", false},
		{"invalid@@name", false},
		{"invalid@name.com", false},
		// Empty string test
		{"", false},
		// Whitespace trimming tests
		{" validname", true},
		{"validname ", true},
		{" validname ", true},
		// Case insensitivity tests
		{"ValidName", true},
		{"VALIDNAME", true},
		// Length boundary tests
		{"a", true}, // 1 character - minimum
		{"a" + strings.Repeat("b", 61) + "c", true},  // 63 characters - maximum
		{"a" + strings.Repeat("b", 62) + "c", false}, // 64 characters - exceeds maximum
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsValidName(test.name)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}
