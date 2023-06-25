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

func AddSite(c *gin.Context) {
	var p model.SiteReq
	if err := c.ShouldBindJSON(&p); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	//logx.Warnf("add site, %v", p.Domain)
	svc := service.SVC.Site
	err := svc.Add(&p)
	handler.ResponseBuilder(c, err, nil)
}

func GetSiteList(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)
	statusStr := c.Query("status")
	domain := c.Query("domain")
	isFuzzy := c.Query("is_fuzzy")

	var status int16
	switch statusStr {
	case "0":
		status = 0
	case "1":
		status = 1
	default:
		status = -1
	}

	svc := service.SVC.Site
	data, err := svc.List(pageNum, pageSize, status, domain, isFuzzy)
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, err, data)
}

func GetAllSite(c *gin.Context) {
	svc := service.SVC.Site
	data, err := svc.GetAll()
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, err, data)
}

//func UpdateSiteConfig(c *gin.Context) {
//	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
//
//	var payload model.SiteConfigReq
//	if err := c.ShouldBindJSON(&payload); err != nil {
//		logx.Warnf("request parameter error: %v", err)
//		handler.ResponseBuilder(c, ecode.ErrParam, nil)
//		return
//	}
//	svc := service.SVC.Site
//	err := svc.UpdateConfig(id, &payload)
//	if err != nil {
//		logx.Warnf("update site err, %v", err)
//	}
//	handler.ResponseBuilder(c, err, nil)
//
//}

func UpdateBasicCfg(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var payload model.SiteBasicCfgReq
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	svc := service.SVC.Site
	err := svc.UpdateBasicConfigs(id, &payload)
	if err != nil {
		logx.Warnf("update site err, %v", err)
	}
	handler.ResponseBuilder(c, err, nil)

}

func UpdateHttps(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var payload model.SiteHttpsReq
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	svc := service.SVC.Site
	err := svc.UpdateHttpsConfigs(id, &payload)
	if err != nil {
		logx.Warnf("update site err, %v", err)
	}
	handler.ResponseBuilder(c, nil, nil)

}

//func GetOrigins(c *gin.Context) {
//	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
//	svc := service.SVC.Site
//	data := svc.GetOrigins(id)
//	handler.ResponseBuilder(c, nil, data)
//}

func GetOriginCfg(c *gin.Context) {
	siteID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.Site
	data, err := svc.GetOriginCfg(siteID)
	handler.ResponseBuilder(c, err, data)
}

func UpdateOrigin(c *gin.Context) {
	//id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	//
	//var payload model.SiteOriginModel
	//if err := c.ShouldBindJSON(&payload); err != nil {
	//	logx.Warnf("request parameter error: %v", err)
	//	handler.ResponseBuilder(c, ecode.ErrParam, nil)
	//	return
	//}
	//svc := service.SVC.Site
	//err := svc.UpdateOrigin(id, &payload)
	//handler.ResponseBuilder(c, err, nil)

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var payload model.OriginCfgReq
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	svc := service.SVC.Site
	err := svc.UpdateOriginCfg(id, payload)
	handler.ResponseBuilder(c, err, nil)

}

//func AddOrigins(c *gin.Context) {
//	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
//
//	payload := make([]*model.SiteOriginModel, 0)
//	if err := c.ShouldBindJSON(&payload); err != nil {
//		logx.Warnf("request parameter error: %v", err)
//		handler.ResponseBuilder(c, ecode.ErrParam, nil)
//		return
//	}
//	svc := service.SVC.Site
//	err := svc.AddSiteOrigins(id, payload)
//	handler.ResponseBuilder(c, err, nil)
//}

func GetHttpsConfigs(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	svc := service.SVC.Site
	data, err := svc.GetHttpsConfigs(id)
	handler.ResponseBuilder(c, err, data)

}

//func GetConfigs(c *gin.Context) {
//	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
//	svc := service.SVC.Site
//	data, err := svc.GetConfigs(id)
//	handler.ResponseBuilder(c, err, data)
//}

func GetBasicConfigs(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	svc := service.SVC.Site
	data, err := svc.GetBasicHttps(id)
	handler.ResponseBuilder(c, err, data)
}

//func DelOrigin(c *gin.Context) {
//	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
//	svc := service.SVC.Site
//	err := svc.DelOrigin(id)
//	handler.ResponseBuilder(c, err, nil)
//}

func DelSite(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	svc := service.SVC.Site
	err := svc.Del(id)
	handler.ResponseBuilder(c, err, nil)
}

func GetCandidateCertificates(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	domain := c.Query("domain")

	data, err := service.SVC.Site.GetCertificates(id, domain)
	handler.ResponseBuilder(c, err, data)
}

//func GetSiteDomain(c *gin.Context) {
//	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
//	data, err := service.SVC.Site.GetSiteDomain(id)
//	handler.ResponseBuilder(c, err, data)
//}
