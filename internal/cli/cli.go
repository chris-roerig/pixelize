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

type preset struct {
	Block   int
	Colors  int
	Dither  bool
	Sharpen bool
	Mode    string
	Palette string
}

var presets = map[string]preset{
	"retro":   {Block: 8, Colors: 32, Dither: true, Mode: "average", Palette: "original"},
	"gameboy": {Block: 6, Colors: 4, Dither: true, Mode: "average", Palette: "gameboy"},
	"pico8":   {Block: 6, Colors: 16, Dither: true, Mode: "average", Palette: "pico8"},
	"nes":     {Block: 6, Colors: 24, Dither: true, Mode: "average", Palette: "nes"},
	"snes":    {Block: 5, Colors: 64, Dither: true, Mode: "average", Palette: "original"},
	"genesis": {Block: 6, Colors: 32, Dither: false, Mode: "dominant", Palette: "original"},
	"gba":     {Block: 4, Colors: 64, Dither: true, Mode: "average", Palette: "original"},
	"n64":     {Block: 5, Colors: 32, Dither: false, Mode: "average", Palette: "original"},
	"ps1":     {Block: 5, Colors: 48, Dither: true, Mode: "median", Palette: "original"},
	"c64":     {Block: 8, Colors: 16, Dither: true, Mode: "average", Palette: "original"},
	"cga":     {Block: 8, Colors: 4, Dither: true, Mode: "average", Palette: "original"},
	"sprite":  {Block: 4, Colors: 0, Dither: false, Sharpen: true, Mode: "dominant", Palette: "original"},
	"poster":  {Block: 12, Colors: 8, Dither: false, Mode: "average", Palette: "original"},
	"chunky":  {Block: 16, Colors: 16, Dither: false, Mode: "average", Palette: "original"},
	"fine":    {Block: 3, Colors: 64, Dither: true, Mode: "median", Palette: "original"},
	"mono":    {Block: 6, Colors: 2, Dither: true, Mode: "average", Palette: "gameboy"},
}

func presetNames() string {
	names := make([]string, 0, len(presets))
	for k := range presets {
		names = append(names, k)
	}
	return strings.Join(names, ", ")
}

// Run parses flags and executes the pixelize pipeline. Returns an error on failure.
func Run() error {
	block := flag.Int("block", 8, "pixel block size")
	scale := flag.Int("scale", 0, "nearest-neighbor output scale (default: match input size)")
	colors := flag.Int("colors", 0, "reduce palette to N colors (0 = unlimited)")
	dither := flag.Bool("dither", false, "apply Floyd-Steinberg dithering")
	bw := flag.Bool("bw", false, "reduce to black and white")
	transparent := flag.Bool("transparent", false, "make white pixels transparent")
	sharpen := flag.Bool("sharpen", false, "apply edge-enhancement sharpening")
	mode := flag.String("mode", "average", "sampling mode (average, median, dominant)")
	palette := flag.String("palette", "original", "color palette (original, pico8, gameboy, nes)")
	presetFlag := flag.String("preset", "", "apply a named preset ("+presetNames()+")")
	logical := flag.Bool("logical", false, "export logical pixel grid without scaling")
	showVersion := flag.Bool("version", false, "print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: pixelize [flags] input output
       pixelize [flags] directory
       pixelize --preset all input

Flags:
      --block int        pixel block size (default 8)
      --scale int        nearest-neighbor output scale (default: match input size)
      --colors int       reduce palette to N colors (0 = unlimited)
      --dither           apply Floyd-Steinberg dithering (auto 32 colors if --colors not set)
      --bw               reduce to black and white (2 colors, dithered)
      --sharpen          apply edge-enhancement sharpening
      --mode string      sampling mode (default "average")
      --palette string   color palette (default "original")
      --preset string    apply a named preset (use "all" to render every preset)
      --logical          export logical pixel grid without scaling
      --version          print version and exit

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
      retro              classic pixel art (block 8, 32 colors, dithered)
      gameboy            Game Boy (block 6, 4 colors, gameboy palette, dithered)
      pico8              PICO-8 (block 6, 16 colors, pico8 palette, dithered)
      nes                NES (block 6, 24 colors, nes palette, dithered)
      snes               SNES (block 5, 64 colors, dithered)
      genesis            Sega Genesis (block 6, 32 colors, dominant mode)
      gba                Game Boy Advance (block 4, 64 colors, dithered)
      n64                Nintendo 64 (block 5, 32 colors)
      ps1                PlayStation 1 (block 5, 48 colors, dithered, median)
      c64                Commodore 64 (block 8, 16 colors, dithered)
      cga                CGA (block 8, 4 colors, dithered)
      sprite             crisp sprites (block 4, dominant mode, sharpened)
      poster             poster style (block 12, 8 colors)
      chunky             big bold pixels (block 16, 16 colors)
      fine               subtle pixelation (block 3, 64 colors, median, dithered)
      mono               monochrome (block 6, 2 colors, gameboy palette, dithered)
      all                render input using every preset

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

	// Apply preset as base, then let explicit flags override
	if *presetFlag != "" && *presetFlag != "all" {
		p, ok := presets[*presetFlag]
		if !ok {
			return fmt.Errorf("unknown preset %q (available: %s, all)", *presetFlag, presetNames())
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

		for name, p := range presets {
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
