package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	slog "github.com/qiafan666/quickweb/commons/log"
	"github.com/valyala/fasthttp"
)

type BaseResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
	Time int64           `json:"time"`
}

type ProxyRequestHeader struct {
	ContentType string
}

var TimeOut = time.Second * 10

func ProxyRequest(ctx context.Context, method string, header http.Header, url string, body []byte) (response []byte, respHeader ProxyRequestHeader, err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetBody(body)
	for s, v := range header {
		for _, v2 := range v {
			req.Header.Set(s, v2)
		}
	}
	req.Header.SetMethod(method)
	req.Header.Set(fasthttp.HeaderConnection, fasthttp.HeaderKeepAlive)
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	if err := fasthttp.DoTimeout(req, resp, time.Second*5); err != nil {
		slog.Slog.InfoF(ctx, "Http Request Do Error %s", err.Error())
		return nil, ProxyRequestHeader{}, err
	}

	return resp.Body(), ProxyRequestHeader{ContentType: string(resp.Header.ContentType())}, nil

}

//append request url
func getRequestURL(url string, params map[string]string) string {
	var urlAddress = ""
	lastCharctor := url[len(url)-1:]
	if lastCharctor == "?" {
		urlAddress = url + urlAddress
	} else {
		urlAddress = url + "?" + urlAddress
	}
	for k, v := range params {
		if len(k) != 0 && len(v) != 0 {
			urlAddress = urlAddress + k + "=" + v + "&"
		}
	}
	return urlAddress
}
