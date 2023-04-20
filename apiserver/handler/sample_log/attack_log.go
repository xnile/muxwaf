package sample_log

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

func AddLog(c *gin.Context) {
	var payload = make([]*model.SampleLogModel, 0)
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	//svc := service.SVC.AttackLog
	//err := svc.Add(c, &payload)
	//if err != nil {
	//	logx.Warnf("add blacklist ip err, %v", err)
	//}
	//handler.ResponseBuilder(c, err, nil)
	service.SVC.AttackLog.Add(c, payload)
}

func GetLogList(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)
	starTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	siteID, _ := strconv.ParseInt(c.Query("site_id"), 10, 64)
	content := strings.TrimSpace(c.Query("content"))

	var action int8
	switch c.Query("action") {
	case "1":
		action = 1
	case "2":
		action = 2
	default:
		action = -1
	}

	svc := service.SVC.AttackLog
	data, err := svc.List(pageNum, pageSize, starTime, endTime, siteID, action, content)
	handler.ResponseBuilder(c, err, data)
}
