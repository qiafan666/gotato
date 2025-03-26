package protocol

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/commons/grpc"
	"io"
)

type textRpcProtocol struct{}

func New() grpc.IProtocol {
	return &textRpcProtocol{}
}

func (t *textRpcProtocol) Encode(ctx context.Context, v *grpc.Message) ([]byte, error) {
	// 发送心跳包时,requestBody为心跳包内容
	if v.Command == grpc.CmdHeartbeat && v.PkgType == grpc.PkgTypeRequest && v.Heartbeat != nil {
		v.Body = t.packHeartbeatRequest(v.Heartbeat)
	}

	return t.encode(
		v.Command,
		v.PkgType,
		v.Result,
		v.Seq,
		v.ReqId,
		v.Body,
	)
}
func (t *textRpcProtocol) Decode(ctx context.Context, reader io.Reader) (*grpc.Message, error) {
	m, err := t.recv(ctx, reader)
	if err != nil {
		return nil, err
	}

	v := &grpc.Message{
		Command:   grpc.Command(m.Command),
		PkgType:   grpc.PkgType(m.PkgType),
		ReqId:     m.ReqId,
		Seq:       m.Seq,
		Result:    m.Seq,
		Body:      m.Body,
		Heartbeat: nil,
	}

	// 如果是发送心跳包,解析body,heartbeat设置
	if v.Command == grpc.CmdHeartbeat && v.PkgType == grpc.PkgTypeRequest {
		var heartRequest *grpc.Heartbeat
		heartRequest, err = t.unpackHeartbeatRequest(v.Body)
		if err != nil {
			return nil, err
		}
		v.Heartbeat = heartRequest
	}

	return v, nil
}

func (t *textRpcProtocol) recv(ctx context.Context, reader io.Reader) (*msg, error) {
	headerData := make([]byte, 0)
	nextReadSize := headerSize

	for {
		// 读取header
		buf, err := t.read(ctx, reader, uint32(nextReadSize))
		if err != nil {
			return nil, err
		}
		headerData = append(headerData, buf...)
		if len(headerData) < headerSize {
			return nil, gerr.New("header data too short", "headerSize", headerSize, "headerData", len(headerData))
		}

		pos := bytes.Index(headerData, endian.AppendUint32([]byte{}, tag))
		if pos < 0 {
			headerData = headerData[headerSize-tagLen:]
			nextReadSize = headerSize - len(headerData)
			continue
		}

		if pos > 0 {
			headerData = headerData[pos:]
			nextReadSize = headerSize - len(headerData)
			continue
		}

		break
	}

	// 读取header
	var newHeader header
	bufReader := bytes.NewReader(headerData)
	e := binary.Read(bufReader, endian, &newHeader)
	if e != nil {
		return nil, gerr.WrapMsg(e, "protocol unpack header error")
	}

	m := &msg{
		header: newHeader,
		Ext:    nil,
		Body:   nil,
	}

	// 读取ext
	if m.ExtSize > 0 {
		buf, err := t.read(ctx, reader, uint32(m.ExtSize))
		if err != nil {
			return nil, gerr.WrapMsg(e, "protocol extSize read error")
		}
		m.Ext = buf
	}

	// 读取body
	if m.BodySize > 0 {
		buf, err := t.read(ctx, reader, m.BodySize)
		if err != nil {
			return nil, gerr.WrapMsg(e, "protocol bodySize read error")
		}
		m.Body = buf
	}

	return m, nil
}

func (t *textRpcProtocol) read(ctx context.Context, reader io.Reader, length uint32) ([]byte, error) {
	bytesBuffer := bytes.NewBuffer([]byte{})

	totalLen := uint32(0) // 已读取的总长度

	// 循环读取,直到读够length
	for {
		if ctx.Err() != nil {
			return nil, gerr.WrapMsg(ctx.Err(), "protocol read context error")
		}

		size := 1024                    // 默认每次读取长度
		leftLength := length - totalLen // 剩余待读取长度

		// 已读满length
		if leftLength <= 0 {
			break
		}

		// 剩余待读取小于默认,最多读取不超过剩余长度的数据
		if size > int(leftLength) {
			size = int(leftLength)
		}
		buf := make([]byte, size)
		l, err := reader.Read(buf)
		if err != nil {
			return nil, gerr.WrapMsg(err, "protocol read buffer error")
		}
		if l == 0 {
			continue
		}
		totalLen += uint32(l)
		bytesBuffer.Write(buf[0:l])

		if totalLen >= length {
			break
		}
	}
	return bytesBuffer.Bytes(), nil
}

func (t *textRpcProtocol) encode(cmd grpc.Command, pkgType grpc.PkgType, result, Seq uint32, reqId int64, data []byte) ([]byte, error) {
	if len(data) > maxBodySize {
		return nil, gerr.New("data too long", "maxBodySize", maxBodySize, "dataSize", len(data))
	}

	newMsg := msg{
		header: header{
			MagicWord: tag,
			Command:   uint32(cmd),
			PkgType:   uint16(pkgType),
			Seq:       Seq,
			ReqId:     reqId,
			BodySize:  uint32(len(data)),
			ExtSize:   uint16(0),
		},
		Ext:  nil,
		Body: data,
	}
	return newMsg.Bytes(), nil
}

func (t *textRpcProtocol) unpackHeartbeatRequest(body []byte) (*grpc.Heartbeat, error) {
	var m heartbeat
	r := bytes.NewReader(body)
	err := binary.Read(r, endian, &m)
	if err != nil {
		return nil, gerr.WrapMsg(err, "protocol unpack heartbeat request error")
	}

	h := &grpc.Heartbeat{
		Timeout: m.Timeout,
	}
	return h, nil
}

func (t *textRpcProtocol) packHeartbeatRequest(h *grpc.Heartbeat) []byte {
	m := &heartbeat{
		Type:        1,         // 心跳包type,发送方固定1
		TimeoutSize: 4,         // 心跳包timeoutSize,发送方固定4
		Timeout:     h.Timeout, // 心跳包timeout
	}

	w := bytes.NewBuffer([]byte{})
	err := binary.Write(w, endian, m)
	if err != nil {
		return nil
	}
	return w.Bytes()
}
