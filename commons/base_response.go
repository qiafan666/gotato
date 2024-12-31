package commons

import (
	"github.com/qiafan666/gotato/commons/gerr"
)

type BaseResponse struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	RequestId string      `json:"request_id"`
}

type BaseResponseHeader struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Time int64  `json:"time"`
}

// BuildResponse return struct of the response code and msg
func BuildResponse(code int, msg string, data interface{}, requestId string) *BaseResponse {
	return &BaseResponse{code, msg, data, requestId}
}

func BuildSuccess(data interface{}, language string, requestId string) *BaseResponse {

	return &BaseResponse{Code: gerr.OK, Msg: gerr.GetLanguageMsg(gerr.OK, language), Data: data, RequestId: requestId}
}
func BuildSuccessWithMsg(msg string, data interface{}, requestId string) *BaseResponse {

	return &BaseResponse{Code: gerr.OK, Msg: msg, Data: data, RequestId: requestId}
}

func BuildFailed(code int, language string, requestId string) *BaseResponse {
	if code == 0 {
		code = gerr.UnKnowError
	}
	return &BaseResponse{
		Code:      code,
		Msg:       gerr.GetLanguageMsg(code, language),
		Data:      struct{}{},
		RequestId: requestId,
	}
}
func BuildFailedWithMsg(code int, msg string, requestId string) *BaseResponse {
	message := msg
	return &BaseResponse{
		Code:      code,
		Msg:       message,
		Data:      struct{}{},
		RequestId: requestId,
	}
}
