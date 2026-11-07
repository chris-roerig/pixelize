package imageio

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// jpegQuality is the quality used when encoding JPEG output.
const jpegQuality = 90

// Load reads and decodes an image from path (PNG, JPEG, or generic fallback).
// A path of "-" reads from standard input.
func Load(path string) (image.Image, error) {
	if path == "-" {
		img, _, err := image.Decode(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("decode stdin: %w", err)
		}
		return img, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open input: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return png.Decode(f)
	case ".jpg", ".jpeg":
		return jpeg.Decode(f)
	default:
		img, _, err := image.Decode(f)
		return img, err
	}
}

// Save encodes img and writes it to path, choosing the format from the file
// extension (PNG by default, JPEG for .jpg/.jpeg). A path of "-" writes PNG to
// standard output.
func Save(path string, img image.Image) error {
	if path == "-" {
		return encode(os.Stdout, ".png", img)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	return encode(f, ext, img)
}

// SavePNG encodes img as PNG and writes it to path. Retained for compatibility;
// Save is preferred.
func SavePNG(path string, img image.Image) error {
	if path == "-" {
		return encode(os.Stdout, ".png", img)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()
	return encode(f, ".png", img)
}

func encode(w io.Writer, ext string, img image.Image) error {
	switch ext {
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(w, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return fmt.Errorf("encode jpeg: %w", err)
		}
	default:
		if err := png.Encode(w, img); err != nil {
			return fmt.Errorf("encode png: %w", err)
		}
	}
	return nil
}
