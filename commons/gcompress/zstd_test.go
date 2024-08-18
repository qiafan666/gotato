package gcompress

import (
	"bytes"
	"fmt"
	"testing"
)

func TestZstd(t *testing.T) {

	//zstd
	data := []byte("testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest")
	fmt.Println(len(data)) //68
	//压缩
	compressData := ZstdEncode(data)
	fmt.Println(len(compressData)) //24

	//解压缩
	decoder, err := ZstdDecode(compressData)
	if err != nil {
		return
	}
	fmt.Println(bytes.NewBuffer(decoder).String())

}
