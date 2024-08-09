package gcommon

import (
	"net/http"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

type ProxyRequestHeader struct {
	ContentType string
}

var TimeOut = time.Second * 5

func ProxyRequest(method string, header http.Header, url string, body []byte) (response []byte, respHeader ProxyRequestHeader, err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetBody(body)
	for s, v := range header {
		for _, v2 := range v {
			req.Header.Set(s, v2)
		}
	}
	req.Header.SetMethod(method)
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	if err = fasthttp.DoTimeout(req, resp, TimeOut); err != nil {
		return nil, ProxyRequestHeader{}, err
	}

	return resp.Body(), ProxyRequestHeader{ContentType: string(resp.Header.ContentType())}, nil
}

// GetRequestURL append request url
func GetRequestURL(url string, params map[string]string) string {
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

	// 移除最后一个 '&' 符号
	if strings.HasSuffix(urlAddress, "&") {
		urlAddress = urlAddress[:len(urlAddress)-1]
	}
	return urlAddress
}
