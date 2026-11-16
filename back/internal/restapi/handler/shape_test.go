package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
