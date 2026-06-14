package pixelizer

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// SamplingMode defines how block colors are computed.
type SamplingMode string

const (
	ModeAverage  SamplingMode = "average"
	ModeMedian   SamplingMode = "median"
	ModeDominant SamplingMode = "dominant"
)

// ValidModes lists supported sampling modes.
var ValidModes = []SamplingMode{ModeAverage, ModeMedian, ModeDominant}

// IsValid returns true if the mode is supported.
func (m SamplingMode) IsValid() bool {
	for _, v := range ValidModes {
		if m == v {
			return true
		}
	}
	return false
}

// Config holds pixelization parameters.
type Config struct {
	BlockSize        int
	Scale            int
	Mode             SamplingMode
	Palette          Palette
	Logical          bool
	Colors           int
	Dither           bool
	Sharpen          bool
	FixedPalette     []color.RGBA
	TransparentWhite bool
}

// Validate returns an error if the config is invalid.
func (c Config) Validate() error {
	if c.BlockSize <= 0 {
		return fmt.Errorf("block size must be > 0")
	}
	if c.Scale <= 0 {
		return fmt.Errorf("scale must be > 0")
	}
	if c.Colors < 0 {
		return fmt.Errorf("colors must be >= 0")
	}
	if !c.Mode.IsValid() {
		return fmt.Errorf("unsupported mode: %q", c.Mode)
	}
	if !c.Palette.IsValid() {
		return fmt.Errorf("unsupported palette: %q", c.Palette)
	}
	return nil
}

// Pixelize produces the pixelized output image according to cfg.
func Pixelize(src image.Image, cfg Config) (*image.RGBA, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logical := pixelizeBlock(src, cfg.BlockSize, cfg.Mode, cfg.Palette)

	if len(cfg.FixedPalette) > 0 {
		logical = QuantizeFixed(logical, cfg.FixedPalette, cfg.Dither)
	} else if cfg.Colors > 0 {
		logical = Quantize(logical, cfg.Colors, cfg.Dither)
	}

	if cfg.Sharpen {
		logical = sharpen(logical)
	}

	if cfg.TransparentWhite {
		logical = whiteToTransparent(logical)
	}

	if cfg.Logical {
		return logical, nil
	}
	return ScaleNearest(logical, cfg.Scale), nil
}

func pixelizeBlock(src image.Image, blockSize int, mode SamplingMode, pal Palette) *image.RGBA {
	bounds := src.Bounds()
	outW := ceilDiv(bounds.Dx(), blockSize)
	outH := ceilDiv(bounds.Dy(), blockSize)
	dst := image.NewRGBA(image.Rect(0, 0, outW, outH))

	for by := 0; by < outH; by++ {
		for bx := 0; bx < outW; bx++ {
			sx := bounds.Min.X + bx*blockSize
			sy := bounds.Min.Y + by*blockSize
			ex := min(sx+blockSize, bounds.Max.X)
			ey := min(sy+blockSize, bounds.Max.Y)

			var c color.RGBA
			switch mode {
			case ModeMedian:
				c = medianColor(src, sx, sy, ex, ey)
			case ModeDominant:
				c = dominantColor(src, sx, sy, ex, ey)
			default:
				c = averageColor(src, sx, sy, ex, ey)
			}
			dst.SetRGBA(bx, by, pal.MapColor(c))
		}
	}
	return dst
}

// ScaleNearest scales src by the given factor using nearest-neighbor.
func ScaleNearest(src image.Image, scale int) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w*scale, h*scale))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := src.At(bounds.Min.X+x, bounds.Min.Y+y)
			rect := image.Rect(x*scale, y*scale, (x+1)*scale, (y+1)*scale)
			draw.Draw(dst, rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
		}
	}
	return dst
}

func whiteToTransparent(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			i := img.PixOffset(x, y)
			r, g, b := img.Pix[i], img.Pix[i+1], img.Pix[i+2]
			if r >= 250 && g >= 250 && b >= 250 {
				img.Pix[i+3] = 0
			}
		}
	}
	return img
}
