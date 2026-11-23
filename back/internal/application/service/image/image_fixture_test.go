package image

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func encodePNG(t *testing.T, width, height int) []byte {
	t.Helper()

	return encodeImage(t, width, height, "png")
}

func encodeJPEG(t *testing.T, width, height int) []byte {
	t.Helper()

	return encodeImage(t, width, height, "jpeg")
}

func encodeGIF(t *testing.T, width, height int) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, gif.Encode(&buf, testImage(width, height), nil))

	return buf.Bytes()
}

func encodeImage(t *testing.T, width, height int, format string) []byte {
	t.Helper()

	var buf bytes.Buffer

	switch format {
	case formatPNG:
		require.NoError(t, png.Encode(&buf, testImage(width, height)))
	case formatJPEG:
		require.NoError(t, jpeg.Encode(&buf, testImage(width, height), &jpeg.Options{Quality: 100}))
	default:
		t.Fatalf("unsupported test format %q", format)
	}

	return buf.Bytes()
}

func testImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := range img.Pix {
		img.Pix[i] = byte(i % 251)
	}

	return img
}

func decodeDataURL(t *testing.T, dataURL string) (image.Config, string) {
	t.Helper()

	require.True(t, strings.HasPrefix(dataURL, previewDataURLPrefix), "unexpected data URL prefix: %q", dataURL)

	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, previewDataURLPrefix))
	require.NoError(t, err)

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	require.NoError(t, err)

	return cfg, format
}
