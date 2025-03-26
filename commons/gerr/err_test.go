package gerr

import (
	"errors"
	"net/http"
	"testing"
)

func TestError(t *testing.T) {
	testErr := NewCode(10000, "test error")
	t.Log(testErr.Error()) // output: test error

	testErr1 := NewCode(10001, "test error1").Wrap()
	t.Log(testErr1.Error()) // output: test error1

	testErr2 := NewCode(10002, "test error2").WrapMsg("wrap msg", "key", "value")
	t.Log(testErr2.Error()) // output: wrap msg, key=value: test error2

	err := Unwrap(testErr2)
	if err != nil {
		t.Log(err.Error()) // output: wrap msg, key=value: test error2
	}

	testErr3 := NewCode(10003, "test error3").WithDetail("msg detail").WrapMsg("wrap msg", "key", "value")
	t.Log(testErr3.Error()) // output: test error3 ;detail=msg detail

	err = Unwrap(testErr3)
	if err == nil {
		return
	}
	t.Log(err.Error()) // output: test error3 ;detail=msg detail

	testErr4 := New("test error4", "key", "value")

	t.Log(testErr4.Error()) // output: test error4, key=value

	testErr5 := New("test error5", "key", "value").Wrap()
	t.Log(testErr5.Error()) // output: test error5, key=value

	testErr6 := New("test error6", "key", "value").WrapMsg("wrap msg", "key", "value")
	t.Log(testErr6.Error()) // output: wrap msg, key=value: test error6, key=value

	testErr7 := WrapMsg(errors.New("test error7"), "wrap msg", "key", "value")
	t.Log(testErr7.Error()) // output: wrap msg, key=value: test error7
}

func TestErrorCode(t *testing.T) {
	testErr := NewCode(10000, "test error").WrapMsg("wrap msg", "key", "value")
	err := Unwrap(testErr)
	if err != nil {
		t.Log(err) // output: 10000
	}

	unwrap := Unwrap(err)
	if codeErr, ok := unwrap.(ICodeError); ok {
		t.Log(codeErr.Code())   // output: 10000
		t.Log(codeErr.Msg())    // output: wrap msg, key=value: test error
		t.Log(codeErr.Detail()) // output:
	}
}

func TestNew(t *testing.T) {
	testErr := New("test error").WrapMsg("wrap msg", "key", "value")
	t.Log(testErr.Error()) // output: test error, key=value

	testErr1 := New("test error")

	t.Log(testErr1.Is(testErr))

	t.Log(Unwrap(testErr))
	t.Log(testErr1.Is(Unwrap(testErr)))

	err := Unwrap(testErr)
	t.Log(errors.Is(err, New(errors.New("test error").Error())))
}

type ApiResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Dlt       string `json:"dlt"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id"`
}

func ApiSuccessWithMsg(data any, msg, requestID string) *ApiResponse {
	return &ApiResponse{Data: data, Msg: msg, RequestID: requestID}
}

func ParseError(err error) *ApiResponse {
	if err == nil {
		return ApiSuccessWithMsg(nil, "", "")
	}
	unwrap := Unwrap(err)
	var codeErr ICodeError
	if errors.As(unwrap, &codeErr) {
		resp := ApiResponse{Code: codeErr.Code(), Msg: codeErr.Msg(), Dlt: codeErr.Detail()}
		if resp.Dlt == "" {
			resp.Dlt = err.Error()
		}
		return &resp
	}
	return &ApiResponse{Code: http.StatusInternalServerError, Msg: err.Error()}
}

func TestParseError(t *testing.T) {
	testErr := NewCode(10000, "test error").WrapMsg("wrap msg", "key", "value")

	parseError := ParseError(testErr)
	t.Log(parseError) // output: 10000
}
