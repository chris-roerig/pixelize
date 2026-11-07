package pixelizer

import (
	"image"
	"image/color"
	"testing"
)

func TestSrgbLinearRoundTrip(t *testing.T) {
	// Round-tripping each 8-bit value through linear and back should land on
	// the same value (within the LUT's resolution).
	for i := 0; i < 256; i++ {
		got := linearToSrgb(srgbToLinear[i])
		if int(got)-i > 1 || i-int(got) > 1 {
			t.Errorf("round trip %d -> %d (off by more than 1)", i, got)
		}
	}
}

func TestLinearToSrgbClamps(t *testing.T) {
	if linearToSrgb(-0.5) != 0 {
		t.Errorf("expected 0 for negative input")
	}
	if linearToSrgb(2.0) != 255 {
		t.Errorf("expected 255 for >1 input")
	}
}

func TestRedmeanZeroForEqual(t *testing.T) {
	if d := redmean([3]uint8{10, 20, 30}, [3]uint8{10, 20, 30}); d != 0 {
		t.Errorf("expected 0 distance for identical colors, got %v", d)
	}
}

func TestRedmeanWeightsGreen(t *testing.T) {
	// A pure-green delta should weigh more than an equal blue delta at low red,
	// reflecting the human eye's green sensitivity in the redmean metric.
	base := [3]uint8{0, 0, 0}
	greenDelta := redmean(base, [3]uint8{0, 40, 0})
	blueDelta := redmean(base, [3]uint8{0, 0, 40})
	if greenDelta <= blueDelta {
		t.Errorf("expected green delta (%v) > blue delta (%v)", greenDelta, blueDelta)
	}
}

func TestNearestEmptyPaletteGuard(t *testing.T) {
	c := color.RGBA{1, 2, 3, 255}
	if got := nearest(c, nil); got != c {
		t.Errorf("expected identity for empty palette, got %v", got)
	}
}

func TestDitherPreservesAlpha(t *testing.T) {
	// A fully transparent logical image dithered to a palette must remain
	// transparent (regression for the dropped-alpha bug).
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, color.RGBA{128, 128, 128, 0})
		}
	}
	out := floydSteinberg(img, []color.RGBA{{0, 0, 0, 255}, {255, 255, 255, 255}})
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			_, _, _, a := out.At(x, y).RGBA()
			if a>>8 != 0 {
				t.Fatalf("expected alpha 0 at (%d,%d), got %d", x, y, a>>8)
			}
		}
	}
}

func TestNamedPaletteTakesPrecedenceOverColors(t *testing.T) {
	// With a named palette AND --colors set, output must only contain gameboy
	// palette colors (the palette wins; no median-cut double reduction).
	src := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src.SetRGBA(x, y, color.RGBA{uint8(x * 30), uint8(y * 30), 120, 255})
		}
	}
	cfg := Config{
		BlockSize: 2, Scale: 1, Mode: ModeAverage,
		Palette: PaletteGameboy, Colors: 32, Logical: true,
	}
	out, err := Pixelize(src, cfg)
	if err != nil {
		t.Fatal(err)
	}
	allowed := make(map[color.RGBA]bool)
	for _, c := range gameboyColors {
		allowed[c] = true
	}
	b := out.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := out.At(x, y).RGBA()
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), uint8(a >> 8)}
			if !allowed[c] {
				t.Fatalf("pixel (%d,%d) = %v is not a gameboy palette color", x, y, c)
			}
		}
	}
}

func TestDominantReturnsRepresentative(t *testing.T) {
	// A block that is mostly one color returns a color close to it, not a
	// bit-masked key snapped to a multiple of 8.
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for i := 0; i < 16; i++ {
		img.SetRGBA(i%4, i/4, color.RGBA{201, 101, 51, 255})
	}
	c := dominantColor(img, 0, 0, 4, 4)
	// If it returned the bit-masked key it would be (200,96,48); a real
	// representative stays near the actual color.
	if c.R < 199 || c.R > 203 {
		t.Errorf("expected R≈201, got %d", c.R)
	}
}
