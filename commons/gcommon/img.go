package gcommon

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"
)

const (
	formatPNG  = ".png"
	formatJPG  = ".jpg"
	formatJPEG = ".jpeg"
	formatGIF  = ".gif"
)

var (
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidImgUrl = errors.New("invalid img url")
)

var imgMap = map[string]Image{
	formatPNG:  &pngStruct{},
	formatJPG:  &jpgStruct{},
	formatJPEG: &jpgStruct{},
	formatGIF:  &gifStruct{},
}

type Image interface {
	Decode(reader io.Reader) (image.Image, error)
}

type pngStruct struct {
}

func (pngStruct) Decode(reader io.Reader) (image.Image, error) {
	return png.Decode(reader)
}

type jpgStruct struct {
}

func (jpgStruct) Decode(reader io.Reader) (image.Image, error) {
	return jpeg.Decode(reader)
}

type gifStruct struct {
}

func (gifStruct) Decode(reader io.Reader) (image.Image, error) {
	return gif.Decode(reader)
}

func DecodeImg(imgUrl string) (image.Image, error) {
	formatIndex := strings.LastIndex(imgUrl, ".")
	if formatIndex < 0 {
		return nil, ErrInvalidFormat
	}
	format := imgUrl[formatIndex:]
	i, ok := imgMap[format]
	if !ok {
		return nil, ErrInvalidFormat
	}

	rsp, err := http.Get(imgUrl)
	if rsp == nil || err != nil {
		return nil, ErrInvalidImgUrl
	}
	defer rsp.Body.Close()

	return i.Decode(rsp.Body)
}
