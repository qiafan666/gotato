package gotato

import (
	"github.com/kataras/iris/v12"
	"testing"
)

func TestStart_Default_Server(t *testing.T) {
	server := GetGotatoInstance()
	server.app.Default()
	server.StartServer()

	Instance.app.Get("/ping", func(context iris.Context) {
		context.WriteString("123")
	})
	server.WaitClose(iris.WithoutBodyConsumptionOnUnmarshal)

}
