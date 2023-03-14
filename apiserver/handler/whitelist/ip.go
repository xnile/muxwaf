package whitelist

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
	svc := service.SVC.Whitelist
	err := svc.AddIP(c)
	handler.ResponseBuilder(c, err, nil)
}

func GetIPList(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)
	starTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	statusStr := c.Query("status")
	ip := c.Query("ip")

	var status int16
	switch statusStr {
	case "0":
		status = 0
	case "1":
		status = 1
	default:
		status = -1
	}

	svc := service.SVC.Whitelist
	data, err := svc.ListIP(pageNum, pageSize, starTime, endTime, ip, status)
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, err, data)
}

func UpdateIPStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.Whitelist
	if err := svc.UpdateIPStatus(id); err != nil {
		handler.ResponseBuilder(c, ecode.InternalServerError, nil)
	}
	handler.ResponseBuilder(c, ecode.Success, nil)
}

func DeleteIP(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.Whitelist
	if err := svc.DeleteIP(id); err != nil {
		handler.ResponseBuilder(c, ecode.InternalServerError, nil)
	}
	handler.ResponseBuilder(c, ecode.Success, nil)
}

func UpdateIP(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var payload model.WhitelistIPModel
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	err := service.SVC.Whitelist.UpdateIP(c, id, &payload)
	handler.ResponseBuilder(c, err, nil)
}

func IsIpIncluded(c *gin.Context) {
	ip, exist := c.GetQuery("ip")
	if !exist {
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	ok, err := service.SVC.Whitelist.IsIpIncluded(ip)
	handler.ResponseBuilder(c, err, ok)

}
