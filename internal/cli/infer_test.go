package cli

import "testing"

func TestInferOutput(t *testing.T) {
	tests := []struct {
		input  string
		preset string
		want   string
	}{
		{"photo.png", "", "photo-pixelized.png"},
		{"photo.png", "gba", "photo-gba.png"},
		{"photo.jpg", "nes", "photo-nes.jpg"},
		{"dir/photo.png", "", "dir/photo-pixelized.png"},
		{"photo.png", "all", "photo-pixelized.png"}, // "all" is not a styling preset
		{"noext", "", "noext-pixelized.png"},
	}
	for _, tt := range tests {
		if got := inferOutput(tt.input, tt.preset); got != tt.want {
			t.Errorf("inferOutput(%q, %q) = %q, want %q", tt.input, tt.preset, got, tt.want)
		}
	}
}
