package ratelimit

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

func Add(c *gin.Context) {
	var payload model.RateLimitModel
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	svc := service.SVC.RateLimit
	err := svc.Add(&payload)
	if err != nil {
		logx.Warnf("add rate limit err, %v", err)
	}
	handler.ResponseBuilder(c, err, nil)
}

func GetList(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)
	siteID, _ := strconv.ParseInt(c.Query("site_id"), 10, 64)
	matchMode, _ := strconv.ParseInt(c.Query("match_mode"), 10, 16)
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

	svc := service.SVC.RateLimit
	data, err := svc.List(pageNum, pageSize, siteID, status, int16(matchMode), url)
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, err, data)
}

func UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.RateLimit
	if err := svc.UpdateStatus(id); err != nil {
		handler.ResponseBuilder(c, ecode.InternalServerError, nil)
	}
	handler.ResponseBuilder(c, ecode.Success, nil)
}

func Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.RateLimit
	if err := svc.Delete(id); err != nil {
		handler.ResponseBuilder(c, ecode.InternalServerError, nil)
	}
	handler.ResponseBuilder(c, ecode.Success, nil)
}

func Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var payload model.RateLimitModel
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	svc := service.SVC.RateLimit
	err := svc.Update(id, &payload)
	if err != nil {
		logx.Warnf("update rate limit err, %v", err)
	}
	handler.ResponseBuilder(c, err, nil)
}
