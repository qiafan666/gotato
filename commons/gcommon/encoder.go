package gcommon

import (
	"bytes"
	"encoding/gob"
	"github.com/qiafan666/gotato/commons/gerr"
)

type IEncoder interface {
	Encode(data any) ([]byte, error)
	Decode(encodeData []byte, decodeData any) error
}

type gobEncoder struct{}

func NewGobEncoder() IEncoder {
	return &gobEncoder{}
}

func (g *gobEncoder) Encode(data any) ([]byte, error) {
	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(data); err != nil {
		return nil, gerr.WrapMsg(err, "GobEncoder.Encode failed")
	}
	return buff.Bytes(), nil
}

func (g *gobEncoder) Decode(encodeData []byte, decodeData any) error {
	buff := bytes.NewBuffer(encodeData)
	dec := gob.NewDecoder(buff)
	if err := dec.Decode(decodeData); err != nil {
		return gerr.WrapMsg(err, "GobEncoder.Decode failed")
	}
	return nil
}
