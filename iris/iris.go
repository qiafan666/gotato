package iris

import (
	"errors"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/qiafan666/gotato/commons"
	"github.com/qiafan666/gotato/config"
	"github.com/qiafan666/gotato/middleware"
	"net/http/pprof"
	"os"
)

type App struct {
	app *iris.Application
}

func (slf *App) Default() {
	slf.app = iris.New()
	//register middleware
	slf.app.UseGlobal(middleware.Default)
	//global error handling
	slf.app.OnAnyErrorCode(func(ctx iris.Context) {
		if ctx.GetStatusCode() == iris.StatusNotFound {
			_ = ctx.JSON(commons.BuildFailed(commons.HttpNotFound, commons.DefualtLanguage))
		} else {
			_ = ctx.JSON(commons.BuildFailed(commons.UnKnowError, commons.DefualtLanguage))
		}
	})
	slf.app.Logger().SetLevel(config.SC.SConfigure.LogLevel)
	slf.app.Logger().SetOutput(os.Stdout)
}

func (slf *App) New() {
	slf.app = iris.New()
	//global error handling
	slf.app.OnAnyErrorCode(func(ctx iris.Context) {
		if ctx.GetStatusCode() == iris.StatusNotFound {
			_ = ctx.JSON(commons.BuildFailed(commons.HttpNotFound, commons.DefualtLanguage))
		} else {
			_ = ctx.JSON(commons.BuildFailed(commons.UnKnowError, commons.DefualtLanguage))
		}
	})
	slf.app.Logger().SetLevel(config.SC.SConfigure.LogLevel)
	slf.app.Logger().SetOutput(os.Stdout)
}

// set middleware
func (slf *App) SetGlobalMiddleware(handlers ...context.Handler) {
	slf.app.UseGlobal(handlers...)
}

// set middleware
func (slf *App) SetMiddleware(handlers ...context.Handler) {
	slf.app.Use(handlers...)
}

// get Iris App
func (slf *App) GetIrisApp() *iris.Application {
	return slf.app
}

func (slf *App) Party(relativePath string, handlers ...context.Handler) {
	slf.app.Party(relativePath, handlers...)
}
func (slf *App) Post(relativePath string, handlers ...context.Handler) {
	slf.app.Post(relativePath, handlers...)
}
func (slf *App) Get(relativePath string, handlers ...context.Handler) {
	slf.app.Get(relativePath, handlers...)
}

// start server
func (slf *App) Start(params ...iris.Configurator) error {
	server := fmt.Sprintf("%s:%d", config.SC.SConfigure.Addr, config.SC.SConfigure.Port)
	if slf.app == nil {
		return errors.New("Server not init")
	}
	//开启swagger
	if config.SC.SwaggerConfig.Enable == true {
		slf.app.Use(func(c *context.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST") // 根据你的需求添加其他允许的方法
			c.Header("Access-Control-Allow-Headers", "*")         // 根据你的需求设置允许的请求头
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Next()
		})
		slf.app.Get("/swagger", SwaggerUI)
		slf.app.Get("/docs/swagger.json", SwaggerJson)
	}
	//开启pprof
	if config.SC.PProfConfig.Enable == true {
		slf.app.Get("/debug/pprof/", iris.FromStd(pprof.Index))
		slf.app.Get("/debug/pprof/cmdline", iris.FromStd(pprof.Cmdline))
		slf.app.Get("/debug/pprof/profile", iris.FromStd(pprof.Profile))
		slf.app.Get("/debug/pprof/symbol", iris.FromStd(pprof.Symbol))
		slf.app.Get("/debug/pprof/trace", iris.FromStd(pprof.Trace))
	}

	params = append(params, iris.WithoutStartupLog)

	return slf.app.Run(iris.Addr(server), params...)
}
