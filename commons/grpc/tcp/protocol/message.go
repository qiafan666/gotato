package protocol

import (
	"bytes"
	"encoding/binary"
	"math"
)

var endian = binary.LittleEndian

const (
	tag         = uint32(0x70656562) // tag 固定值
	maxBodySize = math.MaxUint32
)

const (
	headerSize = 4 + 4 + 2 + 4 + 4 + 8 + 4 + 2
	tagLen     = 4
)

type header struct {
	MagicWord uint32
	Command   uint32
	PkgType   uint16
	Result    uint32
	Seq       uint32
	ReqId     int64
	BodySize  uint32
	ExtSize   uint16
}

type heartbeat struct {
	Type        uint16
	TimeoutSize uint16
	Timeout     uint32
}

type msg struct {
	header
	Ext  []byte
	Body []byte
}

func (m msg) Bytes() []byte {
	w := bytes.NewBuffer([]byte{})
	binary.Write(w, endian, m.header)
	if len(m.Ext) > 0 {
		w.Write(m.Ext)
	}
	if len(m.Body) > 0 {
		w.Write(m.Body)
	}
	return w.Bytes()
}
