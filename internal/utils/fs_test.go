package utils

import "testing"

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Normal File.mp3", "Normal File.mp3"},
		{"Bad/File:Name.mp3", "BadFileName.mp3"},
		{"<Hidden>", "Hidden"},
		{"What? Why*", "What Why"},
		{"Emoji 🎵", "Emoji 🎵"}, // Should preserve unicode
		{"", ""},
	}

	for _, tt := range tests {
		got := SanitizeFilename(tt.input)
		if got != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
