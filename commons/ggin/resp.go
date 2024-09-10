package ggin

import (
	"encoding/json"
	"errors"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/commons/ggin/jsonutil"
	"net/http"
	"reflect"
)

type ApiResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Dlt       string `json:"dlt"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id"`
}

func (r *ApiResponse) MarshalJSON() ([]byte, error) {
	type apiResponse ApiResponse
	tmp := (*apiResponse)(r)
	if tmp.Data != nil {
		if isAllFieldsPrivate(tmp.Data) {
			tmp.Data = json.RawMessage(nil)
		} else {
			data, err := jsonutil.JsonMarshal(tmp.Data)
			if err != nil {
				return nil, err
			}
			tmp.Data = json.RawMessage(data)
		}
	}
	return jsonutil.JsonMarshal(tmp)
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

func ApiSuccess(data any, requestID string) *ApiResponse {
	return &ApiResponse{Data: data, RequestID: requestID}
}

func ApiSuccessWithMsg(data any, msg, requestID string) *ApiResponse {
	return &ApiResponse{Data: data, Msg: msg, RequestID: requestID}
}

func ParseError(err error) *ApiResponse {
	if err == nil {
		return ApiSuccessWithMsg(nil, "", "")
	}
	unwrap := gerr.Unwrap(err)
	var codeErr gerr.CodeError
	if errors.As(unwrap, &codeErr) {
		resp := ApiResponse{Code: codeErr.Code(), Msg: codeErr.Msg(), Dlt: codeErr.Detail(), RequestID: codeErr.RequestID()}
		if resp.Dlt == "" {
			resp.Dlt = err.Error()
		}
		return &resp
	}
	return &ApiResponse{Code: http.StatusInternalServerError, Msg: err.Error()}
}
