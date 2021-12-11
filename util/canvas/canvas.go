package canvas

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"strings"
)

func MergeImage(a string, b string) (string, error) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a))
	ai, _, err := image.Decode(reader)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}
	reader = base64.NewDecoder(base64.StdEncoding, strings.NewReader(b))
	bi, _, err := image.Decode(reader)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}
	rect := image.Rectangle{image.Point{0, 0}, ai.Bounds().Size()}
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, ai, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, rect, bi, image.Point{0, 0}, draw.Over)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, rgba)
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
