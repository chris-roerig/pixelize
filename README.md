# pixelize

A fast, minimal CLI tool that converts images into pixel art using configurable block sampling and nearest-neighbor scaling.

## Install

```bash
go install github.com/chris-roerig/pixelize/cmd/pixelize@latest
```

Or build from source:

```bash
git clone https://github.com/chris-roerig/pixelize.git
cd pixelize
go build -o pixelize ./cmd/pixelize
```

## Usage

Flags can appear before, after, or between positional arguments:

```bash
pixelize input.png output.png --block 8 --scale 4
pixelize --block 16 --mode average input.jpg output.png
pixelize input.png --logical output.png
pixelize --version
pixelize --help
```

### Batch Mode

Pass a directory to pixelize all supported images (`.png`, `.jpg`, `.jpeg`) inside it. Outputs are written to a `pixelized/` subfolder with the same filenames:

```bash
pixelize ./photos --block 8
# Creates ./photos/pixelized/ with processed versions
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--block` | 8 | Pixel block size for sampling |
| `--scale` | match input | Nearest-neighbor output scale factor |
| `--colors` | 0 (unlimited) | Reduce palette to N colors |
| `--dither` | false | Apply Floyd-Steinberg dithering (requires --colors) |
| `--bw` | false | Reduce to black and white (dithered) |
| `--transparent` | false | Make white pixels transparent |
| `--mode` | average | Sampling mode |
| `--palette` | original | Color palette |
| `--sharpen` | false | Apply edge-enhancement sharpening |
| `--preset` | — | Apply a named preset (see below) |
| `--logical` | false | Export the logical pixel grid without scaling |
| `--version` | — | Print version and exit |

### Presets

Use `--preset` for quick styling. Explicit flags override preset values.

| Preset | Description |
|--------|-------------|
| `retro` | Classic pixel art (block 8, 32 colors, dithered) |
| `gameboy` | Game Boy style (block 6, 4 colors, gameboy palette, dithered) |
| `pico8` | PICO-8 style (block 6, 16 colors, pico8 palette, dithered) |
| `nes` | NES style (block 6, 24 colors, nes palette, dithered) |
| `snes` | SNES style (block 5, 64 colors, dithered) |
| `genesis` | Sega Genesis style (block 6, 32 colors, dominant mode) |
| `gba` | Game Boy Advance style (block 4, 64 colors, dithered) |
| `n64` | Nintendo 64 style (block 5, 32 colors) |
| `ps1` | PlayStation 1 style (block 5, 48 colors, dithered, median) |
| `c64` | Commodore 64 style (block 8, 16 colors, dithered) |
| `cga` | CGA style (block 8, 4 colors, dithered) |
| `sprite` | Crisp sprites (block 4, dominant mode, sharpened) |
| `poster` | Poster style (block 12, 8 colors) |
| `chunky` | Big bold pixels (block 16, 16 colors) |
| `fine` | Subtle pixelation (block 3, 64 colors, median mode, dithered) |
| `mono` | Monochrome (block 6, 2 colors, gameboy palette, dithered) |

## Examples

Generated with `pixelize --preset all`:

| retro | gameboy | pico8 | nes |
|-------|---------|-------|-----|
| ![retro](examples/icons-retro.png) | ![gameboy](examples/icons-gameboy.png) | ![pico8](examples/icons-pico8.png) | ![nes](examples/icons-nes.png) |

| snes | genesis | gba | ps1 |
|------|---------|-----|-----|
| ![snes](examples/icons-snes.png) | ![genesis](examples/icons-genesis.png) | ![gba](examples/icons-gba.png) | ![ps1](examples/icons-ps1.png) |

| n64 | c64 | cga | mono |
|-----|-----|-----|------|
| ![n64](examples/icons-n64.png) | ![c64](examples/icons-c64.png) | ![cga](examples/icons-cga.png) | ![mono](examples/icons-mono.png) |

| sprite | poster | chunky | fine |
|--------|--------|--------|------|
| ![sprite](examples/icons-sprite.png) | ![poster](examples/icons-poster.png) | ![chunky](examples/icons-chunky.png) | ![fine](examples/icons-fine.png) |

## How It Works

Given a 1024×1024 input with `--block 8`:

1. The image is divided into 8×8 blocks.
2. Each block is reduced to a single color (the average of all pixels in that block).
3. This produces a **logical** image of 128×128 pixels.

Output depends on flags:

- `--logical` → outputs the 128×128 logical grid directly.
- `--scale 4` → scales the logical grid to 512×512 using nearest-neighbor interpolation.

Partial blocks (when dimensions aren't evenly divisible) are still sampled correctly.

## Supported Modes

| Mode | Description |
|------|-------------|
| `average` | Average color of all pixels in the block |
| `median` | Median color per channel (less affected by outliers) |
| `dominant` | Most frequent color in the block (preserves hard edges) |

## Supported Palettes

| Palette | Description |
|---------|-------------|
| `original` | Use the computed colors as-is |
| `pico8` | PICO-8 16-color palette |
| `gameboy` | Game Boy 4-color green palette |
| `nes` | NES system palette |
