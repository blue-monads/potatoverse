package xutils

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/nfnt/resize"
)

var ErrDecoderNotFound = errors.New("image decoder not found")

type imageProcessers struct {
	exts    []string
	decoder func(io.Reader) (image.Image, error)
}

var imgProcessers = []imageProcessers{
	{
		exts:    []string{".jpeg", ".jpg"},
		decoder: jpeg.Decode,
	},
	{
		exts:    []string{".png"},
		decoder: png.Decode,
	},
}

func GeneratePreview(input []byte, filename string) ([]byte, error) {
	var decoder func(io.Reader) (image.Image, error)

	for _, imgP := range imgProcessers {
		for _, ext := range imgP.exts {
			if strings.HasSuffix(filename, ext) {
				decoder = imgP.decoder
				break
			}
		}
	}

	if decoder == nil {
		return nil, ErrDecoderNotFound
	}

	img, err := decoder(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	m := resize.Resize(100, 0, img, resize.Lanczos3)

	outbuf := new(bytes.Buffer)

	err = jpeg.Encode(outbuf, m, nil)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(outbuf)
	return body, err
}
