package clone

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "normal-filename.txt",
			expected: "normal-filename.txt",
		},
		{
			input:    "file/with/slashes",
			expected: "file_with_slashes",
		},
		{
			input:    "file\\with\\backslashes",
			expected: "file_with_backslashes",
		},
		{
			input:    "file:with:colons",
			expected: "file_with_colons",
		},
		{
			input:    "file*with?invalid<chars>",
			expected: "file_with_invalid_chars_",
		},
		{
			input:    "file\"with|quotes",
			expected: "file_with_quotes",
		},
		{
			input:    "  filename with spaces  ",
			expected: "filename with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilenameLongName(t *testing.T) {
	longName := string(make([]byte, 250))
	for i := range longName {
		longName = longName[:i] + "a" + longName[i+1:]
	}

	result := sanitizeFilename(longName)
	if len(result) > 200 {
		t.Errorf("sanitizeFilename did not truncate long filename, got length %d", len(result))
	}
}
