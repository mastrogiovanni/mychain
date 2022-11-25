package stack

import "fmt"

type Direction bool

const (
	Up   Direction = true
	Down           = false
)

type Handler func(direction Direction, data []byte, layer *Layer)

type Layer struct {
	upper   *Layer
	lower   *Layer
	handler Handler
}

func NewLayer(handler Handler) *Layer {
	return &Layer{
		handler: handler,
	}
}

func (layer *Layer) Recv(direction Direction, data []byte) {
	if layer.handler != nil {
		go layer.handler(direction, data, layer)
	}
}

func (layer *Layer) SendUp(data []byte) {
	if layer.upper != nil {
		layer.upper.Recv(Up, data)
	}
}

func (layer *Layer) SendDown(data []byte) {
	if layer.lower != nil {
		layer.lower.Recv(Down, data)
	}
}

func (layer *Layer) AttachLower(lower *Layer) error {
	if layer.lower != nil {
		return fmt.Errorf("lower layer already existing")
	}
	if lower.upper != nil {
		return fmt.Errorf("upper layer of lower's one must be free")
	}
	layer.lower = lower
	lower.upper = layer
	return nil
}

func (layer *Layer) AttachUpper(upper *Layer) error {
	if layer.upper != nil {
		return fmt.Errorf("upper layer already connected")
	}
	if upper.lower != nil {
		return fmt.Errorf("lower layer of upper's one must be free")
	}
	layer.upper = upper
	upper.lower = layer
	return nil
}
