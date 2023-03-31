package node

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/handler"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/service"
	"github.com/xnile/muxwaf/pkg/logx"
	"strconv"
)

func AddNode(c *gin.Context) {
	payload := new(model.NodeModel)
	if err := c.ShouldBindJSON(payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	err := service.SVC.Node.Add(payload)
	handler.ResponseBuilder(c, err, nil)
}

func DelNode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	svc := service.SVC.Node
	err = svc.Del(id)
	handler.ResponseBuilder(c, err, nil)
}

func GetNodeList(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)

	svc := service.SVC.Node
	data, err := svc.List(pageNum, pageSize)
	handler.ResponseBuilder(c, err, data)
}

func SyncConfigs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	svc := service.SVC.Node
	err = svc.Sync(id)
	handler.ResponseBuilder(c, err, nil)
}

func SwitchSampleLogUpload(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	err = service.SVC.Node.SwitchSampleLogUpload(id)
	handler.ResponseBuilder(c, err, nil)
}

func SwitchStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	err = service.SVC.Node.SwitchStatus(id)
	handler.ResponseBuilder(c, err, nil)
}
