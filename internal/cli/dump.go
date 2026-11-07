package cli

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// dumpAll writes the embedded built-in presets and palettes into the user
// config directory so they can be edited directly. Existing files are left
// untouched unless force is set. Files written here take precedence over the
// built-ins the next time pixelize runs.
func dumpAll(force bool) error {
	targets := []struct {
		fs     embed.FS
		embDir string
		outDir string
		label  string
	}{
		{builtinPresets, "presets", userPresetDir(), "preset"},
		{builtinPalettes, "palettes", userPaletteDir(), "palette"},
	}

	var wrote, skipped int
	for _, t := range targets {
		if t.outDir == "" {
			return fmt.Errorf("could not determine config directory for %ss", t.label)
		}
		if err := os.MkdirAll(t.outDir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", t.outDir, err)
		}
		entries, err := t.fs.ReadDir(t.embDir)
		if err != nil {
			return fmt.Errorf("read embedded %ss: %w", t.label, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasPrefix(e.Name(), "._") {
				continue
			}
			dst := filepath.Join(t.outDir, e.Name())
			if !force {
				if _, err := os.Stat(dst); err == nil {
					fmt.Fprintf(os.Stderr, "skip %s (exists)\n", dst)
					skipped++
					continue
				}
			}
			data, err := t.fs.ReadFile(t.embDir + "/" + e.Name())
			if err != nil {
				return fmt.Errorf("read embedded %s %s: %w", t.label, e.Name(), err)
			}
			if err := os.WriteFile(dst, data, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", dst, err)
			}
			fmt.Fprintf(os.Stderr, "wrote %s\n", dst)
			wrote++
		}
	}

	fmt.Fprintf(os.Stderr, "Dumped %d file(s), skipped %d.\n", wrote, skipped)
	if skipped > 0 && !force {
		fmt.Fprintln(os.Stderr, "Re-run with --force to overwrite existing files.")
	}
	return nil
}
