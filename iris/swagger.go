package iris

import (
	"bytes"
	"github.com/kataras/iris/v12"
	"github.com/qiafan666/gotato/commons/glog"
	"github.com/qiafan666/gotato/gconfig"
	"html/template"
	"io/ioutil"
	"net/http"
)

func SwaggerUI(ctx iris.Context) {

	file, err := ioutil.ReadFile(gconfig.SC.SwaggerConfig.UiPath)
	if err != nil {
		ctx.Header("Content-Type", "text/plain; charset=utf-8")
		ctx.StatusCode(http.StatusInternalServerError)
		glog.Slog.ErrorF(nil, "read file failed err:%s", err.Error())
		_, _ = ctx.WriteString("read file failed")
	}

	swaggerTemplate := template.Must(template.New("swagger").Parse(string(file)))

	var payload bytes.Buffer
	if err := swaggerTemplate.Execute(&payload, struct{}{}); err != nil {
		glog.Slog.ErrorF(nil, "Could not render Swagger")

		ctx.Header("Content-Type", "text/plain; charset=utf-8")
		ctx.StatusCode(http.StatusInternalServerError)
		_, err := ctx.Write([]byte("Could not render Swagger"))
		if err != nil {
			glog.Slog.ErrorF(nil, "Failed to write response")
		}
	}

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.StatusCode(http.StatusOK)
	_, err = ctx.Write(payload.Bytes())
	if err != nil {
		glog.Slog.ErrorF(nil, "Failed to write response")
	}
}

func SwaggerJson(ctx iris.Context) {

	file, err := ioutil.ReadFile(gconfig.SC.SwaggerConfig.JsonPath)
	if err != nil {
		ctx.Header("Content-Type", "text/plain; charset=utf-8")
		ctx.StatusCode(http.StatusInternalServerError)
		glog.Slog.ErrorF(nil, "read file failed err:%s", err.Error())
		_, _ = ctx.WriteString("read file failed")
	}

	ctx.Header("Content-Type", "application/json; charset=utf-8")
	ctx.StatusCode(http.StatusOK)
	if _, err := ctx.Write(file); err != nil {
		glog.Slog.ErrorF(nil, "Failed to write response")
	}
}
