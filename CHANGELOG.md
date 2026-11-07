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
- Preset files are validated on load; invalid or malformed presets report the
  offending file path.

### Changed

- `--help` no longer hardcodes the preset list; it points to `--list-presets`
  so user-defined presets are always reflected.
- `--preset all` now renders presets in a deterministic (alphabetical) order.

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
