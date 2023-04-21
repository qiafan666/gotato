package commons

import (
	"time"
)

type BaseResponse struct {
	Code ResponseCode `json:"code"`
	Msg  string       `json:"msg"`
	Data interface{}  `json:"data"`
	Time int64        `json:"time"`
}

type BaseResponseHeader struct {
	Code ResponseCode `json:"code"`
	Msg  string       `json:"msg"`
	Time int64        `json:"time"`
}

// return struct of the response code and msg
func BuildResponse(code ResponseCode, msg string, data interface{}) *BaseResponse {
	return &BaseResponse{code, msg, data, time.Now().UnixNano() / 1e6}
}

func BuildSuccess(data interface{}, language string) *BaseResponse {

	return &BaseResponse{Code: OK, Msg: GetCodeAndMsg(OK, language), Data: data, Time: time.Now().UnixNano() / 1e6}
}
func BuildSuccessWithMsg(msg string, data interface{}) *BaseResponse {

	return &BaseResponse{Code: OK, Msg: msg, Data: data, Time: time.Now().UnixNano() / 1e6}
}

func BuildFailed(code ResponseCode, language string) *BaseResponse {
	if code == 0 {
		code = UnKnowError
	}
	return &BaseResponse{
		Code: code,
		Msg:  GetCodeAndMsg(code, language),
		Data: struct{}{},
		Time: time.Now().UnixNano() / 1e6,
	}
}
func BuildFailedWithMsg(code ResponseCode, msg string) *BaseResponse {
	message := msg
	return &BaseResponse{
		Code: code,
		Msg:  message,
		Data: struct{}{},
		Time: time.Now().UnixNano() / 1e6,
	}
}
