# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Presets are now JSON files instead of hardcoded values. Built-in presets ship
  embedded in the binary and can be overridden or extended without recompiling.
- Custom preset discovery from `$XDG_CONFIG_HOME/pixelize/presets/` (user) and
  `./.pixelize/presets/` (project), plus `--preset ./path.json` to load a preset
  file directly.
- `--list-presets` and `--list-presets-json` to list available presets.
- `--list-palettes` to list available palettes.
- `--palette-out` extracts the palette an output uses to a reusable JSON file.
- `--grid` draws a subtle pixel-grid / CRT overlay on scaled output.
- Palettes are now JSON files (name + hex colors) instead of hardcoded values.
  Built-in palettes ship embedded and can be overridden or extended without
  recompiling, via `$XDG_CONFIG_HOME/pixelize/palettes/`,
  `./.pixelize/palettes/`, or `--palette ./file.json`.
- Preset files are validated on load; invalid or malformed presets report the
  offending file path.
- Output path is now optional: with a single input, pixelize infers the output
  name (`photo.png` -> `photo-pixelized.png`, or `photo-<preset>.png`).
- Standard input/output support via `-` (e.g. `cat in.png | pixelize - -`).
- JPEG output when the output path ends in `.jpg`/`.jpeg`.
- Batch mode now reports per-file progress (`[n/total]`).
- `curl | sh` installer (`install.sh`) that downloads prebuilt release binaries.
- GoReleaser configuration and GitHub Actions workflows for CI and tagged
  releases.

### Changed

- Block color averaging now happens in linear light instead of gamma-encoded
  sRGB, fixing a systematic darkening bias (output is noticeably less muddy).
- Nearest-color matching (built-in palettes and `--colors`) now uses a
  perceptual distance (redmean) instead of raw Euclidean RGB distance.
- `dominant` mode returns a representative color for the winning region instead
  of a bit-masked key, removing a banding floor.
- The `gameboy` palette now uses a properly spaced 4-shade DMG green ramp, and
  the `gameboy` and `sprite` presets were retuned for better results.
- Color reduction now applies exactly one strategy with clear precedence
  (`--bw` > named palette > `--colors`); a named palette no longer double-reduces
  with `--colors`.
- Sharpening and white-keying now run before color reduction so results stay on
  the chosen palette.
- `--help` no longer hardcodes the preset list; it points to `--list-presets`
  so user-defined presets are always reflected.
- `--preset all` now renders presets in a deterministic (alphabetical) order.
- Output encoding is chosen from the output file extension instead of always
  writing PNG.
- Regenerated all example images with the improved color pipeline.

### Fixed

- `averageColor` no longer double-applies alpha; partial-alpha blocks are
  averaged correctly instead of drifting toward black.
- `--dither` no longer drops transparency (alpha is carried through unchanged).
- `median`/`dominant` sampling now un-premultiply source pixels correctly.
- Guarded nearest-color matching against an empty palette.

## [0.1.0] - 2026-11-07

Initial tagged release. Establishes the baseline feature set.

### Added

- CLI that converts images to pixel art via configurable block sampling and
  nearest-neighbor scaling.
- Sampling modes: `average`, `median`, `dominant`.
- Color palettes: `original`, `pico8`, `gameboy`, `nes`.
- Color reduction via median-cut (`--colors`) with optional Floyd-Steinberg
  dithering (`--dither`).
- Black-and-white mode (`--bw`) and white-to-transparent (`--transparent`).
- Edge-enhancement sharpening (`--sharpen`).
- Logical-grid export without scaling (`--logical`).
- 16 built-in presets (retro, gameboy, pico8, nes, snes, genesis, gba, n64,
  ps1, c64, cga, sprite, poster, chunky, fine, mono).
- `--preset all` to render an input with every preset.
- Batch mode: pass a directory to process all supported images into a
  `pixelized/` subfolder.
- PNG and JPEG input; PNG output.
- Build-time version injection via `-ldflags`.

[Unreleased]: https://github.com/chris-roerig/pixelize/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/chris-roerig/pixelize/releases/tag/v0.1.0
