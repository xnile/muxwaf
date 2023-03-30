package blacklist

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/handler"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/service"
	"github.com/xnile/muxwaf/pkg/logx"
	"strconv"
)

func AddIP(c *gin.Context) {
	//var payload model.BlacklistIPModel
	//if err := c.ShouldBindJSON(&payload); err != nil {
	//	logx.Warnf("request parameter error: %v", err)
	//	handler.ResponseBuilder(c, ecode.ErrParam, nil)
	//	return
	//}

	svc := service.SVC.BlacklistIP
	err := svc.Add(c)
	handler.ResponseBuilder(c, err, nil)
}

func GetIPList(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)
	starTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	ip := c.Query("ip")
	var status int16
	statusStr := c.Query("status")
	if statusStr == "" {
		status = -1
	}
	switch statusStr {
	case "0":
		status = 0
	case "1":
		status = 1
	default:
		status = -1
	}

	svc := service.SVC.BlacklistIP
	data, err := svc.List(c, pageNum, pageSize, starTime, endTime, ip, status)
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, err, data)
}

func UpdateIPStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.BlacklistIP
	if err := svc.UpdateStatus(id); err != nil {
		handler.ResponseBuilder(c, ecode.InternalServerError, nil)
	}
	handler.ResponseBuilder(c, ecode.Success, nil)
}

func DeleteIP(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.BlacklistIP
	if err := svc.Delete(id); err != nil {
		handler.ResponseBuilder(c, ecode.InternalServerError, nil)
	}
	handler.ResponseBuilder(c, ecode.Success, nil)
}

func IsIncluded(c *gin.Context) {
	ip, exist := c.GetQuery("ip")
	if !exist {
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	ok, err := service.SVC.BlacklistIP.IsIncluded(ip)
	handler.ResponseBuilder(c, err, ok)
}

func UpdateIP(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var payload model.BlacklistIPModel
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	err := service.SVC.BlacklistIP.Update(c, id, &payload)
	handler.ResponseBuilder(c, err, err)
}

func BatchAdd(c *gin.Context) {
	var payload model.BlacklistBatchAddReq
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	err := service.SVC.BlacklistIP.BatchAdd(c, &payload)
	handler.ResponseBuilder(c, err, err)
}

func ExportIPList(c *gin.Context) {
	data, err := service.SVC.BlacklistIP.ExportIPList(c)
	handler.ResponseBuilder(c, err, data)
}
