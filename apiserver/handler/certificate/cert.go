package certificate

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/handler"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/service"
	"github.com/xnile/muxwaf/pkg/logx"
	"strconv"
)

func AddCert(c *gin.Context) {
	var payload model.CertModel
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	svc := service.SVC.Cert
	err := svc.Add(payload.Name, payload.Cert, payload.Key)
	if err != nil {
		logx.Warnf("add blacklist ip err, %v", err)
	}
	handler.ResponseBuilder(c, err, nil)
}

func GetCertList(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)

	svc := service.SVC.Cert
	data, err := svc.List(pageNum, pageSize)
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, err, data)
}

func DelCert(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.SVC.Cert
	err := svc.Delete(id)
	handler.ResponseBuilder(c, err, nil)
}

//func GetCertName(c *gin.Context) {
//	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
//
//	svc := service.SVC.Cert
//	certName, err := svc.GetCertName(id)
//	handler.ResponseBuilder(c, err, certName)
//}

func GetAll(c *gin.Context) {
	//data, err := service.SVC.Cert.GetALL()
	data, err := service.SVC.Site.GetCertificates(1, "xnile.cn")
	handler.ResponseBuilder(c, err, data)
}

func UpdateCertificate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := service.SVC.Cert.UpdateCert(c, id)
	handler.ResponseBuilder(c, err, nil)
}
