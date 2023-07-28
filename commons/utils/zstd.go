package utils

import (
	"github.com/klauspost/compress/zstd"
	"io"
	"sync"
)

const WindowSize = 8 << 10 // 8kiB

func NewDecoder(r io.Reader, options ...zstd.DOption) (*zstd.Decoder, error) {
	defaults := []zstd.DOption{

		zstd.WithDecoderConcurrency(1),
		zstd.WithDecoderLowmem(true),
	}

	return zstd.NewReader(r, append(defaults, options...)...)
}

func NewEncoder(w io.Writer, options ...zstd.EOption) (*zstd.Encoder, error) {
	defaults := []zstd.EOption{
		zstd.WithEncoderConcurrency(1),
		zstd.WithWindowSize(WindowSize),
		zstd.WithZeroFrames(true),
	}
	return zstd.NewWriter(w, append(defaults, options...)...)
}

func ZstdEncode(in []byte) []byte {
	encoder, ok := zstdEncoderPool.Get().(*zstd.Encoder)
	if !ok {
		panic("invalid type in sync pool")
	}
	out := encoder.EncodeAll(in, nil)
	_ = encoder.Close()
	zstdEncoderPool.Put(encoder)

	return out
}

func ZstdDecode(in []byte) ([]byte, error) {
	decoder, err := NewDecoder(nil)
	if err != nil {
		return nil, err
	}
	all, err := decoder.DecodeAll(in, nil)
	if err != nil {
		return nil, err
	}
	return all, nil
}

var zstdEncoderPool = &sync.Pool{
	New: func() any {
		encoder, err := NewEncoder(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
		if err != nil {
			panic(err)
		}
		return encoder
	},
}
