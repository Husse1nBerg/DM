package utils

import "testing"

func TestFormatCustomerName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Standard format - lastname, firstname",
			input:    "Doe, John",
			expected: "John Doe",
		},
		{
			name:     "With extra spaces",
			input:    "Smith , Jane ",
			expected: "Jane Smith",
		},
		{
			name:     "Already in correct format",
			input:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "Single name only",
			input:    "Madonna",
			expected: "Madonna",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "With middle name",
			input:    "Kennedy, John F.",
			expected: "John F. Kennedy",
		},
		{
			name:     "Complex name with spaces",
			input:    "Van Der Berg, Hans",
			expected: "Hans Van Der Berg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCustomerName(tt.input)
			if result != tt.expected {
				t.Errorf("FormatCustomerName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
