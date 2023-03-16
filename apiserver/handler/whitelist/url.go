package whitelist

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/handler"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/service"
	"github.com/xnile/muxwaf/pkg/logx"
	"strconv"
	"strings"
)

func AddURL(c *gin.Context) {
	var payload model.WhitelistURLModel
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	svc := service.SVC.Whitelist
	err := svc.AddURL(c, &payload)
	if err != nil {
		logx.Warnf("add whitelist url err, %v", err)
	}
	handler.ResponseBuilder(c, err, nil)
}

func GetURLList(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)
	siteID, _ := strconv.ParseInt(c.Query("site_id"), 10, 64)
	url := strings.TrimSpace(c.Query("url"))

	var status int16
	switch c.Query("status") {
	case "0":
		status = 0
	case "1":
		status = 1
	default:
		status = -1
	}
	svc := service.SVC.Whitelist
	data, err := svc.ListURL(pageNum, pageSize, siteID, status, url)
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, err, data)
}

func UpdateURLStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.Whitelist
	if err := svc.UpdateURLStatus(id); err != nil {
		handler.ResponseBuilder(c, ecode.InternalServerError, nil)
	}
	handler.ResponseBuilder(c, ecode.Success, nil)
}

func DeleteURL(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.Whitelist
	if err := svc.DeleteURL(id); err != nil {
		handler.ResponseBuilder(c, ecode.InternalServerError, nil)
	}
	handler.ResponseBuilder(c, ecode.Success, nil)
}

func UpdateURL(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var payload model.WhitelistURLModel
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	svc := service.SVC.Whitelist
	err := svc.UpdateURL(id, &payload)
	if err != nil {
		logx.Warnf("update whitelist url err, %v", err)
	}
	handler.ResponseBuilder(c, err, nil)

}
