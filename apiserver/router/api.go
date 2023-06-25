package router

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/handler"
	"github.com/xnile/muxwaf/handler/blacklist"
	"github.com/xnile/muxwaf/handler/certificate"
	"github.com/xnile/muxwaf/handler/node"
	"github.com/xnile/muxwaf/handler/ratelimit"
	"github.com/xnile/muxwaf/handler/sample_log"
	"github.com/xnile/muxwaf/handler/site"
	"github.com/xnile/muxwaf/handler/user"
	"github.com/xnile/muxwaf/handler/whitelist"
	"github.com/xnile/muxwaf/router/middleware"
)

func Init() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(middleware.Cors)

	r.NoMethod(handler.RouteNotFound)
	r.NoRoute(handler.RouteNotFound)

	r.Use(middleware.Logging())
	r.Use(middleware.RequestID())

	gin.SetMode("")

	{
		r.POST("/api/auth/login", user.Login)
		r.POST("/api/auth/logout", user.Logout)
		r.POST("/api/logs/sample", sample_log.AddLog)
	}

	api := r.Group("api")
	api.Use(middleware.ParseToken(), middleware.AuthRequired())
	{
		api.GET("/users/info", user.UserInfo)
		// 用户
		api.POST("/users", user.InsertUser)
		api.GET("/users", user.ListUsers)
		api.PUT("/users/:id", user.UpdateUser)
		api.PUT("/users/reset-password", user.ResetPassword)

		// 黑名单
		api.POST("/blacklist/ip", blacklist.AddIP)
		api.GET("/blacklist/ip", blacklist.GetIPList)
		api.PUT("/blacklist/ip/:id", blacklist.UpdateIP)
		api.PUT("/blacklist/ip/:id/status", blacklist.UpdateIPStatus)
		api.DELETE("/blacklist/ip/:id", blacklist.DeleteIP)
		api.GET("/blacklist/ip/isIncluded", blacklist.IsIncluded)
		api.POST("/blacklist/ip/batch", blacklist.BatchAdd)
		api.GET("/blacklist/ip/export", blacklist.ExportIPList)

		// 白名单
		api.POST("/whitelist/ip", whitelist.AddIP)
		api.GET("/whitelist/ip", whitelist.GetIPList)
		api.DELETE("/whitelist/ip/:id", whitelist.DeleteIP)
		api.PUT("/whitelist/ip/:id", whitelist.UpdateIP)
		api.PUT("/whitelist/ip/:id/status", whitelist.UpdateIPStatus)
		api.GET("/whitelist/ip/isIncluded", whitelist.IsIpIncluded)
		api.POST("/whitelist/url", whitelist.AddURL)
		api.GET("/whitelist/url", whitelist.GetURLList)
		api.DELETE("/whitelist/url/:id", whitelist.DeleteURL)
		api.PUT("/whitelist/url/:id", whitelist.UpdateURL)
		api.PUT("/whitelist/url/:id/status", whitelist.UpdateURLStatus)
		api.POST("/whitelist/ip/batch", whitelist.BatchAddIP)

		// 频率限止
		api.POST("/rate-limit", ratelimit.Add)
		api.GET("/rate-limit", ratelimit.GetList)
		api.PUT("/rate-limit/:id", ratelimit.Update)
		api.DELETE("/rate-limit/:id", ratelimit.Delete)
		api.PUT("/rate-limit/:id/status", ratelimit.UpdateStatus)

		// 证书
		api.GET("/certificates", certificate.GetCertList)
		api.GET("/certificates/all", certificate.GetAll)
		api.POST("/certificates", certificate.AddCert)
		api.DELETE("/certificates/:id", certificate.DelCert)
		//api.GET("/certificates/:id/name", certificate.GetCertName)
		api.PUT("/certificates/:id", certificate.UpdateCertificate)

		// 网站
		api.POST("/sites", site.AddSite)
		api.GET("/sites", site.GetSiteList)
		api.DELETE("/sites/:id", site.DelSite)
		//api.PUT("/sites/:id/configs", site.UpdateSiteConfig)
		api.PUT("/sites/:id/configs/basic", site.UpdateBasicCfg)
		api.PUT("/sites/:id/configs/https", site.UpdateHttps)
		api.GET("/sites/all", site.GetAllSite)
		api.GET("/sites/:id/configs/origin", site.GetOriginCfg)
		//api.POST("/sites/:id/origins", site.AddOrigins)
		api.PUT("/sites/:id/configs/origin", site.UpdateOrigin)
		//api.PUT("/sites/origins/:id", site.UpdateOrigin)
		//api.DELETE("/sites/origins/:id", site.DelOrigin)
		api.GET("/sites/:id/configs/https", site.GetHttpsConfigs)
		//api.GET("/sites/:id/domain", site.GetSiteDomain)
		//api.GET("/sites/:id/configs", site.GetConfigs)
		api.GET("/sites/:id/configs/basic", site.GetBasicConfigs)
		api.GET("/sites/:id/certificates", site.GetCandidateCertificates)
		api.GET("/sites/:id/region-blacklist", site.GetRegionBlacklist)
		api.PUT("/sites/:id/region-blacklist", site.UpdateRegionBlacklist)

		// Guard节点
		api.POST("/nodes", node.AddNode)
		api.DELETE("/nodes/:id", node.DelNode)
		api.GET("/nodes", node.GetNodeList)
		api.PUT("/nodes/:id/sync", node.SyncConfigs)
		api.PUT("/nodes/:id/status", node.SwitchStatus)
		api.PUT("/nodes/:id/sample_log_upload", node.SwitchSampleLogUpload)

		// 攻击日志
		api.GET("/sample-logs", sample_log.GetLogList)

	}
	return r
}
