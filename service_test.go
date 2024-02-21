package gotato

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/qiafan666/gotato/commons"
	"testing"
)

func TestStart_Default_Server(t *testing.T) {
	server := GetGotatoInstance()
	server.app.Default()
	server.StartServer()

	Instance.app.Get("/ping", func(ctx iris.Context) {
		c := ctx.Values().Get("ctx").(context.Context)
		msg := commons.BuildSuccessWithMsg("测试成功", nil, c.Value("trace_id").(string))
		marshal, err := json.Marshal(msg)
		if err != nil {
			return
		}
		fmt.Println(string(marshal))
		_ = ctx.JSON(msg)
	})
	server.WaitClose(iris.WithoutBodyConsumptionOnUnmarshal)

}
