package cli

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chris-roerig/pixelize/internal/pixelizer"
)

//go:embed presets/*.json
var builtinPresets embed.FS

// preset holds the styling parameters a named preset applies as a base.
// Explicit command-line flags always override a preset's values.
type preset struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Block       int    `json:"block"`
	Colors      int    `json:"colors"`
	Dither      bool   `json:"dither"`
	Sharpen     bool   `json:"sharpen"`
	Mode        string `json:"mode"`
	Palette     string `json:"palette"`
}

// validate ensures the preset describes a usable configuration. It reuses the
// pixelizer config validation so a malformed community preset fails at load
// time rather than mid-render.
func (p preset) validate() error {
	if p.Name == "" {
		return errors.New("preset is missing a name")
	}
	cfg := pixelizer.Config{
		BlockSize: p.Block,
		Scale:     p.Block, // presets scale by block; any positive value validates the rest
		Colors:    p.Colors,
		Dither:    p.Dither,
		Sharpen:   p.Sharpen,
		Mode:      pixelizer.SamplingMode(p.Mode),
		Palette:   pixelizer.Palette(p.Palette),
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	return nil
}

// parsePreset decodes a single preset from JSON bytes, defaulting the name to
// the file's base name when the JSON omits it.
func parsePreset(data []byte, fallbackName string) (preset, error) {
	var p preset
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return preset{}, err
	}
	if p.Name == "" {
		p.Name = fallbackName
	}
	return p, nil
}

// userPresetDir returns the XDG config directory for user presets.
func userPresetDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "pixelize", "presets")
}

// projectPresetDir returns the project-local preset directory.
func projectPresetDir() string {
	return filepath.Join(".pixelize", "presets")
}

// loadPresets builds the full preset catalog by layering sources in order of
// increasing precedence: built-in (embedded) < user (XDG) < project-local.
// A later source with the same name overrides an earlier one.
func loadPresets() (map[string]preset, error) {
	presets := make(map[string]preset)

	// 1. Built-in embedded presets.
	entries, err := builtinPresets.ReadDir("presets")
	if err != nil {
		return nil, fmt.Errorf("read embedded presets: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasPrefix(e.Name(), "._") {
			continue
		}
		data, err := builtinPresets.ReadFile("presets/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read embedded preset %s: %w", e.Name(), err)
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		p, err := parsePreset(data, name)
		if err != nil {
			return nil, fmt.Errorf("built-in preset %s: %w", e.Name(), err)
		}
		if err := p.validate(); err != nil {
			return nil, fmt.Errorf("built-in preset %s: %w", e.Name(), err)
		}
		presets[p.Name] = p
	}

	// 2. User presets (XDG), then 3. project-local presets.
	for _, dir := range []string{userPresetDir(), projectPresetDir()} {
		if dir == "" {
			continue
		}
		if err := loadPresetDir(dir, presets); err != nil {
			return nil, err
		}
	}

	return presets, nil
}

// loadPresetDir loads every *.json preset in dir into presets. A missing
// directory is not an error. Malformed files report the offending path.
func loadPresetDir(dir string, presets map[string]preset) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read preset dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasPrefix(e.Name(), "._") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read preset %s: %w", path, err)
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		p, err := parsePreset(data, name)
		if err != nil {
			return fmt.Errorf("preset %s: %w", path, err)
		}
		if err := p.validate(); err != nil {
			return fmt.Errorf("preset %s: %w", path, err)
		}
		presets[p.Name] = p
	}
	return nil
}

// loadPresetFile loads a single preset directly from a file path, used when
// --preset points at a .json file rather than a named preset.
func loadPresetFile(path string) (preset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return preset{}, fmt.Errorf("read preset %s: %w", path, err)
	}
	name := strings.TrimSuffix(filepath.Base(path), ".json")
	p, err := parsePreset(data, name)
	if err != nil {
		return preset{}, fmt.Errorf("preset %s: %w", path, err)
	}
	if err := p.validate(); err != nil {
		return preset{}, fmt.Errorf("preset %s: %w", path, err)
	}
	return p, nil
}

// sortedPresetNames returns preset names in alphabetical order.
func sortedPresetNames(presets map[string]preset) []string {
	names := make([]string, 0, len(presets))
	for k := range presets {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// presetNames returns a comma-separated, alphabetically sorted list of names.
func presetNames(presets map[string]preset) string {
	return strings.Join(sortedPresetNames(presets), ", ")
}

// looksLikePresetPath reports whether a --preset value should be treated as a
// file path rather than a named preset.
func looksLikePresetPath(v string) bool {
	return strings.HasSuffix(v, ".json") ||
		strings.ContainsRune(v, filepath.Separator) ||
		strings.HasPrefix(v, ".")
}

// printPresets writes a human-readable, aligned list of presets to stdout.
func printPresets(presets map[string]preset) {
	names := sortedPresetNames(presets)
	width := 0
	for _, n := range names {
		if len(n) > width {
			width = len(n)
		}
	}
	for _, n := range names {
		p := presets[n]
		fmt.Printf("  %-*s  %s\n", width, n, p.Description)
	}
}

// printPresetsJSON writes the presets as a JSON array to stdout, sorted by name.
func printPresetsJSON(presets map[string]preset) error {
	names := sortedPresetNames(presets)
	ordered := make([]preset, 0, len(names))
	for _, n := range names {
		ordered = append(ordered, presets[n])
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(ordered)
}
