package image

import (
	"bytes"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDerivatives(t *testing.T) {
	tests := []struct {
		name              string
		width             int
		height            int
		format            string
		previewMaxSide    int
		fullscreenMaxSide int
		wantPreviewWidth  int
		wantPreviewHeight int
		wantFullSame      bool
		wantFullWidth     int
		wantFullHeight    int
	}{
		{
			name:              "Large PNG",
			width:             400,
			height:            300,
			format:            formatPNG,
			previewMaxSide:    250,
			fullscreenMaxSide: 100,
			wantPreviewWidth:  250,
			wantPreviewHeight: 188,
			wantFullWidth:     100,
			wantFullHeight:    75,
		},
		{
			name:              "Wide JPEG",
			width:             320,
			height:            180,
			format:            formatJPEG,
			previewMaxSide:    250,
			fullscreenMaxSide: 80,
			wantPreviewWidth:  250,
			wantPreviewHeight: 141,
			wantFullWidth:     80,
			wantFullHeight:    45,
		},
		{
			name:              "Small JPEG is not upscaled",
			width:             300,
			height:            200,
			format:            formatJPEG,
			previewMaxSide:    250,
			fullscreenMaxSide: 2048,
			wantPreviewWidth:  250,
			wantPreviewHeight: 167,
			wantFullSame:      true,
		},
		{
			name:              "Small PNG is not upscaled",
			width:             100,
			height:            50,
			format:            formatPNG,
			previewMaxSide:    250,
			fullscreenMaxSide: 2048,
			wantPreviewWidth:  100,
			wantPreviewHeight: 50,
			wantFullSame:      true,
		},
		{
			name:              "Tall PNG",
			width:             50,
			height:            80,
			format:            formatPNG,
			previewMaxSide:    20,
			fullscreenMaxSide: 2048,
			wantPreviewWidth:  13,
			wantPreviewHeight: 20,
			wantFullSame:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := encodeImage(t, tt.width, tt.height, tt.format)

			src, _, err := image.Decode(bytes.NewReader(file))
			require.NoError(t, err)

			preview, err := previewJPEG(src, tt.previewMaxSide)
			require.NoError(t, err)

			previewCfg, previewFormat, err := image.DecodeConfig(bytes.NewReader(preview))
			require.NoError(t, err)

			assert.Equal(t, formatJPEG, previewFormat)
			assert.Equal(t, tt.wantPreviewWidth, previewCfg.Width)
			assert.Equal(t, tt.wantPreviewHeight, previewCfg.Height)
			assert.InDelta(t, imageRatio(tt.width, tt.height), imageRatio(previewCfg.Width, previewCfg.Height), 0.05)

			full, resized, err := fullscreenVariant(src, tt.format, tt.fullscreenMaxSide)
			require.NoError(t, err)

			if tt.wantFullSame {
				assert.False(t, resized)
				assert.Nil(t, full)

				return
			}

			assert.True(t, resized)
			require.NotNil(t, full)

			fullCfg, fullFormat, err := image.DecodeConfig(bytes.NewReader(full))
			require.NoError(t, err)

			assert.Equal(t, tt.format, fullFormat)
			assert.Equal(t, tt.wantFullWidth, fullCfg.Width)
			assert.Equal(t, tt.wantFullHeight, fullCfg.Height)
			assert.InDelta(t, imageRatio(tt.width, tt.height), imageRatio(fullCfg.Width, fullCfg.Height), 0.05)
		})
	}
}

func imageRatio(width, height int) float64 {
	return float64(width) / float64(height)
}
