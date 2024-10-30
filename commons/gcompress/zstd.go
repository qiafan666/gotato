package gcompress

import (
	"github.com/klauspost/compress/zstd"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gerr"
	"io"
	"sync"
)

const WindowSize = 8 << 10 // 8kiB

var EncoderLevel = zstd.SpeedFastest

var zstdEncoderPool = &sync.Pool{
	New: func() any {
		encoder, err := newEncoder(nil, zstd.WithEncoderLevel(EncoderLevel))
		if err != nil {
			panic(err)
		}
		return encoder
	},
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
	decoder, err := newDecoder(nil)
	if err != nil {
		return nil, gerr.WrapMsg(err, "failed to create zstd decoder", "param", gcast.ToString(in))
	}
	all, err := decoder.DecodeAll(in, nil)
	if err != nil {
		return nil, gerr.WrapMsg(err, "failed to decode zstd data", "param", gcast.ToString(in))
	}
	return all, nil
}

func newEncoder(w io.Writer, options ...zstd.EOption) (*zstd.Encoder, error) {
	defaults := []zstd.EOption{
		zstd.WithEncoderConcurrency(1),
		zstd.WithWindowSize(WindowSize),
		zstd.WithZeroFrames(true),
	}
	return zstd.NewWriter(w, append(defaults, options...)...)
}

func newDecoder(r io.Reader, options ...zstd.DOption) (*zstd.Decoder, error) {
	defaults := []zstd.DOption{

		zstd.WithDecoderConcurrency(1),
		zstd.WithDecoderLowmem(true),
	}

	return zstd.NewReader(r, append(defaults, options...)...)
}
