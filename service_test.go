package gotato

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/qiafan666/gotato/service/glog"
	"testing"
)

func TestStart_Default_Server(t *testing.T) {
	glog.Slog.InfoF(context.Background(), "TestStart_Default_Server")
	server := GetGotato()

	server.StartServer(GinService)

	server.App().GET("/ping", func(c *gin.Context) {
		ctx := c.MustGet("ctx").(context.Context)
		glog.Slog.InfoF(ctx, "1")
		glog.Slog.InfoF(ctx, "2")
		glog.Slog.InfoF(ctx, "3")
		glog.Slog.InfoF(ctx, "4")
		glog.Slog.InfoF(ctx, "5")
		glog.Slog.InfoF(ctx, "6")

		c.String(200, "123")
	})
	server.WaitClose()
}
