package pixelizer

import (
	"image"
	"image/color"
	"sort"
)

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
			r, g, b, a := img.At(x, y).RGBA()
			pixels = append(pixels, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
		}
	}

	palette := medianCut(pixels, n)

	if !dither {
		dst := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
				dst.SetRGBA(x, y, nearest(c, palette))
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
			r, g, b, a := img.At(x, y).RGBA()
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			dst.SetRGBA(x, y, nearest(c, palette))
		}
	}
	return dst
}

func floydSteinberg(img *image.RGBA, palette []color.RGBA) *image.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Working buffer in int16 to handle error overflow
	buf := make([][3]int16, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			i := y*w + x
			buf[i] = [3]int16{int16(r >> 8), int16(g >> 8), int16(b >> 8)}
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
	var r, g, b, a uint64
	for _, p := range pixels {
		r += uint64(p.R)
		g += uint64(p.G)
		b += uint64(p.B)
		a += uint64(p.A)
	}
	n := uint64(len(pixels))
	return color.RGBA{uint8(r / n), uint8(g / n), uint8(b / n), uint8(a / n)}
}

func nearest(c color.RGBA, palette []color.RGBA) color.RGBA {
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

func colorDist(a, b color.RGBA) int {
	dr := int(a.R) - int(b.R)
	dg := int(a.G) - int(b.G)
	db := int(a.B) - int(b.B)
	return dr*dr + dg*dg + db*db
}
