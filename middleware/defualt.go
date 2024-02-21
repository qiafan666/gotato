package middleware

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	slog "github.com/qiafan666/gotato/commons/log"
	"github.com/qiafan666/gotato/commons/utils"
	"runtime"
	"time"
)

func Default(ctx iris.Context) {
	value := context.WithValue(context.Background(), "trace_id", utils.GenerateUUID())
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
			slog.Slog.ErrorF(value, logMessage)
			ctx.StatusCode(500)
			ctx.StopExecution()
		}
	}()

	if _, ok := ignoreRequestMap.Load(ctx.Request().URL.Path); !ok {

		start := time.Now()
		ctx.Next()

		addr := ctx.Request().RemoteAddr
		if ctx.GetHeader("X-Forwarded-For") != "" {
			addr = ctx.GetHeader("X-Forwarded-For")
		}
		path := ctx.Request().URL.Path
		if ctx.Request().URL.RawQuery != "" {
			path += "?" + ctx.Request().URL.RawQuery
		}

		slog.Slog.InfoF(value, "[response code:%d] [%s] [%dms] [%s:%s]", ctx.GetStatusCode(), addr, time.Now().Sub(start).Milliseconds(), ctx.Request().Method, path)
	}
}
