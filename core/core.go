package core

import (
	"image"
	"bufio"
	"io"
	"github.com/oliamb/cutter"
)

type Config struct {
	Width, Height int
}

func Crop(rd io.Reader, c Config) (image.Image, error) {
	img, _, err := image.Decode(bufio.NewReader(rd))
	if err != nil {
		return nil, err
	}
	return cutter.Crop(img, cutter.Config{Width: c.Width, Height: c.Height, Mode: cutter.Centered, Options:cutter.Ratio})
}