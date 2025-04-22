package util

import "testing"

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
