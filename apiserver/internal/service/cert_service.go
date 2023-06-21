package service

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/event"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/repository"
	"github.com/xnile/muxwaf/pkg/logx"
	"github.com/xnile/muxwaf/pkg/utils"
	"gorm.io/gorm"
	"strings"
)

type ICertService interface {
	Add(name, cert, key string) error
	List(pageNum, pageSize int64) (*model.ListResp, error)
	Delete(id int64) error
	GetALL() ([]*model.CertCandidateResp, error)
	//GetCertName(id int64) (string, error)
	UpdateCert(c *gin.Context, id int64) error
}

type certService struct {
	repo     *repository.Repository
	eventBus *event.EventBus
}

func NewCertService(repo *repository.Repository, eventBus *event.EventBus) ICertService {
	return &certService{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (svc *certService) Add(name, cert, key string) error {
	cert = strings.TrimSpace(cert)
	if !utils.IsValidPEMCertificate(cert) {
		return ecode.ErrCertInvalid
	}
	if !utils.IsValidPEMPrivateKey(key) {
		return ecode.ErrCertPriKeyInvalid
	}

	block, _ := pem.Decode([]byte(cert))
	if block == nil {
		return ecode.ErrCertInvalid
	}
	_cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		logx.Error("[certificate] failed to parse certificate: ", err.Error())
		return ecode.ErrCertInvalid
	}

	entity := model.CertModel{
		Name:      name,
		Cert:      cert,
		Key:       key,
		CN:        _cert.Issuer.CommonName,
		Sans:      _cert.DNSNames,
		BeginTime: _cert.NotBefore.Unix(),
		EndTime:   _cert.NotAfter.Unix(),
	}
	if err := svc.repo.DB.Create(&entity).Error; err != nil {
		return ecode.InternalServerError
	}

	//update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config := map[string]string{
			"id":   entity.UUID,
			"cert": entity.Cert,
			"key":  entity.Key,
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Certificate, event.OpTypeAdd, configs)
	}

	return nil
}

func (svc *certService) List(pageNum, pageSize int64) (*model.ListResp, error) {
	rsp := new(model.ListResp)
	entities := make([]*model.CertModel, 0)
	var count int64

	certRespList := make([]*model.CertResp, 0)
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)

	gDB := svc.repo.DB.Model(&model.CertModel{})
	if err := gDB.Count(&count).Error; err != nil {
		logx.Error("[certificate]Failed to get the total number of certificates: ", err)
		return nil, ecode.InternalServerError
	}
	gDB = gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize))
	if err := gDB.Order("created_at DESC").
		Select("ID", "Name", "CN", "Sans", "BeginTime", "EndTime").
		Find(&entities).Error; err != nil {
		logx.Error("[certificate]Failed to obtaining certificates: ", err)
		return nil, ecode.InternalServerError
	}

	if err := copier.Copy(&certRespList, &entities); err != nil {
		logx.Error("[certificate]Failed to copy entities: ", err)
		return nil, ecode.InternalServerError
	}

	for _, certResp := range certRespList {
		certResp.Sites = make([]model.CertBindSite, 0)
		_siteConfigEntities := make([]model.SiteConfigModel, 0)
		if err := svc.repo.DB.Select("SiteID").
			Where("cert_id = ?", certResp.ID).
			Find(&_siteConfigEntities).
			Error; err != nil {
			//	TODO:
			logx.Error("[certificate] Failed to get binding site: ", err)
			continue
		}
		for _, siteConfigEntity := range _siteConfigEntities {
			siteEntity := new(model.SiteModel)
			if err := svc.repo.DB.Select("ID", "Domain").
				Where("id = ?", siteConfigEntity.SiteID).
				Find(&siteEntity).
				Error; err != nil {
				logx.Error("[certificate] Failed to get binding site: ", err)
				continue
			}

			certResp.Sites = append(certResp.Sites, model.CertBindSite{
				ID:     siteEntity.ID,
				Domain: siteEntity.Domain,
			})
		}

	}

	rsp.SetValue(certRespList)
	rsp.SetMeta(pageSize, pageNum, count)
	return rsp, nil

}

