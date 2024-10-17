package ggin

import (
	"github.com/qiafan666/gotato/commons/gerr"
	"net/http"
)

func httpJson(w http.ResponseWriter, data any) {
	body, err := gerr.Marshal(data)
	if err != nil {
		http.Error(w, "json marshal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// HttpError err 为gerr.Error类型
func HttpError(w http.ResponseWriter, err error) {
	httpJson(w, ParseError(err))
}

func HttpSuccess(w http.ResponseWriter, data any) {
	httpJson(w, ApiSuccess(data))
}
