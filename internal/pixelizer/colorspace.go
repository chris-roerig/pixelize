package pixelizer

import "math"

// Color operations default to perceptually meaningful behavior:
//   - averaging is done in linear light (not gamma-encoded sRGB), which avoids
//     the darkening bias of averaging sRGB bytes directly;
//   - nearest-color matching uses a "redmean" weighted distance, a cheap
//     approximation of perceived color difference that tracks human vision far
//     better than plain Euclidean RGB distance.

// srgbToLinear maps an 8-bit sRGB channel value to linear light in [0, 1].
var srgbToLinear [256]float64

// linearToSrgbByte maps a linear-light value in [0, 1] back to an 8-bit sRGB
// value. A 4096-entry table keeps the inverse lookup fast and accurate.
var linearToSrgbByte [4097]uint8

func init() {
	for i := 0; i < 256; i++ {
		c := float64(i) / 255.0
		if c <= 0.04045 {
			srgbToLinear[i] = c / 12.92
		} else {
			srgbToLinear[i] = math.Pow((c+0.055)/1.055, 2.4)
		}
	}
	for i := 0; i <= 4096; i++ {
		l := float64(i) / 4096.0
		var c float64
		if l <= 0.0031308 {
			c = l * 12.92
		} else {
			c = 1.055*math.Pow(l, 1.0/2.4) - 0.055
		}
		v := int(math.Round(c * 255.0))
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}
		linearToSrgbByte[i] = uint8(v)
	}
}

// linearToSrgb converts a linear-light value in [0, 1] to an 8-bit sRGB value.
func linearToSrgb(l float64) uint8 {
	if l <= 0 {
		return 0
	}
	if l >= 1 {
		return 255
	}
	return linearToSrgbByte[int(l*4096.0+0.5)]
}

// redmean returns a perceptual color distance between two sRGB colors. Lower is
// closer. The result is a squared magnitude (no sqrt) suitable for comparison.
// See the "redmean" low-cost approximation of CIE76 color difference.
func redmean(a, b [3]uint8) float64 {
	rbar := (float64(a[0]) + float64(b[0])) / 2.0
	dr := float64(a[0]) - float64(b[0])
	dg := float64(a[1]) - float64(b[1])
	db := float64(a[2]) - float64(b[2])
	return (2.0+rbar/256.0)*dr*dr + 4.0*dg*dg + (2.0+(255.0-rbar)/256.0)*db*db
}
