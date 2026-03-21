package vault

import "testing"

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Note", "My_Note"},
		{"Meeting: Notes", "Meeting-_Notes"},
		{"File/Path", "File-Path"},
		{"What?Why!", "What-Why!"},
		{"Colon:Test", "Colon-Test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
