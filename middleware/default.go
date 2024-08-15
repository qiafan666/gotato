package middleware

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/glog"
	"runtime"
	"time"
)

func Default(ctx iris.Context) {
	value := context.WithValue(context.Background(), "trace_id", gcommon.GenerateUUID())
	ctx.Values().Set("ctx", value)
	defer func() {
		if err := recover(); err != nil {
			if ctx.IsStopped() {
				return
			}

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
			glog.Slog.ErrorF(value, logMessage)
			ctx.StatusCode(500)
			ctx.StopExecution()
		}
	}()

	start := time.Now()
	ctx.Next()

	if !IsIgnoredRequest(ctx.Request().URL.Path) {

		addr := ctx.Request().RemoteAddr
		if ctx.GetHeader("X-Real-IP") != "" {
			addr = ctx.GetHeader("X-Real-IP")
		}
		if ctx.GetHeader("X-Forwarded-For") != "" {
			addr = ctx.GetHeader("X-Forwarded-For")
		}

		path := ctx.Request().URL.Path
		if ctx.Request().URL.RawQuery != "" {
			path += "?" + ctx.Request().URL.RawQuery
		}

		glog.Slog.InfoF(value, "【response code:%d】【%s】【%dms】【%s:%s】", ctx.GetStatusCode(), addr, time.Now().Sub(start).Milliseconds(), ctx.Request().Method, path)
	}
}
