package pixelizer

// Palette is the name of a color palette. "original" (or empty) means no
// palette mapping — the computed colors are used as-is. All other palettes are
// defined as data (see the cli package's palette loader) and their resolved
// colors are passed to Pixelize via Config.PaletteColors.
type Palette string

const (
	// PaletteOriginal performs no color mapping.
	PaletteOriginal Palette = "original"
)
