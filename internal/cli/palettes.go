package cli

import (
	"embed"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chris-roerig/pixelize/internal/imageio"
	"github.com/chris-roerig/pixelize/internal/pixelizer"
)

//go:embed palettes/*.json
var builtinPalettes embed.FS

// palette is a named, community-definable color palette loaded from JSON.
type palette struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Colors      []string `json:"colors"`
}

// resolved holds a palette with its colors parsed into RGBA.
type resolvedPalette struct {
	Name        string
	Description string
	Colors      []color.RGBA
}

// parseHexColor parses "#rgb", "#rrggbb", or "#rrggbbaa" (with or without the
// leading '#') into an opaque-by-default color.RGBA.
func parseHexColor(s string) (color.RGBA, error) {
	h := strings.TrimPrefix(strings.TrimSpace(s), "#")
	switch len(h) {
	case 3: // #rgb -> #rrggbb
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	case 6, 8:
	default:
		return color.RGBA{}, fmt.Errorf("invalid hex color %q", s)
	}
	var vals [4]uint8
	vals[3] = 255
	for i := 0; i < len(h)/2; i++ {
		b, err := hexByte(h[i*2], h[i*2+1])
		if err != nil {
			return color.RGBA{}, fmt.Errorf("invalid hex color %q: %w", s, err)
		}
		vals[i] = b
	}
	return color.RGBA{vals[0], vals[1], vals[2], vals[3]}, nil
}

func hexByte(hi, lo byte) (uint8, error) {
	h, err := hexNibble(hi)
	if err != nil {
		return 0, err
	}
	l, err := hexNibble(lo)
	if err != nil {
		return 0, err
	}
	return h<<4 | l, nil
}

func hexNibble(c byte) (uint8, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	}
	return 0, fmt.Errorf("invalid hex digit %q", string(c))
}

// parsePalette decodes a palette from JSON and resolves its colors.
func parsePalette(data []byte, fallbackName string) (resolvedPalette, error) {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	var p palette
	if err := dec.Decode(&p); err != nil {
		return resolvedPalette{}, err
	}
	if p.Name == "" {
		p.Name = fallbackName
	}
	if len(p.Colors) == 0 {
		return resolvedPalette{}, fmt.Errorf("palette %q has no colors", p.Name)
	}
	colors := make([]color.RGBA, 0, len(p.Colors))
	for _, hex := range p.Colors {
		c, err := parseHexColor(hex)
		if err != nil {
			return resolvedPalette{}, err
		}
		colors = append(colors, c)
	}
	return resolvedPalette{Name: p.Name, Description: p.Description, Colors: colors}, nil
}

func userPaletteDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "pixelize", "palettes")
}

func projectPaletteDir() string {
	return filepath.Join(".pixelize", "palettes")
}

// loadPalettes builds the palette catalog: built-in (embedded) < user (XDG) <
// project-local. "original" is implicit and not a file.
func loadPalettes() (map[string]resolvedPalette, error) {
	palettes := make(map[string]resolvedPalette)

	entries, err := builtinPalettes.ReadDir("palettes")
	if err != nil {
		return nil, fmt.Errorf("read embedded palettes: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasPrefix(e.Name(), "._") {
			continue
		}
		data, err := builtinPalettes.ReadFile("palettes/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read embedded palette %s: %w", e.Name(), err)
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		p, err := parsePalette(data, name)
		if err != nil {
			return nil, fmt.Errorf("built-in palette %s: %w", e.Name(), err)
		}
		palettes[p.Name] = p
	}

	for _, dir := range []string{userPaletteDir(), projectPaletteDir()} {
		if dir == "" {
			continue
		}
		if err := loadPaletteDir(dir, palettes); err != nil {
			return nil, err
		}
	}
	return palettes, nil
}

func loadPaletteDir(dir string, palettes map[string]resolvedPalette) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read palette dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasPrefix(e.Name(), "._") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read palette %s: %w", path, err)
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		p, err := parsePalette(data, name)
		if err != nil {
			return fmt.Errorf("palette %s: %w", path, err)
		}
		palettes[p.Name] = p
	}
	return nil
}

func sortedPaletteNames(palettes map[string]resolvedPalette) []string {
	names := make([]string, 0, len(palettes)+1)
	names = append(names, "original")
	for k := range palettes {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func paletteNames(palettes map[string]resolvedPalette) string {
	return strings.Join(sortedPaletteNames(palettes), ", ")
}

// printPalettes writes an aligned, human-readable palette list to stdout.
func printPalettes(palettes map[string]resolvedPalette) {
	names := sortedPaletteNames(palettes)
	width := 0
	for _, n := range names {
		if len(n) > width {
			width = len(n)
		}
	}
	for _, n := range names {
		if n == "original" {
			fmt.Printf("  %-*s  %s\n", width, n, "use the computed colors as-is (no mapping)")
			continue
		}
		p := palettes[n]
		desc := p.Description
		if desc == "" {
			desc = fmt.Sprintf("%d colors", len(p.Colors))
		}
		fmt.Printf("  %-*s  %s (%d colors)\n", width, n, desc, len(p.Colors))
	}
}

// resolvePaletteColors maps a palette name to its colors. "original" or empty
// returns nil (no mapping). A path to a .json file loads that palette directly.
// An unknown name is an error listing the available palettes.
func resolvePaletteColors(name string, palettes map[string]resolvedPalette) ([]color.RGBA, error) {
	if name == "" || name == "original" {
		return nil, nil
	}
	if looksLikePresetPath(name) { // reuse: same path heuristic as presets
		data, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("read palette %s: %w", name, err)
		}
		p, err := parsePalette(data, strings.TrimSuffix(filepath.Base(name), ".json"))
		if err != nil {
			return nil, fmt.Errorf("palette %s: %w", name, err)
		}
		return p.Colors, nil
	}
	p, ok := palettes[name]
	if !ok {
		return nil, fmt.Errorf("unknown palette %q (available: %s)", name, paletteNames(palettes))
	}
	return p.Colors, nil
}

// writePaletteFile processes input through cfg and writes the resulting
// palette to a JSON file usable with --palette.
func writePaletteFile(input, outPath string, cfg pixelizer.Config) error {
	src, err := imageio.Load(input)
	if err != nil {
		return err
	}
	out, err := pixelizer.Pixelize(src, cfg)
	if err != nil {
		return err
	}
	colors := pixelizer.ExtractPalette(out)
	hex := make([]string, len(colors))
	for i, c := range colors {
		hex[i] = fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	name := strings.TrimSuffix(filepath.Base(outPath), ".json")
	p := palette{
		Name:        name,
		Description: fmt.Sprintf("%d colors extracted from %s", len(hex), filepath.Base(input)),
		Colors:      hex,
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return fmt.Errorf("write palette %s: %w", outPath, err)
	}
	fmt.Fprintf(os.Stderr, "Wrote %s (%d colors)\n", outPath, len(hex))
	return nil
}
