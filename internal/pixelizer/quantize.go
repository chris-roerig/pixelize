package pixelizer

import (
	"image"
	"image/color"
	"sort"
)

// straightRGBA reads an internally-built *image.RGBA whose pixels were written
// with SetRGBA, i.e. stored as straight (non-premultiplied) bytes. It returns
// the 8-bit channels and reports ok=false for fully transparent pixels. This is
// distinct from unpremul, which is for decoded source images (premultiplied).
func straightRGBA(img *image.RGBA, x, y int) (r, g, b, a uint8, ok bool) {
	i := img.PixOffset(x, y)
	a = img.Pix[i+3]
	if a == 0 {
		return 0, 0, 0, 0, false
	}
	return img.Pix[i], img.Pix[i+1], img.Pix[i+2], a, true
}

// Quantize reduces the image to at most n colors using median-cut.
// If dither is true, Floyd-Steinberg error diffusion is applied.
func Quantize(img *image.RGBA, n int, dither bool) *image.RGBA {
	if n <= 0 {
		return img
	}

	bounds := img.Bounds()
	pixels := make([]color.RGBA, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a, ok := straightRGBA(img, x, y)
			if !ok {
				continue // don't let transparent pixels bias the palette
			}
			pixels = append(pixels, color.RGBA{r, g, b, a})
		}
	}

	palette := medianCut(pixels, n)

	if !dither {
		dst := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a, ok := straightRGBA(img, x, y)
				if !ok {
					continue // leave fully transparent pixels transparent
				}
				chosen := nearest(color.RGBA{r, g, b, 255}, palette)
				chosen.A = a
				dst.SetRGBA(x, y, chosen)
			}
		}
		return dst
	}

	return floydSteinberg(img, palette)
}

// QuantizeFixed maps all pixels to the nearest color in a fixed palette.
func QuantizeFixed(img *image.RGBA, palette []color.RGBA, dither bool) *image.RGBA {
	if dither {
		return floydSteinberg(img, palette)
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a, ok := straightRGBA(img, x, y)
			if !ok {
				continue
			}
			chosen := nearest(color.RGBA{r, g, b, 255}, palette)
			chosen.A = a
			dst.SetRGBA(x, y, chosen)
		}
	}
	return dst
}

func floydSteinberg(img *image.RGBA, palette []color.RGBA) *image.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Working buffer in int16 to handle error overflow
	buf := make([][3]int16, w*h)
	alpha := make([]uint8, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a, ok := straightRGBA(img, bounds.Min.X+x, bounds.Min.Y+y)
			i := y*w + x
			buf[i] = [3]int16{int16(r), int16(g), int16(b)}
			if ok {
				alpha[i] = a
			}
		}
	}

	dst := image.NewRGBA(bounds)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			old := buf[i]
			clamped := color.RGBA{
				R: clamp(old[0]),
				G: clamp(old[1]),
				B: clamp(old[2]),
				A: 255,
			}
			chosen := nearest(clamped, palette)
			// Carry the source pixel's alpha through unchanged; only RGB is
			// dithered. This keeps transparency intact on the dither path.
			chosen.A = alpha[i]
			dst.SetRGBA(bounds.Min.X+x, bounds.Min.Y+y, chosen)

			errR := old[0] - int16(chosen.R)
			errG := old[1] - int16(chosen.G)
			errB := old[2] - int16(chosen.B)

			diffuse(buf, w, h, x+1, y, errR, errG, errB, 7)
			diffuse(buf, w, h, x-1, y+1, errR, errG, errB, 3)
			diffuse(buf, w, h, x, y+1, errR, errG, errB, 5)
			diffuse(buf, w, h, x+1, y+1, errR, errG, errB, 1)
		}
	}
	return dst
}

func diffuse(buf [][3]int16, w, h, x, y int, errR, errG, errB int16, weight int16) {
	if x < 0 || x >= w || y < 0 || y >= h {
		return
	}
	i := y*w + x
	buf[i][0] += errR * weight / 16
	buf[i][1] += errG * weight / 16
	buf[i][2] += errB * weight / 16
}

