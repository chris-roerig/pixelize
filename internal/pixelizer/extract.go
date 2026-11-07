package pixelizer

import (
	"image"
	"image/color"
	"sort"
)

// ExtractPalette returns the distinct fully-opaque colors present in img,
// sorted by luminance (dark to light) for a stable, readable ordering. Only
// fully opaque pixels count: at soft/anti-aliased edges the partial-alpha
// colors are blending artifacts, not part of the image's palette. This is used
// to dump the palette an output actually uses.
func ExtractPalette(img image.Image) []color.RGBA {
	bounds := img.Bounds()
	seen := make(map[color.RGBA]struct{})
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a>>8 < 255 {
				continue
			}
			seen[color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}] = struct{}{}
		}
	}
	out := make([]color.RGBA, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		return luminance(out[i]) < luminance(out[j])
	})
	return out
}

// luminance returns a perceptual brightness for sorting (Rec. 601 weights).
func luminance(c color.RGBA) float64 {
	return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
}
