package pixelizer

import (
	"image"
	"image/color"
	"testing"
)

func TestExtractPalette(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{255, 255, 255, 255})
	img.SetRGBA(1, 0, color.RGBA{0, 0, 0, 255})
	img.SetRGBA(0, 1, color.RGBA{255, 255, 255, 255}) // duplicate of white
	img.SetRGBA(1, 1, color.RGBA{255, 0, 0, 255})

	pal := ExtractPalette(img)
	if len(pal) != 3 {
		t.Fatalf("expected 3 distinct colors, got %d: %v", len(pal), pal)
	}
	// Sorted dark -> light: black, red, white.
	if pal[0] != (color.RGBA{0, 0, 0, 255}) {
		t.Errorf("expected first color black, got %v", pal[0])
	}
	if pal[len(pal)-1] != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("expected last color white, got %v", pal[len(pal)-1])
	}
}

func TestExtractPaletteIgnoresTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{10, 20, 30, 255})
	img.SetRGBA(1, 0, color.RGBA{0, 0, 0, 0}) // fully transparent, ignored
	pal := ExtractPalette(img)
	if len(pal) != 1 {
		t.Fatalf("expected 1 color, got %d", len(pal))
	}
}

func TestApplyGridDarkensEdges(t *testing.T) {
	// A 4x4 image at step 2: the top row and left column of each 2x2 block
	// (x%2==0 or y%2==0) are darkened; interior pixels (x%2==1 && y%2==1) stay.
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, color.RGBA{200, 200, 200, 255})
		}
	}
	applyGrid(img, 2)

	// Interior pixel (1,1) untouched.
	r, _, _, _ := img.At(1, 1).RGBA()
	if r>>8 != 200 {
		t.Errorf("interior pixel should be unchanged, got %d", r>>8)
	}
	// Edge pixel (0,0) darkened.
	r0, _, _, _ := img.At(0, 0).RGBA()
	if r0>>8 >= 200 {
		t.Errorf("edge pixel should be darkened, got %d", r0>>8)
	}
}

func TestApplyGridNoopBelowStep2(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	img.SetRGBA(0, 0, color.RGBA{100, 100, 100, 255})
	applyGrid(img, 1)
	r, _, _, _ := img.At(0, 0).RGBA()
	if r>>8 != 100 {
		t.Errorf("step<2 should be a no-op, got %d", r>>8)
	}
}
