package middleware

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qiafan666/gotato/commons"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/glog"
	"io"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

var SimpleStdout bool

func Default(ctx *gin.Context) {
	uuid := gcommon.GenerateUUID()
	value := context.WithValue(ctx, "trace_id", uuid)
	ctx.Set("trace_id", uuid)
	ctx.Set("ctx", value)
	atomic.AddInt64(&commons.ActiveRequests, 1)
	defer atomic.AddInt64(&commons.ActiveRequests, -1)
	defer func() {
		if err := recover(); err != nil {
			var stacktrace string
			for i := 1; ; i++ {
				_, f, l, got := runtime.Caller(i)
				if !got {
					break
				}
				stacktrace += fmt.Sprintf("%s:%d\n", f, l)
			}
			// when stack finishes
			logMessage := fmt.Sprintf("Recovered from a route's Handler('%s')\n", ctx.HandlerName())
			logMessage += fmt.Sprintf("Trace: %s", err)
			logMessage += fmt.Sprintf("\n%s", stacktrace)
			glog.Slog.ErrorF(ctx, logMessage)
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
	}()

	blw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
	ctx.Writer = blw

	if !IsIgnoredRequest(ctx.Request.URL.Path) {
		var bodyBytes []byte
		var err error
		var requestBody *bytes.Buffer
		if ctx.Request.Method == http.MethodPost {
			bodyBytes, err = io.ReadAll(ctx.Request.Body)
			if err != nil {
				glog.Slog.ErrorF(ctx, "ReadAll %s", err)
			} else if len(bodyBytes) > 0 {
				requestBody = bytes.NewBuffer(bodyBytes)
				ctx.Request.Body = io.NopCloser(requestBody)
			} else {
				requestBody = bytes.NewBuffer([]byte(""))
			}
		}
		start := time.Now()
		ctx.Next()

		path := ctx.Request.URL.Path
		if ctx.Request.URL.RawQuery != "" {
			path += "?" + ctx.Request.URL.RawQuery
		}

		if SimpleStdout {
			glog.Slog.InfoF(ctx, "【%s:%s】【%s】【%dms】【response code:%d】", ctx.Request.Method, path, ctx.ClientIP(), time.Now().Sub(start).Milliseconds(), ctx.Writer.Status())
		} else {
			glog.Slog.InfoF(ctx, "【%s:%s】【%s】【%dms】【response code:%d】【request:%s】【response:%s】", ctx.Request.Method, path, ctx.ClientIP(), time.Now().Sub(start).Milliseconds(), ctx.Writer.Status(), string(bodyBytes), blw.body.String())
		}
	} else {
		ctx.Next()
	}

}
