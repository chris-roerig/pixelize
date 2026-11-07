package cli

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/chris-roerig/pixelize/internal/imageio"
	"github.com/chris-roerig/pixelize/internal/pixelizer"
	"github.com/chris-roerig/pixelize/internal/version"
)

// normalizeArgs converts single-dash long flags (e.g. -scale) to double-dash (--scale)
// so pflag can parse them correctly.
func normalizeArgs(args []string) []string {
	out := make([]string, len(args))
	for i, a := range args {
		if strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") && len(a) > 2 {
			out[i] = "-" + a
		} else {
			out[i] = a
		}
	}
	return out
}

var supportedExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true,
}

// Run parses flags and executes the pixelize pipeline. Returns an error on failure.
func Run() error {
	presets, err := loadPresets()
	if err != nil {
		return err
	}

	block := flag.Int("block", 8, "pixel block size")
	scale := flag.Int("scale", 0, "nearest-neighbor output scale (default: match input size)")
	colors := flag.Int("colors", 0, "reduce palette to N colors (0 = unlimited)")
	dither := flag.Bool("dither", false, "apply Floyd-Steinberg dithering")
	bw := flag.Bool("bw", false, "reduce to black and white")
	transparent := flag.Bool("transparent", false, "make white pixels transparent")
	sharpen := flag.Bool("sharpen", false, "apply edge-enhancement sharpening")
	mode := flag.String("mode", "average", "sampling mode (average, median, dominant)")
	palette := flag.String("palette", "original", "color palette (original, pico8, gameboy, nes)")
	presetFlag := flag.String("preset", "", "apply a named preset ("+presetNames(presets)+")")
	logical := flag.Bool("logical", false, "export logical pixel grid without scaling")
	listPresets := flag.Bool("list-presets", false, "list available presets and exit")
	listPresetsJSON := flag.Bool("list-presets-json", false, "list available presets as JSON and exit")
	showVersion := flag.Bool("version", false, "print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: pixelize [flags] input output
       pixelize [flags] directory
       pixelize --preset all input

Flags:
      --block int          pixel block size (default 8)
      --scale int          nearest-neighbor output scale (default: match input size)
      --colors int         reduce palette to N colors (0 = unlimited)
      --dither             apply Floyd-Steinberg dithering (auto 32 colors if --colors not set)
      --bw                 reduce to black and white (2 colors, dithered)
      --transparent        make white pixels transparent
      --sharpen            apply edge-enhancement sharpening
      --mode string        sampling mode (default "average")
      --palette string     color palette (default "original")
      --preset string      apply a named preset, a path to a .json preset, or "all"
      --list-presets       list available presets and exit
      --list-presets-json  list available presets as JSON and exit
      --logical            export logical pixel grid without scaling
      --version            print version and exit

Modes:
      average            average color of all pixels in the block
      median             median color per channel (less affected by outliers)
      dominant           most frequent color in the block (preserves hard edges)

Palettes:
      original           use the computed colors as-is
      pico8              PICO-8 16-color palette
      gameboy            Game Boy 4-color green palette
      nes                NES system palette

Presets:
      Run "pixelize --list-presets" to see all available presets.
      Built-in presets ship with pixelize; add your own JSON files in
      $XDG_CONFIG_HOME/pixelize/presets/ or ./.pixelize/presets/.
      Use --preset all to render an input with every preset.

Examples:
      pixelize input.png output.png --block 8 --scale 4
      pixelize input.png output.png --preset ps1
      pixelize input.png output.png --dither --colors 16
      pixelize --preset all input.png
      pixelize ./photos --block 8
`)
	}

	os.Args = append(os.Args[:1], normalizeArgs(os.Args[1:])...)
	flag.Parse()

	if *showVersion {
		fmt.Println("pixelize " + version.Version)
		return nil
	}

	if *listPresets {
		printPresets(presets)
		return nil
	}

	if *listPresetsJSON {
		return printPresetsJSON(presets)
	}

	// Apply preset as base, then let explicit flags override
	if *presetFlag != "" && *presetFlag != "all" {
		p, ok := presets[*presetFlag]
		if !ok {
			// Fall back to treating the value as a path to a .json preset.
			if looksLikePresetPath(*presetFlag) {
				loaded, err := loadPresetFile(*presetFlag)
				if err != nil {
					return err
				}
				p = loaded
			} else {
				return fmt.Errorf("unknown preset %q (available: %s, all)", *presetFlag, presetNames(presets))
			}
		}
		if !flag.CommandLine.Changed("block") {
			*block = p.Block
		}
		if !flag.CommandLine.Changed("colors") {
			*colors = p.Colors
		}
		if !flag.CommandLine.Changed("dither") {
			*dither = p.Dither
		}
		if !flag.CommandLine.Changed("sharpen") {
			*sharpen = p.Sharpen
		}
		if !flag.CommandLine.Changed("mode") {
			*mode = p.Mode
		}
		if !flag.CommandLine.Changed("palette") {
			*palette = p.Palette
		}
	}

	// If --dither is set but no --colors, default to 32
	if *dither && *colors == 0 {
		*colors = 32
	}

	// --bw forces black and white with dithering
	if *bw {
		*colors = 0
		*dither = true
	}

	args := flag.Args()

	// Handle --preset all: render every preset
	if *presetFlag == "all" {
		if len(args) < 1 {
			return errors.New("usage: pixelize --preset all input.png")
		}
		input := args[0]
		ext := filepath.Ext(input)
		base := strings.TrimSuffix(input, ext)

		src, err := imageio.Load(input)
		if err != nil {
			return err
		}

		for _, name := range sortedPresetNames(presets) {
			p := presets[name]
			cfg := pixelizer.Config{
				BlockSize: p.Block,
				Scale:     p.Block,
				Colors:    p.Colors,
				Dither:    p.Dither,
				Sharpen:   p.Sharpen,
				Mode:      pixelizer.SamplingMode(p.Mode),
				Palette:   pixelizer.Palette(p.Palette),
			}
			if *scale != 0 {
				cfg.Scale = *scale
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("preset %s: %w", name, err)
			}
			out, err := pixelizer.Pixelize(src, cfg)
			if err != nil {
				return fmt.Errorf("preset %s: %w", name, err)
			}
			output := fmt.Sprintf("%s-%s.png", base, name)
			if err := imageio.SavePNG(output, out); err != nil {
				return fmt.Errorf("preset %s: %w", name, err)
			}
			fmt.Fprintf(os.Stderr, "Wrote %s\n", output)
		}
		fmt.Fprintf(os.Stderr, "Rendered %d presets\n", len(presets))
		return nil
	}

	scaleVal := *scale
	if scaleVal == 0 {
		scaleVal = *block
	}

	var fixedPal []color.RGBA
	if *bw {
		fixedPal = []color.RGBA{{0, 0, 0, 255}, {255, 255, 255, 255}}
	}

	cfg := pixelizer.Config{
		BlockSize:        *block,
		Scale:            scaleVal,
		Colors:           *colors,
		Dither:           *dither,
		Sharpen:          *sharpen,
		Mode:             pixelizer.SamplingMode(*mode),
		Palette:          pixelizer.Palette(*palette),
		Logical:          *logical,
		FixedPalette:     fixedPal,
		TransparentWhite: *transparent,
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	// Batch mode: single arg that is a directory
	if len(args) == 1 {
		info, err := os.Stat(args[0])
		if err != nil {
			return err
		}
		if info.IsDir() {
			return runBatch(args[0], cfg)
		}
	}

	if len(args) < 2 {
		return errors.New("usage: pixelize input.png output.png [--block 8] [--scale 4]")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments (expected 2, got %d)", len(args))
	}

	return processFile(args[0], args[1], cfg)
}

func processFile(input, output string, cfg pixelizer.Config) error {
	src, err := imageio.Load(input)
	if err != nil {
		return err
	}

	out, err := pixelizer.Pixelize(src, cfg)
	if err != nil {
		return err
	}

	if err := imageio.SavePNG(output, out); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Wrote %s\n", output)
	return nil
}

func runBatch(dir string, cfg pixelizer.Config) error {
	outDir := filepath.Join(dir, "pixelized")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !supportedExts[ext] {
			continue
		}
		input := filepath.Join(dir, e.Name())
		output := filepath.Join(outDir, e.Name())
		if err := processFile(input, output, cfg); err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
		count++
	}

	if count == 0 {
		return errors.New("no supported images found in " + dir)
	}
	fmt.Fprintf(os.Stderr, "Processed %d image(s)\n", count)
	return nil
}
