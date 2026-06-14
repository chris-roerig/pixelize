package pixelizer

import (
	"image"
	"image/color"
	"testing"
)

func TestCeilDiv(t *testing.T) {
	tests := []struct{ a, b, want int }{
		{1024, 8, 128},
		{1024, 16, 64},
		{100, 8, 13},
		{7, 8, 1},
		{8, 8, 1},
		{9, 8, 2},
	}
	for _, tt := range tests {
		if got := ceilDiv(tt.a, tt.b); got != tt.want {
			t.Errorf("ceilDiv(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestAverageColor(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{100, 0, 0, 255})
	img.SetRGBA(1, 0, color.RGBA{200, 0, 0, 255})
	img.SetRGBA(0, 1, color.RGBA{100, 0, 0, 255})
	img.SetRGBA(1, 1, color.RGBA{200, 0, 0, 255})

	avg := averageColor(img, 0, 0, 2, 2)
	if avg.R != 150 {
		t.Errorf("expected R=150, got %d", avg.R)
	}
	if avg.A != 255 {
		t.Errorf("expected A=255, got %d", avg.A)
	}
}

func TestAverageColorEmpty(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	avg := averageColor(img, 2, 2, 2, 2) // zero-area
	if avg != (color.RGBA{}) {
		t.Errorf("expected zero RGBA for empty region, got %v", avg)
	}
}

func TestPixelizeLogicalSize(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	cfg := Config{BlockSize: 8, Scale: 1, Mode: ModeAverage, Palette: PaletteOriginal, Logical: true}
	out, err := Pixelize(src, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if out.Bounds().Dx() != 128 || out.Bounds().Dy() != 128 {
		t.Errorf("expected 128x128, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}

func TestPixelizeScaledSize(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	cfg := Config{BlockSize: 8, Scale: 4, Mode: ModeAverage, Palette: PaletteOriginal, Logical: false}
	out, err := Pixelize(src, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if out.Bounds().Dx() != 512 || out.Bounds().Dy() != 512 {
		t.Errorf("expected 512x512, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}

func TestPixelizeUnevenDimensions(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	cfg := Config{BlockSize: 8, Scale: 1, Mode: ModeAverage, Palette: PaletteOriginal, Logical: true}
	out, err := Pixelize(src, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if out.Bounds().Dx() != 13 || out.Bounds().Dy() != 13 {
		t.Errorf("expected 13x13, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		ok   bool
	}{
		{"valid", Config{BlockSize: 8, Scale: 1, Mode: ModeAverage, Palette: PaletteOriginal}, true},
		{"zero block", Config{BlockSize: 0, Scale: 1, Mode: ModeAverage, Palette: PaletteOriginal}, false},
		{"negative scale", Config{BlockSize: 8, Scale: -1, Mode: ModeAverage, Palette: PaletteOriginal}, false},
		{"bad mode", Config{BlockSize: 8, Scale: 1, Mode: "invalid", Palette: PaletteOriginal}, false},
		{"bad palette", Config{BlockSize: 8, Scale: 1, Mode: ModeAverage, Palette: "invalid"}, false},
	}
	for _, tt := range tests {
		err := tt.cfg.Validate()
		if tt.ok && err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("%s: expected error", tt.name)
		}
	}
}
