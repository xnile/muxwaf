package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/internal/ecode"
	"net/http"
)

// ResponseBuilder response wrapper
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ResponseBuilder json response wrapper
func ResponseBuilder(c *gin.Context, err error, data interface{}) {
	code, msg := ecode.DecodeErr(err)

	// always return http.StatusOK
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

// RouteNotFound can't find route
func RouteNotFound(c *gin.Context) {
	c.String(http.StatusNotFound, "the route not found")
}
