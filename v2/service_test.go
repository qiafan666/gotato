package v2

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/qiafan666/gotato/commons/log"
	"testing"
)

func TestStart_Default_Server(t *testing.T) {
	log.Slog.InfoF(context.Background(), "TestStart_Default_Server")
	server := GetGotatoInstance()

	server.StartServer(GinService)

	server.App().GET("/ping", func(c *gin.Context) {
		ctx := c.MustGet("ctx").(context.Context)
		log.Slog.InfoF(ctx, "1")
		log.Slog.InfoF(ctx, "2")
		log.Slog.InfoF(ctx, "3")
		log.Slog.InfoF(ctx, "4")
		log.Slog.InfoF(ctx, "5")
		log.Slog.InfoF(ctx, "6")

		c.String(200, "123")
	})
	server.WaitClose()
}
