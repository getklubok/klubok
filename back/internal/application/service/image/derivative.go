package image

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"

	xdraw "golang.org/x/image/draw"
)

const (
	previewJPEGQuality    = 85
	fullscreenJPEGQuality = 90
)

// previewJPEG уменьшает изображение и всегда кодирует его в JPEG.
func previewJPEG(src image.Image, maxSide int) ([]byte, error) {
	scaled := scaleTo(src, maxSide)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, scaled, &jpeg.Options{Quality: previewJPEGQuality}); err != nil {
		return nil, fmt.Errorf("encode preview: %w", err)
	}

	return buf.Bytes(), nil
}

// fullscreenVariant уменьшает изображение в формате оригинала.
// Если уменьшение не требуется, возвращает nil и false — вызывающий код использует байты оригинала.
func fullscreenVariant(src image.Image, format string, maxSide int) ([]byte, bool, error) {
	bounds := src.Bounds()
	if max(bounds.Dx(), bounds.Dy()) <= maxSide {
		return nil, false, nil
	}

	scaled := scaleTo(src, maxSide)

	var buf bytes.Buffer

	switch format {
	case formatPNG:
		if err := png.Encode(&buf, scaled); err != nil {
			return nil, false, fmt.Errorf("encode fullscreen png: %w", err)
		}
	case formatJPEG:
		if err := jpeg.Encode(&buf, scaled, &jpeg.Options{Quality: fullscreenJPEGQuality}); err != nil {
			return nil, false, fmt.Errorf("encode fullscreen jpeg: %w", err)
		}
	default:
		return nil, false, fmt.Errorf("unsupported image format %q", format)
	}

	return buf.Bytes(), true, nil
}

func scaleTo(src image.Image, maxSide int) image.Image {
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	dstWidth, dstHeight := fitSize(width, height, maxSide)
	if dstWidth == width && dstHeight == height {
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, xdraw.Src, nil)

	return dst
}

func fitSize(width, height, maxSide int) (dstWidth, dstHeight int) {
	longSide := max(width, height)
	if longSide <= maxSide {
		return width, height
	}

	dstWidth = max(int(math.Round(float64(width)*float64(maxSide)/float64(longSide))), 1)
	dstHeight = max(int(math.Round(float64(height)*float64(maxSide)/float64(longSide))), 1)

	return dstWidth, dstHeight
}
