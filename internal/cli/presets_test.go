package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBuiltinPresets(t *testing.T) {
	presets, err := loadPresets()
	if err != nil {
		t.Fatalf("loadPresets: %v", err)
	}
	// All 16 built-ins must be present and valid.
	want := []string{
		"retro", "gameboy", "pico8", "nes", "snes", "genesis", "gba", "n64",
		"ps1", "c64", "cga", "sprite", "poster", "chunky", "fine", "mono",
	}
	for _, name := range want {
		p, ok := presets[name]
		if !ok {
			t.Errorf("missing built-in preset %q", name)
			continue
		}
		if p.Description == "" {
			t.Errorf("preset %q has no description", name)
		}
		if err := p.validate(); err != nil {
			t.Errorf("built-in preset %q invalid: %v", name, err)
		}
	}
}

func TestParsePresetFallbackName(t *testing.T) {
	p, err := parsePreset([]byte(`{"block":8,"mode":"average","palette":"original"}`), "fromfile")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "fromfile" {
		t.Errorf("expected name to fall back to %q, got %q", "fromfile", p.Name)
	}
}

func TestParsePresetExplicitNameWins(t *testing.T) {
	p, err := parsePreset([]byte(`{"name":"explicit","block":8,"mode":"average","palette":"original"}`), "fromfile")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "explicit" {
		t.Errorf("expected explicit name to win, got %q", p.Name)
	}
}

func TestParsePresetRejectsUnknownField(t *testing.T) {
	_, err := parsePreset([]byte(`{"name":"x","colour":5}`), "x")
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestLoadPresetFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.json")
	content := `{"name":"custom","description":"d","block":10,"colors":12,"dither":true,"mode":"average","palette":"original"}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := loadPresetFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "custom" || p.Block != 10 || p.Colors != 12 || !p.Dither {
		t.Errorf("unexpected preset: %+v", p)
	}
}

func TestLoadPresetFileValidationFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{"name":"bad","block":8,"mode":"nope","palette":"original"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadPresetFile(path); err == nil {
		t.Fatal("expected validation error for bad mode")
	}
}

func TestProjectPresetOverridesBuiltin(t *testing.T) {
	// Run in a temp working dir containing a project preset that shadows a
	// built-in name.
	dir := t.TempDir()
	presetsDir := filepath.Join(dir, ".pixelize", "presets")
	if err := os.MkdirAll(presetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	override := `{"name":"retro","description":"overridden","block":99,"colors":2,"mode":"average","palette":"original"}`
	if err := os.WriteFile(filepath.Join(presetsDir, "retro.json"), []byte(override), 0o644); err != nil {
		t.Fatal(err)
	}

	// Isolate XDG so a developer's real config can't interfere.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg-empty"))

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	presets, err := loadPresets()
	if err != nil {
		t.Fatal(err)
	}
	p := presets["retro"]
	if p.Block != 99 || p.Description != "overridden" {
		t.Errorf("project preset did not override built-in: %+v", p)
	}
}

func TestLooksLikePresetPath(t *testing.T) {
	cases := map[string]bool{
		"gameboy":     false,
		"my.json":     true,
		"./my.json":   true,
		"dir/my.json": true,
		".pixelize/x": true,
		"all":         false,
	}
	for in, want := range cases {
		if got := looksLikePresetPath(in); got != want {
			t.Errorf("looksLikePresetPath(%q) = %v, want %v", in, got, want)
		}
	}
}
