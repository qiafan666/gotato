package chanrpc

import (
	"github.com/qiafan666/gotato/commons/gapp2/logger"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"reflect"
	"runtime/debug"
)

// Handler 方法句柄  处理CallInfo
type Handler func(reqCtx *ReqCtx)

// Callback 回调
type Callback func(ackCtx *AckCtx)

// Raw 未解码的原始消息
type Raw struct {
	Raw []byte // 原始数据
}

// ReqCtx 调用参数
type ReqCtx struct {
	reqID   int64        // 请求唯一ID
	id      uint32       // 消息类型id
	Req     any          // 入参
	chanAck chan *AckCtx // 结果信息返回通道
	replied bool         // 是否已经返回 由被调用方使用
	uid     uint32       // 调用方ID
}

// ReqID 返回唯一请求ID
func (reqCtx *ReqCtx) ReqID() int64 {
	return reqCtx.reqID
}

// Reply 执行请求响应逻辑
func (reqCtx *ReqCtx) Reply(ack any) {
	reqCtx.doReply(&AckCtx{
		reqID: reqCtx.reqID,
		Ack:   ack,
	})
}

// ReplyErr 回复通用错误信息，通常由框架层使用
func (reqCtx *ReqCtx) ReplyErr(err error) {
	reqCtx.doReply(&AckCtx{
		reqID: reqCtx.reqID,
		Err:   err,
	})
}

func (reqCtx *ReqCtx) doReply(ackCtx *AckCtx) {
	defer func() {
		if stack := gcommon.PrintPanicStack(); stack != "" {
			logger.DefaultLogger.ErrorF("chanrpc Client exec panic error: %s", stack)
		}
	}()
	// 检查返回通道，如果是Cast消息，则忽略Reply
	if reqCtx.chanAck == nil {
		// 如果Cast消息遇到错误，打印日志
		if ackCtx.Err != nil {
			logger.DefaultLogger.ErrorF("chanrpc ReqCtx doReply cast msg %v with error: %v, msg: %+v", reflect.TypeOf(reqCtx.Req), ackCtx.Err, reqCtx.Req)
		}
		return
	}
	// 检查是否已经被响应过
	if reqCtx.replied {
		logger.DefaultLogger.ErrorF("chanrpc ReqCtx doReply can not ret twice, %v", string(debug.Stack()))
		return
	}
	reqCtx.replied = true
	reqCtx.PendAck(ackCtx)
}

// PendAck 将结果信息放入返回通道 goroutine safe
func (reqCtx *ReqCtx) PendAck(ackCtx *AckCtx) bool {
	select {
	case reqCtx.chanAck <- ackCtx:
		return true
	default:
		logger.DefaultLogger.ErrorF("chanrpc ReqCtx PendAck channel full, Ack: %v, Err: %v", ackCtx.Ack, ackCtx.Err)
		return false
	}
}

// GetMsgID 调用消息ID
func (reqCtx *ReqCtx) GetMsgID() uint32 {
	return reqCtx.id
}

// GetStatName 获取消息统计Key
func (reqCtx *ReqCtx) GetStatName() string {
	if reqCtx.Req == nil {
		return "ChanMsg_nil"
	}

	return reflect.TypeOf(reqCtx.Req).String()
}

// GetUid 获取调用方ID
func (reqCtx *ReqCtx) GetUid() uint32 {
	return reqCtx.uid
}

// AckCtx 结果信息
type AckCtx struct {
	reqID int64
	Ack   any   // 结果值 作为回调函数的入参
	Err   error // 错误
	Ctx   sval.M
	uid   uint32 // 调用方ID
}

// GetStatName 获取消息统计Key
func (ackCtx *AckCtx) GetStatName() string {
	if ackCtx.Err != nil || ackCtx.Ack == nil {
		return "ChanAck_unknown"
	}

	return reflect.TypeOf(ackCtx.Ack).String()
}

// GetUid 获取调用方ID
func (ackCtx *AckCtx) GetUid() uint32 {
	return ackCtx.uid
}

// IMsgID 消息可实现该接口来自定义MsgID，达成如消息结构体复用等高级功能
type IMsgID interface {
	MsgID() uint32
}

// MsgID 求消息的消息ID，传入值必须是指针
func MsgID(m any) uint32 {
	if msgIDGen, ok := m.(IMsgID); ok {
		return msgIDGen.MsgID()
	}
	typ := reflect.TypeOf(m)
	if typ.Kind() == reflect.Struct {
		return gcommon.Str2Uint32(typ.Name())
	}
	return gcommon.Str2Uint32(typ.Elem().Name())
}
