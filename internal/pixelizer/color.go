package pixelizer

import (
	"image"
	"image/color"
	"sort"
)

func averageColor(img image.Image, startX, startY, endX, endY int) color.RGBA {
	var rSum, gSum, bSum float64
	var aSum float64
	var count int

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// RGBA() returns 16-bit premultiplied values. Reduce to 8-bit.
			a8 := a >> 8
			if a8 == 0 {
				continue
			}
			// Un-premultiply to straight sRGB, then linearize so averaging
			// happens in linear light (avoids the sRGB darkening bias).
			sr := uint8(clampU32((r >> 8) * 255 / a8))
			sg := uint8(clampU32((g >> 8) * 255 / a8))
			sb := uint8(clampU32((b >> 8) * 255 / a8))
			alpha := float64(a8)
			// Weight each color by its coverage (alpha) so translucent pixels
			// contribute proportionally.
			rSum += srgbToLinear[sr] * alpha
			gSum += srgbToLinear[sg] * alpha
			bSum += srgbToLinear[sb] * alpha
			aSum += alpha
			count++
		}
	}

	if aSum == 0 {
		return color.RGBA{}
	}

	return color.RGBA{
		R: linearToSrgb(rSum / aSum),
		G: linearToSrgb(gSum / aSum),
		B: linearToSrgb(bSum / aSum),
		A: uint8(aSum/float64(count) + 0.5),
	}
}

// clampU32 clamps a value to the 0..255 range (un-premultiply can exceed 255
// by one due to rounding).
func clampU32(v uint32) uint32 {
	if v > 255 {
		return 255
	}
	return v
}

// unpremul reads a pixel and returns straight (un-premultiplied) 8-bit sRGB
// channels plus the 8-bit alpha. Returns ok=false for fully transparent pixels.
func unpremul(img image.Image, x, y int) (r, g, b, a uint8, ok bool) {
	pr, pg, pb, pa := img.At(x, y).RGBA()
	a8 := pa >> 8
	if a8 == 0 {
		return 0, 0, 0, 0, false
	}
	r = uint8(clampU32((pr >> 8) * 255 / a8))
	g = uint8(clampU32((pg >> 8) * 255 / a8))
	b = uint8(clampU32((pb >> 8) * 255 / a8))
	a = uint8(a8)
	return r, g, b, a, true
}

func medianColor(img image.Image, startX, startY, endX, endY int) color.RGBA {
	var rs, gs, bs, as []uint8
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			r, g, b, a, ok := unpremul(img, x, y)
			if !ok {
				continue
			}
			rs = append(rs, r)
			gs = append(gs, g)
			bs = append(bs, b)
			as = append(as, a)
		}
	}
	if len(rs) == 0 {
		return color.RGBA{}
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i] < rs[j] })
	sort.Slice(gs, func(i, j int) bool { return gs[i] < gs[j] })
	sort.Slice(bs, func(i, j int) bool { return bs[i] < bs[j] })
	sort.Slice(as, func(i, j int) bool { return as[i] < as[j] })
	mid := len(rs) / 2
	return color.RGBA{R: rs[mid], G: gs[mid], B: bs[mid], A: as[mid]}
}

// dominantAccum tracks the pixels that fall into one quantization bucket so the
// winning bucket can report a true representative color instead of the
// bit-masked key.
type dominantAccum struct {
	count            int
	rSum, gSum, bSum float64 // linear-light sums
	aSum             float64
}

func dominantColor(img image.Image, startX, startY, endX, endY int) color.RGBA {
	// Bucket to 5 bits per channel for counting, but accumulate the real
	// (linear-light) colors so the winner is a representative, not a snapped key.
	buckets := make(map[uint32]*dominantAccum)
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			r, g, b, a, ok := unpremul(img, x, y)
			if !ok {
				continue
			}
			key := uint32(r&^0x07)<<16 | uint32(g&^0x07)<<8 | uint32(b&^0x07)
			acc := buckets[key]
			if acc == nil {
				acc = &dominantAccum{}
				buckets[key] = acc
			}
			acc.count++
			acc.rSum += srgbToLinear[r]
			acc.gSum += srgbToLinear[g]
			acc.bSum += srgbToLinear[b]
			acc.aSum += float64(a)
		}
	}
	if len(buckets) == 0 {
		return color.RGBA{}
	}
	var best *dominantAccum
	bestCount := 0
	for _, acc := range buckets {
		if acc.count > bestCount {
			bestCount = acc.count
			best = acc
		}
	}
	n := float64(best.count)
	return color.RGBA{
		R: linearToSrgb(best.rSum / n),
		G: linearToSrgb(best.gSum / n),
		B: linearToSrgb(best.bSum / n),
		A: uint8(best.aSum/n + 0.5),
	}
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
