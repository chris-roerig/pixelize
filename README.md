# pixelize

A fast, minimal command-line tool that converts images into pixel art using
configurable block sampling, perceptually accurate color reduction, and a
catalog of retro console presets.

| Before | After |
|--------|-------|
| ![original](examples/cassette.png) | ![pixelized](examples/cassette-gba.png) |
| ![original](examples/castle.png) | ![pixelized](examples/castle-snes.png) |
| ![original](examples/game.png) | ![pixelized](examples/game-ps1.png) |

**Website:** https://chris-roerig.github.io/pixelize/

## Contents

- [Install](#install)
  - [Shell completions](#shell-completions)
- [Quick Start](#quick-start)
- [Command Reference](#command-reference)
  - [Flags](#flags)
  - [Commands](#commands)
- [Working With Files](#working-with-files)
  - [Output path inference](#output-path-inference)
  - [Output format](#output-format)
  - [Pipes (stdin / stdout)](#pipes-stdin--stdout)
  - [Batch mode](#batch-mode)
- [Presets](#presets)
- [Sampling Modes](#sampling-modes)
- [Palettes](#palettes)
- [Effects](#effects)
  - [Dithering](#dithering)
  - [Color reduction](#color-reduction)
  - [Transparency](#transparency)
  - [Sharpen](#sharpen)
  - [Pixel grid](#pixel-grid)
- [Customization](#customization)
  - [Custom presets](#custom-presets)
  - [Custom palettes](#custom-palettes)
  - [Editing the built-ins (`dump-all`)](#editing-the-built-ins-dump-all)
  - [Extracting a palette](#extracting-a-palette)
- [How It Works](#how-it-works)
- [Examples](#examples)
- [Contributing](#contributing)

## Install

**Quick install (Linux and macOS)** — downloads a prebuilt binary and verifies
its checksum:

```bash
curl -fsSL https://raw.githubusercontent.com/chris-roerig/pixelize/main/install.sh | sh
```

Install a specific version, or to a custom directory (note: `export` the
variable so the piped shell inherits it):

```bash
export PIXELIZE_VERSION=v1.1.0
curl -fsSL https://raw.githubusercontent.com/chris-roerig/pixelize/main/install.sh | sh

export PIXELIZE_BINDIR="$HOME/.local/bin"
curl -fsSL https://raw.githubusercontent.com/chris-roerig/pixelize/main/install.sh | sh
```

**With Go:**

```bash
go install github.com/chris-roerig/pixelize/cmd/pixelize@latest
```

**From source:**

```bash
git clone https://github.com/chris-roerig/pixelize.git
cd pixelize
go build -o pixelize ./cmd/pixelize
```

Verify the install:

```bash
pixelize --version
```

### Shell completions

The `curl | sh` installer sets up tab completion automatically for zsh, bash,
and fish (restart your shell afterward). To install manually, grab the files
from the [`completions/`](completions/) directory (or the release archive):

**zsh** — place `_pixelize` on your `$fpath` and reload:

```bash
mkdir -p ~/.zsh/completions
cp completions/_pixelize ~/.zsh/completions/_pixelize
# in ~/.zshrc, before `compinit`:
#   fpath=(~/.zsh/completions $fpath)
#   autoload -Uz compinit && compinit
```

**bash** — source it from `~/.bashrc`:

```bash
cp completions/pixelize.bash ~/.local/share/bash-completion/completions/pixelize
# or: source /path/to/completions/pixelize.bash
```

**fish** — drop it in your fish completions directory:

```bash
cp completions/pixelize.fish ~/.config/fish/completions/pixelize.fish
```

Completion covers all flags and the `dump-all` command, and completes
`--preset` and `--palette` values dynamically — including any custom presets or
palettes you've added.

## Quick Start

The simplest invocation takes one image; the output name is inferred:

```bash
pixelize photo.png            # -> photo-pixelized.png
```

Apply a retro console look with a preset:

```bash
pixelize photo.png --preset gba      # -> photo-gba.png
pixelize photo.png --preset gameboy  # -> photo-gameboy.png
```

Control the look directly with flags:

```bash
pixelize photo.png output.png --block 12 --colors 16 --dither
```

Process an entire folder:

```bash
pixelize ./photos --preset nes       # writes ./photos/pixelized/
```

See every available preset and palette:

```bash
pixelize --list-presets
pixelize --list-palettes
```

## Command Reference

```
pixelize [flags] input [output]
pixelize [flags] directory
pixelize --preset all input
pixelize dump-all [--force]
```

Flags may appear before, after, or between positional arguments. The `input`
may be a file, a directory (batch mode), or `-` for standard input. The
`output` is optional — see [output path inference](#output-path-inference).

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--block` | 8 | Pixel block size for sampling (larger = chunkier) |
| `--scale` | match input | Nearest-neighbor output scale factor |
| `--colors` | 0 (unlimited) | Reduce to N colors via median-cut |
| `--dither` | false | Apply Floyd-Steinberg dithering (auto 32 colors if `--colors` unset) |
| `--bw` | false | Reduce to black and white (dithered) |
| `--transparent` | false | Make white pixels transparent |
| `--mode` | average | Sampling mode: `average`, `median`, or `dominant` |
| `--palette` | original | Color palette, or a path to a `.json` palette |
| `--sharpen` | false | Apply edge-enhancement sharpening |
| `--grid` | false | Draw a subtle pixel-grid overlay on scaled output |
| `--preset` | — | Apply a named preset, a path to a `.json` preset, or `all` |
| `--palette-out` | — | Write the output's palette to a JSON file and exit |
| `--logical` | false | Export the logical pixel grid without scaling |
| `--list-presets` | — | List available presets and exit |
| `--list-presets-json` | — | List available presets as JSON and exit |
| `--list-palettes` | — | List available palettes and exit |
| `--force` | false | Overwrite existing files (used with `dump-all`) |
| `--version` | — | Print version and exit |
| `--help` | — | Print usage and exit |

### Commands

| Command | Description |
|---------|-------------|
| `dump-all` | Copy the built-in presets and palettes into your config directory for editing. See [Editing the built-ins](#editing-the-built-ins-dump-all). |

## Working With Files

### Output path inference

The `output` argument is optional. With a single input, pixelize derives the
output name automatically:

```bash
pixelize photo.png                 # -> photo-pixelized.png
pixelize photo.png --preset gba    # -> photo-gba.png
```

### Output format

The output format follows the output file's extension: `.png` (default) or
`.jpg` / `.jpeg`.

```bash
pixelize photo.png output.jpg --preset nes   # writes JPEG
```

### Pipes (stdin / stdout)

Use `-` to read from standard input and/or write to standard output, so
pixelize composes in shell pipelines:

```bash
cat input.png | pixelize - - --preset gba > output.png
curl -s https://example.com/img.png | pixelize - --preset nes > out.png
```

### Batch mode

Pass a directory to pixelize every supported image (`.png`, `.jpg`, `.jpeg`)
inside it. Outputs are written to a `pixelized/` subfolder with the same
filenames, and progress is reported per file:

```bash
pixelize ./photos --preset gba
# [1/12] a.png
# ...
# Processed 12 image(s) into ./photos/pixelized
```

## Presets

A preset bundles a set of flags into a named look. Use `--preset <name>`;
explicit flags still override individual preset values.

```bash
pixelize photo.png --preset ps1
pixelize photo.png --preset ps1 --block 10   # override just the block size
```

Render an input with **every** preset at once (writes `photo-<preset>.png` for
each):

```bash
pixelize photo.png --preset all
```

List all presets (built-in plus any you've added):

```bash
pixelize --list-presets
pixelize --list-presets-json   # machine-readable
```

| Preset | Description |
|--------|-------------|
| `retro` | Classic pixel art (block 8, 32 colors, dithered) |
| `gameboy` | Game Boy (block 5, gameboy palette, dithered) |
| `pico8` | PICO-8 (block 6, 16 colors, pico8 palette, dithered) |
| `nes` | NES (block 6, 24 colors, nes palette, dithered) |
| `snes` | SNES (block 5, 64 colors, dithered) |
| `genesis` | Sega Genesis (block 6, 32 colors, dominant mode) |
| `gba` | Game Boy Advance (block 4, 64 colors, dithered) |
| `n64` | Nintendo 64 (block 5, 32 colors) |
| `ps1` | PlayStation 1 (block 5, 48 colors, dithered, median) |
| `c64` | Commodore 64 (block 8, 16 colors, dithered) |
| `cga` | CGA (block 8, 4 colors, dithered) |
| `sprite` | Crisp sprites (block 8, 16 colors, dominant mode) |
| `poster` | Poster style (block 12, 8 colors) |
| `chunky` | Big bold pixels (block 16, 16 colors) |
| `fine` | Subtle pixelation (block 3, 64 colors, median, dithered) |
| `mono` | Monochrome (block 6, 2 colors, gameboy palette, dithered) |

To create your own, see [Custom presets](#custom-presets).

## Sampling Modes

`--mode` controls how each block is reduced to a single color.

| Mode | Description | Good for |
|------|-------------|----------|
| `average` | Average of all pixels in the block (computed in linear light) | Photos, smooth gradients |
| `median` | Median per channel — less affected by outliers | Noisy images |
| `dominant` | Representative color of the block's most common region | Hard edges, flat art, sprites |

```bash
pixelize photo.png --mode dominant --block 8
```

## Palettes

`--palette` maps the image onto a fixed set of colors. Palette matching uses a
perceptual distance metric, so colors snap to the nearest color the eye
expects.

List all palettes (built-in plus any you've added):

```bash
pixelize --list-palettes
```

| Palette | Description |
|---------|-------------|
| `original` | Use the computed colors as-is (no mapping) |
| `pico8` | PICO-8 16-color palette |
| `gameboy` | Game Boy 4-shade DMG green ramp |
| `nes` | NES system palette |

```bash
pixelize photo.png --palette gameboy
pixelize photo.png --palette ./my-palette.json   # a custom palette file
```

To create your own, see [Custom palettes](#custom-palettes).

## Effects

### Dithering

`--dither` applies Floyd-Steinberg error diffusion, trading solid color regions
for patterns that approximate colors the palette lacks. It pairs with color
reduction; if `--colors` is unset, `--dither` defaults to 32 colors.

```bash
pixelize photo.png --colors 16 --dither
```

### Color reduction

Reduce the palette to N colors with median-cut:

```bash
pixelize photo.png --colors 8
```

Only one color-reduction strategy applies at a time, in precedence order:
`--bw` > a named `--palette` > `--colors`. When a named palette is set,
`--colors` is ignored rather than applied as a confusing second reduction.

Black and white (2 colors, dithered):

```bash
pixelize photo.png --bw
```

### Transparency

`--transparent` makes near-white pixels transparent (useful for sprites on a
white background):

```bash
pixelize sprite.png --transparent
```

### Sharpen

`--sharpen` applies edge enhancement before color reduction, keeping edges
crisp:

```bash
pixelize photo.png --sharpen --preset sprite
```

### Pixel grid

`--grid` draws a subtle darkened grid between pixels on scaled output — a
graph-paper / CRT look that emphasizes the pixel structure:

```bash
pixelize photo.png output.png --block 16 --scale 16 --grid
```

## Customization

Presets and palettes are plain JSON files. pixelize ships built-ins embedded in
the binary, and you can add or override them without recompiling.

Both are loaded from these locations, in order of increasing precedence (a
later source overrides an earlier one with the same `name`):

1. **Built-in** — shipped inside the binary.
2. **User** — `$XDG_CONFIG_HOME/pixelize/` (defaults to `~/.config/pixelize/`).
3. **Project** — `./.pixelize/` relative to where you run the command.

You can also point `--preset` or `--palette` directly at a `.json` file.

### Custom presets

A preset is a named bundle of settings:

```json
{
  "name": "my-style",
  "description": "my custom look",
  "block": 10,
  "colors": 12,
  "dither": true,
  "sharpen": false,
  "mode": "average",
  "palette": "original"
}
```

Save it to `~/.config/pixelize/presets/my-style.json` and use it:

```bash
pixelize photo.png --preset my-style
pixelize photo.png --preset ./my-style.json   # or point at a file directly
```

Notes:

- The `name` field is the identifier you pass to `--preset`. If omitted, the
  file name (without `.json`) is used.
- Unknown fields are rejected, so typos like `"colour"` fail loudly with the
  offending file path.
- Each preset is validated on load; an invalid `mode`, `palette`, or `block`
  reports an error pointing at the file.

### Custom palettes

A palette is a name plus a list of hex colors (`#rgb`, `#rrggbb`, or
`#rrggbbaa`; the `#` is optional):

```json
{
  "name": "sunset",
  "description": "warm four-color sunset",
  "colors": ["#1a1a2e", "#e94560", "#f9ed69", "#f08a5d"]
}
```

Save it to `~/.config/pixelize/palettes/sunset.json` and use it:

```bash
pixelize photo.png --palette sunset
pixelize photo.png --palette ./sunset.json   # or point at a file directly
```

### Editing the built-ins (`dump-all`)

The built-in presets and palettes are embedded in the binary. To edit them,
copy them into your config directory with `dump-all`:

```bash
pixelize dump-all          # writes built-ins to ~/.config/pixelize/
pixelize dump-all --force  # overwrite existing copies
```

Existing files are left untouched unless `--force` is given. Edited copies in
the config directory override the built-ins the next time you run pixelize.

### Extracting a palette

`--palette-out` dumps the palette an output actually uses to a reusable JSON
file, so you can lift a palette from one image and apply it to others:

```bash
pixelize photo.png --palette-out mypalette.json --colors 8
pixelize other.png output.png --palette ./mypalette.json
```

## How It Works

Given a 1024×1024 input with `--block 8`:

1. The image is divided into 8×8 blocks.
2. Each block is reduced to a single color (by default, the average of all
   pixels in that block).
3. This produces a **logical** image of 128×128 pixels.
4. Colors are optionally reduced (a palette or `--colors`), then the grid is
   scaled back up.

Output depends on flags:

- `--logical` → outputs the 128×128 logical grid directly.
- `--scale 4` → scales the logical grid to 512×512 using nearest-neighbor
  interpolation.

Partial blocks (when dimensions aren't evenly divisible) are still sampled
correctly.

**Pipeline order:** block sampling → optional sharpen → optional white-keying →
exactly one color reduction → nearest-neighbor scaling → optional grid overlay.

**Color handling.** pixelize keeps output faithful to the source:

- **Linear-light averaging** — block colors are averaged in linear light rather
  than gamma-encoded sRGB, avoiding the darkening bias that makes naive
  pixelizers look muddy.
- **Perceptual color matching** — nearest-color decisions (palettes and
  `--colors`) use a perceptual distance (redmean) instead of raw RGB distance.
- **Alpha-correct** — transparency is handled consistently through sampling,
  dithering, and quantization.

## Examples

Generated with `pixelize cassette.png --preset all`:

| retro | gameboy | pico8 | nes |
|-------|---------|-------|-----|
| ![retro](examples/cassette-retro.png) | ![gameboy](examples/cassette-gameboy.png) | ![pico8](examples/cassette-pico8.png) | ![nes](examples/cassette-nes.png) |

| snes | genesis | gba | ps1 |
|------|---------|-----|-----|
| ![snes](examples/cassette-snes.png) | ![genesis](examples/cassette-genesis.png) | ![gba](examples/cassette-gba.png) | ![ps1](examples/cassette-ps1.png) |

| n64 | c64 | cga | mono |
|-----|-----|-----|------|
| ![n64](examples/cassette-n64.png) | ![c64](examples/cassette-c64.png) | ![cga](examples/cassette-cga.png) | ![mono](examples/cassette-mono.png) |

| sprite | poster | chunky | fine |
|--------|--------|--------|------|
| ![sprite](examples/cassette-sprite.png) | ![poster](examples/cassette-poster.png) | ![chunky](examples/cassette-chunky.png) | ![fine](examples/cassette-fine.png) |

## Contributing

To contribute a preset or palette upstream, add a JSON file to
`internal/cli/presets/` or `internal/cli/palettes/` and open a pull request —
no Go required. Run `pixelize dump-all` first to get editable copies of the
built-ins as a starting point.

For code changes, the standard Go workflow applies:

```bash
go build ./...
go test ./...
```

Releases follow [Semantic Versioning](https://semver.org/) and are documented
in [CHANGELOG.md](CHANGELOG.md). See [RELEASING.md](RELEASING.md) for the
release process.
