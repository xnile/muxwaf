package site

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/handler"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/service"
	"github.com/xnile/muxwaf/pkg/logx"
	"strconv"
)

func GetRegionBlacklist(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	data, err := service.SVC.Site.GetRegionBlacklist(id)
	handler.ResponseBuilder(c, err, data)
}

func UpdateRegionBlacklist(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	payload := new(model.SiteRegionBlacklistModel)
	if err := c.ShouldBindJSON(payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	err := service.SVC.Site.UpdateRegionBlacklist(id, payload)
	handler.ResponseBuilder(c, err, nil)
}
