package canvas

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
)

func MergeImage(a []byte, b []byte) ([]byte, error) {
	ai, _, err := image.Decode(bytes.NewReader(a))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	bi, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	rect := image.Rectangle{image.Point{0, 0}, ai.Bounds().Size()}
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, ai, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, rect, bi, image.Point{0, 0}, draw.Over)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, rgba)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return buf.Bytes(), nil
}
