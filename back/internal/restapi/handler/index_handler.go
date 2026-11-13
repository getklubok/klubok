package handler

import (
	"context"

	"klubok/internal/restapi/oas"
)

type Index struct{}

func NewIndex() *Index {
	return &Index{}
}

// Get index.
//
// GET /.
func (*Index) Index(_ context.Context) (oas.IndexRes, error) { //nolint:unparam // It's ok
	return &oas.IndexResponse{
		Name: "Klubok API",
	}, nil
}
