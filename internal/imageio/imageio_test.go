package imageio

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func sampleImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x * 60), uint8(y * 60), 128, 255})
		}
	}
	return img
}

func TestSaveLoadPNGRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.png")
	if err := Save(path, sampleImage()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Bounds().Dx() != 4 || got.Bounds().Dy() != 4 {
		t.Errorf("expected 4x4, got %v", got.Bounds())
	}
	// PNG is lossless; a known pixel should survive exactly.
	r, g, b, a := got.At(3, 0).RGBA()
	if uint8(r>>8) != 180 || uint8(g>>8) != 0 || uint8(b>>8) != 128 || uint8(a>>8) != 255 {
		t.Errorf("pixel mismatch after PNG round trip: %d,%d,%d,%d", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestSaveJPEGByExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jpg")
	if err := Save(path, sampleImage()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Verify the file starts with the JPEG SOI marker (0xFF 0xD8).
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		t.Errorf("expected JPEG SOI marker, got % x", data[:min(2, len(data))])
	}
	// And it should decode back as an image of the right size.
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load jpeg: %v", err)
	}
	if got.Bounds().Dx() != 4 || got.Bounds().Dy() != 4 {
		t.Errorf("expected 4x4, got %v", got.Bounds())
	}
}

func TestSaveDefaultsToPNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.bin") // unknown extension
	if err := Save(path, sampleImage()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// PNG magic number.
	png := []byte{0x89, 0x50, 0x4E, 0x47}
	if len(data) < 4 || string(data[:4]) != string(png) {
		t.Errorf("expected PNG magic for unknown extension, got % x", data[:min(4, len(data))])
	}
}
