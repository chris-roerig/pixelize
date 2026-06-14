package pixelizer

import (
	"image"
	"image/color"
	"sort"
)

func averageColor(img image.Image, startX, startY, endX, endY int) color.RGBA {
	var rSum, gSum, bSum uint64
	var aSum uint64
	var count uint64

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			alpha := uint64(a >> 8)
			if alpha == 0 {
				continue
			}
			// Weight by alpha for proper transparency handling
			rSum += uint64(r>>8) * alpha
			gSum += uint64(g>>8) * alpha
			bSum += uint64(b>>8) * alpha
			aSum += alpha
			count++
		}
	}

	if aSum == 0 {
		return color.RGBA{}
	}

	return color.RGBA{
		R: uint8(rSum / aSum),
		G: uint8(gSum / aSum),
		B: uint8(bSum / aSum),
		A: uint8(aSum / count),
	}
}

func medianColor(img image.Image, startX, startY, endX, endY int) color.RGBA {
	var rs, gs, bs, as []uint8
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a>>8 == 0 {
				continue
			}
			rs = append(rs, uint8(r>>8))
			gs = append(gs, uint8(g>>8))
			bs = append(bs, uint8(b>>8))
			as = append(as, uint8(a>>8))
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

func dominantColor(img image.Image, startX, startY, endX, endY int) color.RGBA {
	counts := make(map[color.RGBA]int)
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a>>8 == 0 {
				continue
			}
			// Quantize to reduce unique colors (5-bit per channel)
			c := color.RGBA{
				R: uint8(r>>8) &^ 0x07,
				G: uint8(g>>8) &^ 0x07,
				B: uint8(b>>8) &^ 0x07,
				A: uint8(a >> 8),
			}
			counts[c]++
		}
	}
	if len(counts) == 0 {
		return color.RGBA{}
	}
	var best color.RGBA
	bestCount := 0
	for c, n := range counts {
		if n > bestCount {
			bestCount = n
			best = c
		}
	}
	return best
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
