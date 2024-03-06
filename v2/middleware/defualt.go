package middleware

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	slog "github.com/qiafan666/gotato/commons/log"
	"github.com/qiafan666/gotato/commons/utils"
	"io"
	"net/http"
	"runtime"
	"time"
)

func Default(ctx *gin.Context) {
	value := context.WithValue(context.Background(), "trace_id", utils.GenerateUUID())
	ctx.Set("ctx", value)
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
			slog.Slog.ErrorF(value, logMessage)
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
	}()

	blw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
	ctx.Writer = blw

	if _, ok := ignoreRequestMap.Load(ctx.Request.URL.Path); !ok {
		if ctx.Request.Method == http.MethodPost {
			all, err := io.ReadAll(ctx.Request.Body)
			if err != nil {
				slog.Slog.ErrorF(value, "ReadAll %s", err)
			} else if len(all) > 0 {
				slog.Slog.InfoF(value, "Body \n%s", string(all))
				ctx.Request.Body = io.NopCloser(bytes.NewBuffer(all))
			}
		}
		start := time.Now()
		ctx.Next()
		path := ctx.Request.URL.Path
		if ctx.Request.URL.RawQuery != "" {
			path += "?" + ctx.Request.URL.RawQuery
		}
		slog.Slog.InfoF(value, "[%s:%s] [%s] [%dms] [response code:%d] [response:%s]", ctx.Request.Method, path, ctx.ClientIP(), time.Now().Sub(start).Milliseconds(), ctx.Writer.Status(), blw.body.String())
	} else {
		ctx.Next()
	}

}
