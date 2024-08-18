package gcommon

import (
	"bytes"
	"encoding/gob"
	"github.com/pkg/errors"
)

type Encoder interface {
	Encode(data any) ([]byte, error)
	Decode(encodeData []byte, decodeData any) error
}

type GobEncoder struct{}

func NewGobEncoder() *GobEncoder {
	return &GobEncoder{}
}

func (g *GobEncoder) Encode(data any) ([]byte, error) {
	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(data); err != nil {
		return nil, errors.Wrap(err, "GobEncoder.Encode failed")
	}
	return buff.Bytes(), nil
}

func (g *GobEncoder) Decode(encodeData []byte, decodeData any) error {
	buff := bytes.NewBuffer(encodeData)
	dec := gob.NewDecoder(buff)
	if err := dec.Decode(decodeData); err != nil {
		return errors.Wrap(err, "GobEncoder.Decode failed")
	}
	return nil
}
