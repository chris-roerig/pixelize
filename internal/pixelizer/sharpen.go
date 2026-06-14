package pixelizer

import (
	"image"
	"image/color"
)

// sharpen applies a 3x3 unsharp convolution kernel for edge enhancement.
func sharpen(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(bounds)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := bounds.Min.X + x
			py := bounds.Min.Y + y

			r0, g0, b0, a0 := img.At(px, py).RGBA()
			r := int(r0>>8) * 5
			g := int(g0>>8) * 5
			b := int(b0>>8) * 5

			for _, off := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
				nx, ny := px+off[0], py+off[1]
				if nx < bounds.Min.X || nx >= bounds.Max.X || ny < bounds.Min.Y || ny >= bounds.Max.Y {
					r -= int(r0 >> 8)
					g -= int(g0 >> 8)
					b -= int(b0 >> 8)
				} else {
					nr, ng, nb, _ := img.At(nx, ny).RGBA()
					r -= int(nr >> 8)
					g -= int(ng >> 8)
					b -= int(nb >> 8)
				}
			}

			dst.SetRGBA(px, py, color.RGBA{
				R: clamp(int16(r)),
				G: clamp(int16(g)),
				B: clamp(int16(b)),
				A: uint8(a0 >> 8),
			})
		}
	}
	return dst
}
