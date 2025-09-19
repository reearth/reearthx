package util

import (
	"strings"
	"testing"
)

func TestIsSafePathName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"validname", true},
		{"valid-name", true},
		{"valid123", true},
		{"123valid", true},
		{"a1b2c3", true},
		{"ValidName", true},          // uppercase now allowed
		{"VALIDNAME", true},          // uppercase now allowed
		{"Valid-Name", true},         // mixed case now allowed
		{"invalid_name", false},      // underscores not allowed
		{"invalid@name", false},      // @ not allowed
		{"invalid.name", false},      // dots not allowed
		{"invalid@test.name", false}, // email not allowed
		{"invalid--name", false},     // consecutive hyphens not allowed
		{"-invalid", false},          // leading hyphen not allowed
		{"invalid-", false},          // trailing hyphen not allowed
		{"", false},                  // empty string
		// Whitespace trimming tests
		{" validname", true},  // leading space trimmed
		{"validname ", true},  // trailing space trimmed
		{" validname ", true}, // both spaces trimmed
		{" ValidName ", true}, // mixed case with spaces trimmed
		// Length boundary tests - now only 5-32 characters allowed
		{"a", false}, // 1 character - too short
		{"A", false}, // 1 uppercase character - too short
		{"1", false}, // 1 digit - too short
		{"a" + strings.Repeat("b", 5) + "c", true},   // 7 characters - valid range
		{"A" + strings.Repeat("B", 5) + "C", true},   // 7 uppercase characters - valid range
		{"a" + strings.Repeat("b", 30) + "c", true},  // 32 characters - maximum valid
		{"a" + strings.Repeat("b", 31) + "c", false}, // 33 characters - too long
		{"ab", false},     // 2 characters - too short
		{"abc", false},    // 3 characters - too short
		{"abcd", false},   // 4 characters - too short
		{"abcde", true},   // 5 characters - minimum valid
		{"abcdef", true},  // 6 characters - valid
		{"abcdefg", true}, // 7 characters - valid
		// Non-ASCII character tests - should all fail as per printascii constraint
		{"こんにちは", false},         // Japanese hiragana
		{"コンニチハ", false},         // Japanese katakana
		{"今日は", false},           // Japanese kanji
		{"你好世界", false},          // Chinese simplified
		{"您好世界", false},          // Chinese traditional
		{"안녕하세요", false},         // Korean
		{"مرحبا", false},         // Arabic
		{"Здравствуй", false},    // Russian
		{"café", false},          // French with accent
		{"naïve", false},         // English with diaeresis
		{"piñata", false},        // Spanish with tilde
		{"résumé", false},        // French accents
		{"😀😀😀😀😀", false},         // Emoji
		{"test😀name", false},     // Mixed ASCII and emoji
		{"testこんにちはname", false}, // Mixed ASCII and Japanese
		{"test你好name", false},    // Mixed ASCII and Chinese
		{"한글test", false},        // Mixed Korean and ASCII
		{"مرحباtest", false},     // Mixed Arabic and ASCII
		{"тестname", false},      // Mixed Cyrillic and ASCII
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsSafePathName(test.name)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}
