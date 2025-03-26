package gcompress

import (
	"bytes"
	"compress/gzip"
	"github.com/qiafan666/gotato/commons/gerr"
	"io"
	"sync"
)

var (
	gzipWriterPool = sync.Pool{New: func() any { return gzip.NewWriter(nil) }}
	gzipReaderPool = sync.Pool{New: func() any { return new(gzip.Reader) }}
)

type ICompressor interface {
	Compress(rawData []byte) ([]byte, error)
	CompressWithPool(rawData []byte) ([]byte, error)
	DeCompress(compressedData []byte) ([]byte, error)
	DecompressWithPool(compressedData []byte) ([]byte, error)
}

type gzipCompressor struct {
	compressProtocol string
}

func NewGzipCompressor() ICompressor {
	return &gzipCompressor{compressProtocol: "gzip"}
}

func (g *gzipCompressor) Compress(rawData []byte) ([]byte, error) {

	gzipBuffer := bytes.Buffer{}
	gz := gzip.NewWriter(&gzipBuffer)

	if _, err := gz.Write(rawData); err != nil {
		return nil, gerr.WrapMsg(err, "GzipCompressor.Compress: writing to gzip writer failed")
	}

	if err := gz.Close(); err != nil {
		return nil, gerr.WrapMsg(err, "GzipCompressor.Compress: closing gzip writer failed")
	}

	return gzipBuffer.Bytes(), nil
}

func (g *gzipCompressor) CompressWithPool(rawData []byte) ([]byte, error) {
	gz := gzipWriterPool.Get().(*gzip.Writer)
	defer gzipWriterPool.Put(gz)

	gzipBuffer := bytes.Buffer{}
	gz.Reset(&gzipBuffer)

	if _, err := gz.Write(rawData); err != nil {
		return nil, gerr.WrapMsg(err, "GzipCompressor.CompressWithPool: error writing data")
	}
	if err := gz.Close(); err != nil {
		return nil, gerr.WrapMsg(err, "GzipCompressor.CompressWithPool: error closing gzip writer")
	}
	return gzipBuffer.Bytes(), nil
}

func (g *gzipCompressor) DeCompress(compressedData []byte) ([]byte, error) {
	buff := bytes.NewBuffer(compressedData)
	reader, err := gzip.NewReader(buff)
	if err != nil {
		return nil, gerr.WrapMsg(err, "GzipCompressor.DeCompress: NewReader creation failed")
	}
	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, gerr.WrapMsg(err, "GzipCompressor.DeCompress: reading from gzip reader failed")
	}
	if err = reader.Close(); err != nil {
		return decompressedData, gerr.WrapMsg(err, "GzipCompressor.DeCompress: closing gzip reader failed")
	}
	return decompressedData, nil
}

func (g *gzipCompressor) DecompressWithPool(compressedData []byte) ([]byte, error) {
	reader := gzipReaderPool.Get().(*gzip.Reader)
	defer gzipReaderPool.Put(reader)

	err := reader.Reset(bytes.NewReader(compressedData))
	if err != nil {
		return nil, gerr.WrapMsg(err, "GzipCompressor.DecompressWithPool: resetting gzip reader failed")
	}

	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, gerr.WrapMsg(err, "GzipCompressor.DecompressWithPool: reading from pooled gzip reader failed")
	}
	if err = reader.Close(); err != nil {
		return decompressedData, gerr.WrapMsg(err, "GzipCompressor.DecompressWithPool: closing pooled gzip reader failed")
	}
	return decompressedData, nil
}
