package pixelizer

import "image/color"

// Palette defines a color palette strategy.
type Palette string

const (
	PaletteOriginal Palette = "original"
	PalettePico8    Palette = "pico8"
	PaletteGameboy  Palette = "gameboy"
	PaletteNES      Palette = "nes"
)

// ValidPalettes lists supported palette values.
var ValidPalettes = []Palette{PaletteOriginal, PalettePico8, PaletteGameboy, PaletteNES}

// IsValid returns true if the palette is supported.
func (p Palette) IsValid() bool {
	for _, v := range ValidPalettes {
		if p == v {
			return true
		}
	}
	return false
}

// MapColor returns the palette-mapped color.
func (p Palette) MapColor(c color.RGBA) color.RGBA {
	switch p {
	case PalettePico8:
		return nearestInPalette(c, pico8Colors)
	case PaletteGameboy:
		return nearestInPalette(c, gameboyColors)
	case PaletteNES:
		return nearestInPalette(c, nesColors)
	default:
		return c
	}
}

func nearestInPalette(c color.RGBA, pal []color.RGBA) color.RGBA {
	best := pal[0]
	bestDist := colorDist(c, best)
	for _, p := range pal[1:] {
		d := colorDist(c, p)
		if d < bestDist {
			bestDist = d
			best = p
		}
	}
	best.A = c.A
	return best
}

var pico8Colors = []color.RGBA{
	{0, 0, 0, 255},
	{29, 43, 83, 255},
	{126, 37, 83, 255},
	{0, 135, 81, 255},
	{171, 82, 54, 255},
	{95, 87, 79, 255},
	{194, 195, 199, 255},
	{255, 241, 232, 255},
	{255, 0, 77, 255},
	{255, 163, 0, 255},
	{255, 236, 39, 255},
	{0, 228, 54, 255},
	{41, 173, 255, 255},
	{131, 118, 156, 255},
	{255, 119, 168, 255},
	{255, 204, 170, 255},
}

var gameboyColors = []color.RGBA{
	{15, 56, 15, 255},
	{48, 98, 48, 255},
	{139, 172, 15, 255},
	{155, 188, 15, 255},
}

var nesColors = []color.RGBA{
	{0, 0, 0, 255},
	{252, 252, 252, 255},
	{248, 56, 0, 255},
	{0, 0, 188, 255},
	{68, 40, 188, 255},
	{148, 0, 132, 255},
	{168, 0, 32, 255},
	{168, 16, 0, 255},
	{136, 20, 0, 255},
	{80, 48, 0, 255},
	{0, 120, 0, 255},
	{0, 104, 0, 255},
	{0, 88, 0, 255},
	{0, 64, 88, 255},
	{0, 0, 0, 255},
	{188, 188, 188, 255},
	{0, 120, 248, 255},
	{0, 88, 248, 255},
	{104, 68, 252, 255},
	{216, 0, 204, 255},
	{228, 0, 88, 255},
	{248, 56, 0, 255},
	{228, 92, 16, 255},
	{172, 124, 0, 255},
	{0, 184, 0, 255},
	{0, 168, 0, 255},
	{0, 168, 68, 255},
	{0, 136, 136, 255},
	{124, 124, 124, 255},
	{248, 184, 248, 255},
	{248, 184, 248, 255},
	{216, 168, 248, 255},
	{248, 120, 248, 255},
	{248, 112, 176, 255},
	{248, 152, 120, 255},
	{248, 168, 88, 255},
	{248, 184, 0, 255},
	{184, 248, 24, 255},
	{88, 216, 84, 255},
	{88, 248, 152, 255},
	{0, 232, 216, 255},
	{120, 120, 120, 255},
	{252, 224, 168, 255},
	{184, 248, 184, 255},
	{184, 248, 216, 255},
	{0, 252, 252, 255},
	{248, 216, 248, 255},
}
