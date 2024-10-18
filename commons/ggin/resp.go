package ggin

import (
	"encoding/json"
	"errors"
	"github.com/qiafan666/gotato/commons/gerr"
	"reflect"
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
	if tmp.Data != nil {
		if isAllFieldsPrivate(tmp.Data) {
			tmp.Data = json.RawMessage(nil)
		} else {
			data, err := gerr.Marshal(tmp.Data)
			if err != nil {
				return nil, err
			}
			tmp.Data = json.RawMessage(data)
		}
	}
	return gerr.Marshal(tmp)
}

func isAllFieldsPrivate(v any) bool {
	typeOf := reflect.TypeOf(v)
	if typeOf == nil {
		return false
	}
	for typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}
	if typeOf.Kind() != reflect.Struct {
		return false
	}
	num := typeOf.NumField()
	for i := 0; i < num; i++ {
		c := typeOf.Field(i).Name[0]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
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
	var codeErr gerr.CodeError
	if errors.As(unwrap, &codeErr) {
		resp := ApiResponse{Code: codeErr.Code(), Msg: codeErr.Msg(), Dlt: codeErr.Detail(), Time: time.Now().UnixNano() / 1e6, RequestId: codeErr.RequestID()}
		if resp.Dlt == "" {
			resp.Dlt = err.Error()
		}
		return &resp
	}
	return &ApiResponse{Code: gerr.UnKnowError, Msg: err.Error()}
}
