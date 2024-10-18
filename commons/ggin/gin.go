package ggin

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GinError(c *gin.Context, err error) {
	c.JSON(http.StatusOK, ParseError(err))
}

// GinSuccess strings[0] = requestID strings[1] = message
func GinSuccess(c *gin.Context, data any, strings ...string) {
	c.JSON(http.StatusOK, ApiSuccess(data, strings...))
}

// GinCustom 自定义
func GinCustom(c *gin.Context, code int, msg string, data any, requestId string) {
	c.JSON(http.StatusOK, Api(code, msg, data, requestId))
}
