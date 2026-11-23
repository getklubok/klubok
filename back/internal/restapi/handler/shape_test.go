package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func TestDiagramShapeTextAlign(t *testing.T) {
	tests := []struct {
		name        string
		apiValue    oas.OptDiagramShapeTextAlignEnum
		wantAPI     oas.OptDiagramShapeTextAlignEnum
		domainValue domain.DiagramShapeTextAlign
	}{
		{
			name:        "default is center",
			apiValue:    oas.OptDiagramShapeTextAlignEnum{},
			wantAPI:     oas.NewOptDiagramShapeTextAlignEnum(oas.DiagramShapeTextAlignEnumCenter),
			domainValue: domain.DiagramShapeTextAlignCenter,
		},
		{
			name:        "left",
			apiValue:    oas.NewOptDiagramShapeTextAlignEnum(oas.DiagramShapeTextAlignEnumLeft),
			wantAPI:     oas.NewOptDiagramShapeTextAlignEnum(oas.DiagramShapeTextAlignEnumLeft),
			domainValue: domain.DiagramShapeTextAlignLeft,
		},
		{
			name:        "center",
			apiValue:    oas.NewOptDiagramShapeTextAlignEnum(oas.DiagramShapeTextAlignEnumCenter),
			wantAPI:     oas.NewOptDiagramShapeTextAlignEnum(oas.DiagramShapeTextAlignEnumCenter),
			domainValue: domain.DiagramShapeTextAlignCenter,
		},
		{
			name:        "right",
			apiValue:    oas.NewOptDiagramShapeTextAlignEnum(oas.DiagramShapeTextAlignEnumRight),
			wantAPI:     oas.NewOptDiagramShapeTextAlignEnum(oas.DiagramShapeTextAlignEnumRight),
			domainValue: domain.DiagramShapeTextAlignRight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.domainValue, diagramShapeTextAlignToDomain(tt.apiValue))
			assert.Equal(t, tt.wantAPI, oas.NewOptDiagramShapeTextAlignEnum(diagramShapeTextAlignFromDomain(tt.domainValue)))
		})
	}

	t.Run("empty domain value is center", func(t *testing.T) {
		assert.Equal(
			t,
			oas.NewOptDiagramShapeTextAlignEnum(oas.DiagramShapeTextAlignEnumCenter),
			oas.NewOptDiagramShapeTextAlignEnum(diagramShapeTextAlignFromDomain("")),
		)
	})
}

func TestDiagramShapeImage(t *testing.T) {
	const imageID = "01KM5CJP5ZHAGX8MJ0CN6988AF"

	t.Run("image shape round trip", func(t *testing.T) {
		api := oas.DiagramShape{
			ID:  oas.Ulid("shape-1"),
			Tp:  oas.DiagramShapeTypeEnumI,
			W:   100,
			H:   80,
			Img: oas.NewOptString(imageID),
		}

		shape, err := diagramShapeToDomain(api)
		require.NoError(t, err)
		assert.Equal(t, domain.DiagramShapeTypeImage, shape.Type)
		assert.Equal(t, imageID, shape.Img)
		assert.Empty(t, shape.Pv)

		got, err := diagramShapeFromDomain(shape)
		require.NoError(t, err)
		assert.Equal(t, oas.DiagramShapeTypeEnumI, got.Tp)
		assert.Equal(t, oas.NewOptString(imageID), got.Img)
		assert.False(t, got.Pv.Set)
	})

	t.Run("client pv is ignored on write", func(t *testing.T) {
		shape, err := diagramShapeToDomain(oas.DiagramShape{
			ID:  oas.Ulid("shape-1"),
			Tp:  oas.DiagramShapeTypeEnumI,
			Img: oas.NewOptString(imageID),
			Pv:  oas.NewOptString("data:image/jpeg;base64,AAAA"),
		})
		require.NoError(t, err)
		assert.Empty(t, shape.Pv)
	})

	t.Run("image shape gets placeholder text", func(t *testing.T) {
		shape, err := diagramShapeToDomain(oas.DiagramShape{
			ID:  oas.Ulid("shape-1"),
			Tp:  oas.DiagramShapeTypeEnumI,
			Img: oas.NewOptString(imageID),
		})
		require.NoError(t, err)
		assert.Equal(t, imageShapeText, shape.Text)
	})

	t.Run("explicit image text is kept", func(t *testing.T) {
		shape, err := diagramShapeToDomain(oas.DiagramShape{
			ID:  oas.Ulid("shape-1"),
			Tp:  oas.DiagramShapeTypeEnumI,
			Txt: "схема",
			Img: oas.NewOptString(imageID),
		})
		require.NoError(t, err)
		assert.Equal(t, "схема", shape.Text)
	})

	t.Run("other shape keeps empty text", func(t *testing.T) {
		shape, err := diagramShapeToDomain(oas.DiagramShape{
			ID: oas.Ulid("shape-1"),
			Tp: oas.DiagramShapeTypeEnumS,
		})
		require.NoError(t, err)
		assert.Empty(t, shape.Text)
	})

	t.Run("shape without image fields omits them", func(t *testing.T) {
		got, err := diagramShapeFromDomain(domain.DiagramShape{
			ID:   "shape-1",
			Type: domain.DiagramShapeTypeSquare,
		})
		require.NoError(t, err)
		assert.False(t, got.Img.Set)
		assert.False(t, got.Pv.Set)
	})

	t.Run("unknown type", func(t *testing.T) {
		_, err := diagramShapeToDomain(oas.DiagramShape{Tp: oas.DiagramShapeTypeEnum("X")})
		require.Error(t, err)

		_, err = diagramShapeFromDomain(domain.DiagramShape{Type: domain.DiagramShapeType("X")})
		require.Error(t, err)
	})
}

func TestValidateShape(t *testing.T) {
	const imageID = "01KM5CJP5ZHAGX8MJ0CN6988AF"

	tests := []struct {
		name    string
		shape   domain.DiagramShape
		wantErr bool
	}{
		{
			name:  "image shape with valid img",
			shape: domain.DiagramShape{Type: domain.DiagramShapeTypeImage, Img: imageID},
		},
		{
			name:    "image shape without img",
			shape:   domain.DiagramShape{Type: domain.DiagramShapeTypeImage},
			wantErr: true,
		},
		{
			name:    "image shape with invalid img",
			shape:   domain.DiagramShape{Type: domain.DiagramShapeTypeImage, Img: "not-an-ulid"},
			wantErr: true,
		},
		{
			name:  "square shape without img",
			shape: domain.DiagramShape{Type: domain.DiagramShapeTypeSquare},
		},
		{
			name:    "square shape with img",
			shape:   domain.DiagramShape{Type: domain.DiagramShapeTypeSquare, Img: imageID},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateShape(tt.shape)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidShape)

				return
			}

			require.NoError(t, err)
		})
	}
}
