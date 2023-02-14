package spi

import (
	"context"
)

type RenderingData struct {
	Name   string
	Values []float64
	Labels []string
}

type Renderer interface {
	Render(ctx context.Context, sink Sink, data []*RenderingData) error
}
