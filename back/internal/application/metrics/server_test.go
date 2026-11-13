package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestServer_Stop(t *testing.T) {
	tests := []struct {
		name string
		s    *Server
	}{
		{
			name: "Nil httpServer does not panic",
			s:    &Server{log: zap.NewNop()},
		},
		{
			name: "Zero value Server does not panic",
			s:    &Server{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				tt.s.Stop()
			})
		})
	}
}
