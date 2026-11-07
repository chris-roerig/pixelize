package cli

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestParseHexColor(t *testing.T) {
	cases := []struct {
		in   string
		want color.RGBA
		ok   bool
	}{
		{"#000000", color.RGBA{0, 0, 0, 255}, true},
		{"#ffffff", color.RGBA{255, 255, 255, 255}, true},
		{"88c070", color.RGBA{0x88, 0xc0, 0x70, 255}, true},
		{"#f00", color.RGBA{255, 0, 0, 255}, true},
		{"#1234", color.RGBA{}, false},
		{"#gggggg", color.RGBA{}, false},
		{"#11223344", color.RGBA{0x11, 0x22, 0x33, 0x44}, true},
	}
	for _, c := range cases {
		got, err := parseHexColor(c.in)
		if c.ok && err != nil {
			t.Errorf("parseHexColor(%q) unexpected error: %v", c.in, err)
			continue
		}
		if !c.ok {
			if err == nil {
				t.Errorf("parseHexColor(%q) expected error", c.in)
			}
			continue
		}
		if got != c.want {
			t.Errorf("parseHexColor(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestLoadBuiltinPalettes(t *testing.T) {
	palettes, err := loadPalettes()
	if err != nil {
		t.Fatalf("loadPalettes: %v", err)
	}
	for _, name := range []string{"gameboy", "pico8", "nes"} {
		p, ok := palettes[name]
		if !ok {
			t.Errorf("missing built-in palette %q", name)
			continue
		}
		if len(p.Colors) == 0 {
			t.Errorf("palette %q has no colors", name)
		}
	}
	if len(palettes["gameboy"].Colors) != 4 {
		t.Errorf("expected gameboy to have 4 colors, got %d", len(palettes["gameboy"].Colors))
	}
}

func TestResolvePaletteColorsOriginal(t *testing.T) {
	palettes, _ := loadPalettes()
	for _, name := range []string{"", "original"} {
		c, err := resolvePaletteColors(name, palettes)
		if err != nil {
			t.Fatalf("resolvePaletteColors(%q): %v", name, err)
		}
		if c != nil {
			t.Errorf("expected nil colors for %q, got %v", name, c)
		}
	}
}

func TestResolvePaletteColorsUnknown(t *testing.T) {
	palettes, _ := loadPalettes()
	if _, err := resolvePaletteColors("nope", palettes); err == nil {
		t.Fatal("expected error for unknown palette")
	}
}

func TestResolvePaletteColorsFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mine.json")
	if err := os.WriteFile(path, []byte(`{"name":"mine","colors":["#ff0000","#00ff00"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := resolvePaletteColors(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(c) != 2 || c[0] != (color.RGBA{255, 0, 0, 255}) {
		t.Errorf("unexpected colors: %v", c)
	}
}

func TestProjectPaletteOverridesBuiltin(t *testing.T) {
	dir := t.TempDir()
	pdir := filepath.Join(dir, ".pixelize", "palettes")
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pdir, "gameboy.json"),
		[]byte(`{"name":"gameboy","colors":["#111111","#222222"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg-empty"))

	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	palettes, err := loadPalettes()
	if err != nil {
		t.Fatal(err)
	}
	if len(palettes["gameboy"].Colors) != 2 {
		t.Errorf("project palette did not override built-in: %d colors", len(palettes["gameboy"].Colors))
	}
}