func clamp(v int16) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func medianCut(pixels []color.RGBA, n int) []color.RGBA {
	if len(pixels) == 0 {
		return nil
	}
	buckets := [][]color.RGBA{pixels}
	for len(buckets) < n {
		// Find the bucket with the largest range to split
		best := 0
		bestRange := 0
		for i, b := range buckets {
			if len(b) < 2 {
				continue
			}
			r := channelRange(b)
			if r > bestRange {
				bestRange = r
				best = i
			}
		}
		if bestRange == 0 {
			break
		}
		b := buckets[best]
		ch := dominantChannel(b)
		sort.Slice(b, func(i, j int) bool {
			return channelVal(b[i], ch) < channelVal(b[j], ch)
		})
		mid := len(b) / 2
		buckets[best] = b[:mid]
		buckets = append(buckets, b[mid:])
	}

	palette := make([]color.RGBA, len(buckets))
	for i, b := range buckets {
		palette[i] = averageBucket(b)
	}
	return palette
}

func channelRange(pixels []color.RGBA) int {
	var minR, minG, minB uint8 = 255, 255, 255
	var maxR, maxG, maxB uint8
	for _, p := range pixels {
		if p.R < minR {
			minR = p.R
		}
		if p.R > maxR {
			maxR = p.R
		}
		if p.G < minG {
			minG = p.G
		}
		if p.G > maxG {
			maxG = p.G
		}
		if p.B < minB {
			minB = p.B
		}
		if p.B > maxB {
			maxB = p.B
		}
	}
	rR := int(maxR) - int(minR)
	rG := int(maxG) - int(minG)
	rB := int(maxB) - int(minB)
	m := rR
	if rG > m {
		m = rG
	}
	if rB > m {
		m = rB
	}
	return m
}

func dominantChannel(pixels []color.RGBA) int {
	var minR, minG, minB uint8 = 255, 255, 255
	var maxR, maxG, maxB uint8
	for _, p := range pixels {
		if p.R < minR {
			minR = p.R
		}
		if p.R > maxR {
			maxR = p.R
		}
		if p.G < minG {
			minG = p.G
		}
		if p.G > maxG {
			maxG = p.G
		}
		if p.B < minB {
			minB = p.B
		}
		if p.B > maxB {
			maxB = p.B
		}
	}
	rR := int(maxR) - int(minR)
	rG := int(maxG) - int(minG)
	rB := int(maxB) - int(minB)
	if rR >= rG && rR >= rB {
		return 0
	}
	if rG >= rB {
		return 1
	}
	return 2
}

func channelVal(c color.RGBA, ch int) uint8 {
	switch ch {
	case 0:
		return c.R
	case 1:
		return c.G
	default:
		return c.B
	}
}

func averageBucket(pixels []color.RGBA) color.RGBA {
	if len(pixels) == 0 {
		return color.RGBA{}
	}
	var r, g, b, a float64
	for _, p := range pixels {
		// Average in linear light to avoid the sRGB darkening bias.
		r += srgbToLinear[p.R]
		g += srgbToLinear[p.G]
		b += srgbToLinear[p.B]
		a += float64(p.A)
	}
	n := float64(len(pixels))
	return color.RGBA{
		R: linearToSrgb(r / n),
		G: linearToSrgb(g / n),
		B: linearToSrgb(b / n),
		A: uint8(a/n + 0.5),
	}
}

func nearest(c color.RGBA, palette []color.RGBA) color.RGBA {
	if len(palette) == 0 {
		return c
	}
	best := palette[0]
	bestDist := colorDist(c, best)
	for _, p := range palette[1:] {
		d := colorDist(c, p)
		if d < bestDist {
			bestDist = d
			best = p
		}
	}
	return best
}

// colorDist returns a perceptual distance between two colors (redmean). Lower
// is closer.
func colorDist(a, b color.RGBA) float64 {
	return redmean([3]uint8{a.R, a.G, a.B}, [3]uint8{b.R, b.G, b.B})
}
