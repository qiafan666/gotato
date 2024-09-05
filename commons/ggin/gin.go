package apiresp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GinError(c *gin.Context, err error) {
	c.JSON(http.StatusOK, ParseError(err))
}

func GinSuccess(c *gin.Context, data any, requestID string) {
	c.JSON(http.StatusOK, ApiSuccess(data, requestID))
}

func GinSuccessWithMsg(c *gin.Context, data any, msg, requestID string) {
	c.JSON(http.StatusOK, ApiSuccessWithMsg(data, msg, requestID))
}
