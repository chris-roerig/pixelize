package pixelizer

import "image"

// applyGrid darkens the top and left edge of every scaled block to produce a
// subtle pixel-grid / CRT separation effect. step is the scale factor (the size
// of one logical pixel in the output); a step < 2 leaves the image unchanged
// since there is nothing to separate. Fully transparent pixels are left alone.
func applyGrid(img *image.RGBA, step int) *image.RGBA {
	if step < 2 {
		return img
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	const darken = 0.72 // multiply edge pixels toward black
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x%step != 0 && y%step != 0 {
				continue // interior pixel of a block
			}
			i := img.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)
			if img.Pix[i+3] == 0 {
				continue
			}
			img.Pix[i] = uint8(float64(img.Pix[i]) * darken)
			img.Pix[i+1] = uint8(float64(img.Pix[i+1]) * darken)
			img.Pix[i+2] = uint8(float64(img.Pix[i+2]) * darken)
		}
	}
	return img
}
