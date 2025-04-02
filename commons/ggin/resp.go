package ggin

import (
	"errors"
	"github.com/qiafan666/gotato/commons/gerr"
	"time"
)

type ApiResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Dlt       string `json:"dlt,omitempty"`
	Data      any    `json:"data"`
	Time      int64  `json:"time"`
	RequestId string `json:"request_id"`
}

func (r *ApiResponse) MarshalJSON() ([]byte, error) {
	type apiResponse ApiResponse
	tmp := (*apiResponse)(r)
	return gerr.Marshal(tmp)
}

func Api(code int, msg string, data any, requestId string) *ApiResponse {
	return &ApiResponse{Code: code, Msg: msg, Data: data, Time: time.Now().UnixNano() / 1e6, RequestId: requestId}
}

// ApiSuccess data 数据 strings[1]:requestID strings[2]:msg
func ApiSuccess(data any, strings ...string) *ApiResponse {
	msg := "suc"
	if len(strings) == 0 {
		return &ApiResponse{Code: gerr.OK, Data: data, Msg: msg, Time: time.Now().UnixNano() / 1e6}
	} else if len(strings) == 1 {
		return &ApiResponse{Code: gerr.OK, Data: data, RequestId: strings[0], Msg: msg, Time: time.Now().UnixNano() / 1e6}
	} else {
		msg = strings[1]
		return &ApiResponse{Code: gerr.OK, Data: data, RequestId: strings[0], Msg: msg, Time: time.Now().UnixNano() / 1e6}
	}
}

func ParseError(err error) *ApiResponse {
	if err == nil {
		return ApiSuccess(nil)
	}
	unwrap := gerr.Unwrap(err)
	var codeErr gerr.ICodeError
	if errors.As(unwrap, &codeErr) {
		resp := ApiResponse{Code: codeErr.Code(), Msg: codeErr.Msg(), Dlt: codeErr.Detail(), Time: time.Now().UnixNano() / 1e6, RequestId: codeErr.RequestID()}
		if resp.Dlt == "" && codeErr.Msg() != err.Error() {
			resp.Dlt = err.Error()
		}
		return &resp
	}
	return &ApiResponse{Code: gerr.UnKnowError, Msg: err.Error(), Dlt: "error type is not gerr.CodeError", Time: time.Now().UnixNano() / 1e6, RequestId: ""}
}