func (svc *certService) Delete(id int64) error {
	entity := new(model.CertModel)
	if err := svc.repo.DB.Where("id = ?", id).First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrIDNotFound
		} else {
			logx.Error("[certificate] get certificate err: ", err)
			return ecode.InternalServerError
		}
	}

	siteConfigEntities := make([]model.SiteConfigModel, 0)
	if err := svc.repo.DB.Where("cert_id = ?", id).Find(&siteConfigEntities).Error; err != nil {
		logx.Error("[certificate] searching who is using a certificate err: ", err)
		return ecode.InternalServerError
	}

	if len(siteConfigEntities) > 0 {
		return ecode.ErrCertInUse
	}

	if err := svc.repo.DB.Where("id = ?", id).Delete(&model.CertModel{}).Error; err != nil {
		logx.Error("[certificate] delete certificate err: ", err)
		return ecode.InternalServerError
	}

	//update guard
	{
		configs := make(model.GuardDelArrayRsp, 0)
		configs = append(configs, entity.UUID)
		svc.eventBus.PushEvent(event.Certificate, event.OpTypeDel, configs)
	}
	return nil
}

func (svc *certService) GetALL() ([]*model.CertCandidateResp, error) {
	//entities, err := svc.repo.Cert.All()
	//if err != nil {
	//	logx.Error(err)
	//}
	//return entities
	certAllRespList := make([]*model.CertCandidateResp, 0)
	entities := make([]*model.CertModel, 0)
	if err := svc.repo.DB.Select("ID", "Name").Find(&entities).Error; err != nil {
		logx.Error("[certificate] Faild get all certificates: ", err)
		return nil, ecode.InternalServerError
	}

	if err := copier.Copy(&certAllRespList, &entities); err != nil {
		logx.Error("[certificate]Failed to copy 'CertCandidateResp' entities: ", err)
		return nil, ecode.InternalServerError
	}

	return certAllRespList, nil
}

//func (svc *certService) GetCertName(id int64) (string, error) {
//	entity := new(model.CertModel)
//	if err := svc.repo.DB.Where("id = ?", id).
//		Select("Name").
//		First(entity).Error; err != nil {
//		logx.Error("[certificate] get cert name err: ", err)
//		return "", ecode.InternalServerError
//	}
//	return entity.Name, nil
//}

func (svc *certService) UpdateCert(c *gin.Context, id int64) error {
	entity := new(model.CertModel)
	if err := svc.repo.DB.Where("id = ?", id).Select("UUID").First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrCertNotFound
		}
		logx.Error("[certificate]Failed to check if the certificate exists: ", err)
		return ecode.ErrCertNotFound
	}

	payload := new(model.CertModel)
	if err := c.ShouldBindJSON(payload); err != nil {
		return ecode.ErrParam
	}

	if !utils.IsValidPEMCertificate(payload.Cert) {
		return ecode.ErrCertInvalid
	}

	if !utils.IsValidPEMPrivateKey(payload.Key) {
		return ecode.ErrCertPriKeyInvalid
	}

	block, _ := pem.Decode([]byte(payload.Cert))
	if block == nil {
		return ecode.ErrCertInvalid
	}
	_cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		logx.Error("[certificate] failed to parse certificate: ", err.Error())
		return ecode.ErrCertInvalid
	}

	fields := model.CertModel{
		Name:      payload.Name,
		Cert:      payload.Cert,
		Key:       payload.Key,
		CN:        _cert.Issuer.CommonName,
		Sans:      _cert.DNSNames,
		BeginTime: _cert.NotBefore.Unix(),
		EndTime:   _cert.NotAfter.Unix(),
	}

	if err := svc.repo.DB.Model(&model.CertModel{}).
		Where("id = ?", id).
		Select("Name", "Cert", "Key", "CN", "Sans", "BeginTime", "EndTime").
		Updates(fields).Error; err != nil {
		logx.Error("[certificate]Failed to update certificate: ", err)
		return ecode.ErrUpdate
	}

	// 更新guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config := map[string]string{
			"id":   entity.UUID,
			"cert": entity.Cert,
			"key":  entity.Key,
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Certificate, event.OpTypeUpdate, configs)
	}
	return nil
}
