package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDumpAllWritesBuiltins(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := dumpAll(false); err != nil {
		t.Fatalf("dumpAll: %v", err)
	}

	presetsDir := filepath.Join(dir, "pixelize", "presets")
	palettesDir := filepath.Join(dir, "pixelize", "palettes")

	countJSON := func(d string) int {
		entries, err := os.ReadDir(d)
		if err != nil {
			t.Fatalf("read %s: %v", d, err)
		}
		n := 0
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".json" {
				n++
			}
		}
		return n
	}

	if got := countJSON(presetsDir); got != 16 {
		t.Errorf("expected 16 preset files, got %d", got)
	}
	if got := countJSON(palettesDir); got != 3 {
		t.Errorf("expected 3 palette files, got %d", got)
	}
}

func TestDumpAllSkipsExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Pre-create a file that dumpAll would otherwise write.
	pdir := filepath.Join(dir, "pixelize", "palettes")
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	custom := []byte(`{"name":"gameboy","colors":["#000000"]}`)
	target := filepath.Join(pdir, "gameboy.json")
	if err := os.WriteFile(target, custom, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := dumpAll(false); err != nil {
		t.Fatal(err)
	}
	// Without --force the existing custom file must be preserved.
	got, _ := os.ReadFile(target)
	if string(got) != string(custom) {
		t.Errorf("dumpAll overwrote an existing file without --force")
	}

	// With --force it should be replaced by the built-in.
	if err := dumpAll(true); err != nil {
		t.Fatal(err)
	}
	got2, _ := os.ReadFile(target)
	if string(got2) == string(custom) {
		t.Errorf("dumpAll --force did not overwrite the existing file")
	}
}
