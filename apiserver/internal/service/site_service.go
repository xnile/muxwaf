package service

import (
	"errors"
	"github.com/jinzhu/copier"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/event"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/repository"
	"github.com/xnile/muxwaf/pkg/logx"
	"github.com/xnile/muxwaf/pkg/utils"
	"gorm.io/gorm"
	"strings"
)

type ISiteService interface {
	Add(site *model.SiteReq) error
	Del(id int64) error
	List(pageNum, pageSize int64, status int16, domain, isFuzzy string) (*model.ListResp, error)
	GetAll() ([]map[string]any, error)
	GetHttps(id int64) (*model.SiteHttpsRsp, error)
	GetOrigins(id int64) []*model.SiteOriginRsp
	GetConfigs(siteID int64) (*model.SiteConfigRsp, error)
	GetCertificates(siteID int64, domain string) ([]*model.CertCandidateResp, error)
	GetRegionBlacklist(id int64) (*model.SiteRegionBlacklistRsp, error)
	UpdateConfig(siteID int64, configEntity *model.SiteConfigReq) error
	UpdateHttps(siteID int64, config *model.SiteHttpsReq) error
	UpdateOrigin(id int64, originModel *model.SiteOriginModel) error
	UpdateRegionBlacklist(id int64, region *model.SiteRegionBlacklistModel) error
	AddSiteOrigins(id int64, origins []*model.SiteOriginModel) error
	DelOrigin(id int64) error
}

type siteService struct {
	repo     *repository.Repository
	eventBus *event.EventBus
}

func NewSiteService(repo *repository.Repository, eventBus *event.EventBus) ISiteService {
	return &siteService{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (svc *siteService) Add(site *model.SiteReq) (err error) {
	siteEntity := model.SiteModel{
		Domain: site.Domain,
		Status: 1,
	}
	if err := svc.repo.DB.Create(&siteEntity).Error; err != nil {
		logx.Error("[site] add site err: ", err)
		return ecode.InternalServerError
	}

	if err := svc.repo.DB.Create(&model.SiteConfigModel{
		SiteID:             siteEntity.ID,
		CertID:             0,
		HttpPort:           80,
		HttpsPort:          443,
		OriginProtocol:     1,
		IsRealIPFromHeader: 0,
		RealIPHeader:       "",
	}).Error; err != nil {
		logx.Error("[site] add site config err: ", err)
		return ecode.InternalServerError
	}

	for _, origin := range site.Origins {
		if err := svc.repo.DB.Create(&model.SiteOriginModel{
			SiteID:   siteEntity.ID,
			HttpPort: origin.HttpPort,
			Weight:   origin.Weight,
			Type:     1,
			Host:     origin.Host,
		}).Error; err != nil {
			logx.Error("[site] add site origin err: ", err)
			return ecode.InternalServerError
		}
	}

	if err := svc.repo.DB.Create(&model.SiteRegionBlacklistModel{
		SiteID: siteEntity.ID,
		Status: 1,
	}).Error; err != nil {
		logx.Error("[site]Failed to add site region blacklist: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config := map[string]any{
			"id":   siteEntity.UUID,
			"host": siteEntity.Domain,
			"config": map[string]any{
				"is_https":               0,
				"is_real_ip_from_header": 0,
				"real_ip_header":         "",
				"origin_protocol":        1,
				"cert_id":                "",
			},
		}
		_origins := make([]map[string]any, 0)
		for _, origin := range site.Origins {
			_origin := map[string]interface{}{
				"host":       origin.Host,
				"http_port":  origin.HttpPort,
				"https_port": 443,
				"weight":     origin.Weight,
			}
			_origins = append(_origins, _origin)
		}
		config["origins"] = _origins
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeAdd, configs)
	}

	return nil
}

func (svc *siteService) UpdateConfig(siteID int64, req *model.SiteConfigReq) error {
	fields := make([]string, 0)
	configEntity := new(model.SiteConfigModel)
	if req.IsRealIPFromHeader != nil {
		fields = append(fields, "IsRealIPFromHeader")
		configEntity.IsRealIPFromHeader = *req.IsRealIPFromHeader
	}
	if req.RealIPHeader != nil {
		fields = append(fields, "RealIPHeader")
		configEntity.RealIPHeader = *req.RealIPHeader
	}
	if req.OriginProtocol != nil {
		fields = append(fields, "OriginProtocol")
		configEntity.OriginProtocol = *req.OriginProtocol
	}

	if err := svc.repo.DB.Model(&model.SiteConfigModel{}).Where("site_id = ?", siteID).
		Select(fields).
		Updates(configEntity).Error; err != nil {
		logx.Error("[site] update site config err: ", err)
		return ecode.InternalServerError
	}

	// guard update
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(siteID)
		if err != nil {
			return err
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
	}

	return nil
}

func (svc *siteService) UpdateHttps(siteID int64, config *model.SiteHttpsReq) error {
	configEntity := new(model.SiteConfigModel)
	configEntity.IsHttps = *config.IsHttps
	if *config.IsHttps == 0 {
		configEntity.CertID = 0
	} else {
		configEntity.CertID = *config.CertID
	}

	if err := svc.repo.DB.Model(&model.SiteConfigModel{}).Where("site_id = ?", siteID).
		Select("CertID", "IsHttps").
		Updates(configEntity).
		Error; err != nil {
		logx.Error("[site] update https config err: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(siteID)
		if err != nil {
			return err
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
	}
	return nil
}

func (svc *siteService) List(pageNum, pageSize int64, status int16, domain, isFuzzy string) (*model.ListResp, error) {
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)
	rsp := new(model.ListResp)
	entities := make([]*model.SiteModel, 0)
	var count int64

	gDB := svc.repo.DB.Model(&model.SiteModel{})
	if status != -1 {
		gDB.Where("status = ?", status)
	}
	if len(domain) > 0 {
		if isFuzzy == "1" {
			gDB.Where("domain like ?", "%"+domain+"%")
		} else {
			gDB.Where("domain = ?", domain)
		}
	}
	if err := gDB.Count(&count).Error; err != nil {
		logx.Error("[site] Failed to count site: ", err.Error())
		return nil, ecode.InternalServerError
	}
	if err := gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).
		Order("created_at DESC").Find(&entities).Error; err != nil {
		logx.Error("[site]Failed to obtaining site list: ", err.Error())
		return nil, ecode.InternalServerError
	}

	sitesRsp := make([]*model.SiteRsp, 0)
	for _, entity := range entities {
		siteRsp := model.SiteRsp{}
		originsRsp := make([]*model.SiteOriginRsp, 0)
		configRsp := model.SiteConfigRsp{}

		copier.Copy(&siteRsp, &entity)

		origins, _ := svc.repo.SiteOriginRepo.GetBySiteID(entity.ID)
		copier.Copy(&originsRsp, &origins)
		siteRsp.Origins = originsRsp

		config, _ := svc.repo.SiteConfigRepo.GetBySiteID(entity.ID)
		copier.Copy(&configRsp, &config)
		siteRsp.Config = &configRsp

		sitesRsp = append(sitesRsp, &siteRsp)
	}

	rsp.SetValue(sitesRsp)
	rsp.SetMeta(pageSize, pageNum, count)
	return rsp, nil
}

func (svc *siteService) GetAll() ([]map[string]any, error) {
	rsp := make([]map[string]any, 0)
	entities, err := svc.repo.Site.GetAll()
	if err != nil {
		return rsp, err
	}
	for _, entity := range entities {
		v := make(map[string]any)
		v["id"] = entity.ID
		v["domain"] = entity.Domain
		rsp = append(rsp, v)
	}
	return rsp, nil
}

func (svc *siteService) GetOrigins(id int64) []*model.SiteOriginRsp {
	r := make([]*model.SiteOriginRsp, 0)
	origins, err := svc.repo.SiteOriginRepo.GetBySiteID(id)
	copier.Copy(&r, &origins)
	if err != nil {
		logx.Error("get site origins err: ", err)
		return nil
	}
	return r

}

func (svc *siteService) AddSiteOrigins(id int64, origins []*model.SiteOriginModel) error {
	for _, origin := range origins {
		if err := svc.repo.SiteOriginRepo.Insert(&model.SiteOriginModel{
			SiteID:   id,
			HttpPort: origin.HttpPort,
			Weight:   origin.Weight,
			Type:     1,
			Host:     origin.Host,
		}); err != nil {
			logx.Error("add site origins err: ", err)
			return ecode.InternalServerError
		}
	}

	// update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(id)
		if err != nil {
			return err
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
	}
	return nil
}

func (svc *siteService) UpdateOrigin(id int64, originModel *model.SiteOriginModel) error {
	if err := svc.repo.DB.Model(&model.SiteOriginModel{}).Where("id = ?", id).
		Select("HttpPort", "Weight", "Host").
		Updates(originModel).Error; err != nil {
		//	TODO log
		return ecode.InternalServerError
	}

	originEntity := new(model.SiteOriginModel)
	_ = svc.repo.DB.Where("id = ?", id).First(originEntity).Error

	// update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(originEntity.SiteID)
		if err != nil {
			return err
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
	}
	return nil
}

func (svc *siteService) DelOrigin(id int64) error {
	origin := new(model.SiteOriginModel)
	if err := svc.repo.DB.Where("id = ?", id).Select("SiteID", "UUID").First(origin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrIDNotFound
		} else {
			logx.Error("[site] fetching site origin err: ", err)
			return ecode.InternalServerError
		}
	}

	origins := make([]*model.SiteOriginModel, 0)
	if err := svc.repo.DB.Where("site_id = ?", origin.SiteID).Select("ID").Find(&origins).Error; err != nil {
		logx.Error("[site] searching site origins err: ", err)
		return ecode.InternalServerError
	}
	if len(origins) < 2 {
		return ecode.ErrAtLeastOneOrigin
	}

	if err := svc.repo.DB.Where("id = ?", id).Delete(&model.SiteOriginModel{}).Error; err != nil {
		//	TODO log
		return ecode.InternalServerError
	}

	// update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(origin.SiteID)
		if err != nil {
			return err
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
	}
	return nil
}

func (svc *siteService) GetHttps(id int64) (*model.SiteHttpsRsp, error) {
	//r := new(model.SiteHttpsRsp)
	config, err := svc.repo.SiteConfigRepo.GetBySiteID(id)
	if err != nil {
		return nil, err
	}

	if config.CertID != 0 {
		if cert, err := svc.repo.Cert.Get(id); err == nil {
			return &model.SiteHttpsRsp{
				Https:    true,
				CertName: cert.Name,
			}, nil
		}
	}
	return &model.SiteHttpsRsp{
		Https:    false,
		CertName: "",
	}, nil
}

func (svc *siteService) GetConfigs(siteID int64) (*model.SiteConfigRsp, error) {
	var configEntity model.SiteConfigModel
	if err := svc.repo.DB.Where("site_id = ?", siteID).First(&configEntity).Error; err != nil {
		logx.Error("[ste] get site config err: ", err)
		return nil, ecode.InternalServerError
	}

	configRsp := model.SiteConfigRsp{}
	copier.Copy(&configRsp, &configEntity)
	return &configRsp, nil
}

func (svc *siteService) Del(id int64) error {
	var err error
	entity := new(model.SiteModel)
	if err = svc.repo.DB.Where("id = ?", id).First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrIDNotFound
		} else {
			return ecode.InternalServerError
		}
	}

	//检查是否存在相关规则
	{
		// url名单
		whitelistURLSlice := make([]*model.WhitelistURLModel, 0)
		if err := svc.repo.DB.Where("site_id = ?", id).Find(&whitelistURLSlice).Error; err != nil {
			logx.Error("[site]Failed to get url whitelist associated with the site: ", err.Error())
			return ecode.InternalServerError
		}
		if len(whitelistURLSlice) > 0 {
			return errors.New("此站点还有关联的URL白名单规则，请先删除")
		}

		//频率限制
		rateLimitsSlice := make([]*model.RateLimitModel, 0)
		if err := svc.repo.DB.Where("id = ?", id).Find(&rateLimitsSlice).Error; err != nil {
			logx.Error("[site]Failed to get rate limits associated with the site: ", err.Error())
			return ecode.InternalServerError
		}
		if len(rateLimitsSlice) > 0 {
			return errors.New("此站点还有关联的频率限制规则，请先删除")
		}
	}

	// 删除站点
	if err = svc.repo.DB.Where("id = ?", id).Delete(&model.SiteModel{}).Error; err != nil {
		logx.Error("[site] delete site err: ", err)
		return ecode.InternalServerError
	}
	//删除站点配置
	if err = svc.repo.DB.Where("site_id = ?", id).Delete(&model.SiteConfigModel{}).Error; err != nil {
		logx.Error("[site] delete site config err: ", err)
		return ecode.InternalServerError
	}
	//删除站点源站
	if err = svc.repo.DB.Where("site_id = ?", id).Delete(&model.SiteOriginModel{}).Error; err != nil {
		logx.Error("[site] delete site origins err: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{
		configs := make(model.GuardDelArrayRsp, 0)
		configs = append(configs, entity.UUID)
		svc.eventBus.PushEvent(event.Site, event.OpTypeDel, configs)
	}
	return nil
}

func (svc *siteService) GetCertificates(siteID int64, domain string) ([]*model.CertCandidateResp, error) {
	domain = strings.TrimSpace(domain)
	domain = strings.ToLower(domain)
	domain = strings.TrimRight(domain, ".")

	certEntities := make([]*model.CertModel, 0)
	certCandidateResp := make([]*model.CertCandidateResp, 0)

	if len(domain) == 0 {
		if siteID <= 0 {
			return nil, ecode.ErrParam
		}
		siteEntity := new(model.SiteModel)
		if err := svc.repo.DB.Where("id = ?", siteID).First(&siteEntity).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ecode.ErrSiteNotFound
			}
			logx.Error("[site]Failed to obtain the domain name of the site: ", err.Error())
			return nil, ecode.InternalServerError
		}
		domain = siteEntity.Domain
	}

	_labels := strings.Split(domain, ".")
	_labels[0] = "*"
	wildcard := strings.Join(_labels, ".")

	if err := svc.repo.DB.Where("(? = ANY(sans)) OR (? = ANY(sans))", domain, wildcard).Find(&certEntities).Error; err != nil {
		logx.Error("[site]Failed to get certificates: ", err.Error())
		return nil, ecode.InternalServerError
	}

	if err := copier.Copy(&certCandidateResp, &certEntities); err != nil {
		logx.Error("[site]Failed to copy 'CertCandidateResp' entities: ", err.Error())
		return nil, ecode.InternalServerError
	}
	return certCandidateResp, nil

}

func (svc *siteService) assembleSiteGuardRsp(siteID int64) (*model.SiteGuardRsp, error) {
	siteEntity := new(model.SiteModel)
	siteConfigEntity := new(model.SiteConfigModel)
	certEntity := new(model.CertModel)
	siteOriginEntities := make([]*model.SiteOriginModel, 0)
	if err := svc.repo.DB.Where("id = ?", siteID).First(siteEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrIDNotFound
		} else {
			return nil, ecode.InternalServerError
		}
	}

	if err := svc.repo.DB.Where("site_id = ?", siteID).First(siteConfigEntity).Error; err != nil {
		return nil, ecode.InternalServerError
	}

	if err := svc.repo.DB.Where("site_id = ?", siteID).Find(&siteOriginEntities).Error; err != nil {
		return nil, ecode.InternalServerError
	}

	siteGuardRsp := new(model.SiteGuardRsp)
	siteGuardRsp.ID = siteEntity.UUID
	siteGuardRsp.Host = siteEntity.Domain

	_siteConfigGuard := new(model.SiteConfigGuard)
	if err := copier.Copy(_siteConfigGuard, siteConfigEntity); err != nil {
		logx.Error("[site] Failed to copy site guard config: ", err)
		return nil, err
	}

	if siteConfigEntity.CertID == 0 {
		_siteConfigGuard.CertID = ""
	} else {
		if err := svc.repo.DB.Where("id =?", siteConfigEntity.CertID).First(certEntity).Error; err != nil {
			logx.Error("[site] get cert uuid err: ", err)
			_siteConfigGuard.CertID = ""
		}
		_siteConfigGuard.CertID = certEntity.UUID
	}

	siteGuardRsp.Config = _siteConfigGuard

	_siteOriginsGuard := make([]*model.SiteOriginGuard, 0)
	copier.Copy(&_siteOriginsGuard, &siteOriginEntities)
	siteGuardRsp.Origins = _siteOriginsGuard

	return siteGuardRsp, nil
}

func (svc *siteService) GetRegionBlacklist(siteID int64) (*model.SiteRegionBlacklistRsp, error) {
	ety := new(model.SiteRegionBlacklistModel)
	if err := svc.repo.DB.Where("site_id = ?", siteID).First(&ety).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("站点不存在")
		}
		logx.Error("[site]Failed to get region blacklist: ", err)
		return nil, ecode.InternalServerError
	}

	rsp := new(model.SiteRegionBlacklistRsp)

	if err := copier.Copy(rsp, ety); err != nil {
		logx.Error("[site]Failed to copy struct: ", err)
		return nil, ecode.InternalServerError
	}
	return rsp, nil

}

func (svc *siteService) UpdateRegionBlacklist(siteID int64, payload *model.SiteRegionBlacklistModel) error {
	if err := svc.repo.DB.Where("site_id = ?", siteID).
		Select("countries", "regions", "match_mode").
		Updates(payload).Error; err != nil {
		logx.Error("[site]Failed to update region blacklist: ", err)
		return ecode.InternalServerError
	}

	//update guard
	{
		ety := new(model.SiteRegionBlacklistModel)
		if err := svc.repo.DB.Where("site_id = ?", siteID).
			Select("uuid", "countries", "regions", "match_mode").
			First(ety).Error; err != nil {
			logx.Error("[site]Failed to get site region blacklist: ", err)
			return ecode.InternalServerError
		}
		cfg := new(model.SiteRegionBlacklistGuard)
		if err := copier.Copy(cfg, ety); err != nil {
			logx.Error("[site]Failed to copy struct: ", err)
			return ecode.InternalServerError
		}
		if siteUUID, err := svc.getSiteUUID(siteID); err != nil {
			return err
		} else {
			cfg.SiteID = siteUUID
		}

		configs := make(model.GuardArrayRsp, 0)
		configs = append(configs, cfg)
		svc.eventBus.PushEvent(event.BlacklistRegion, event.OpTypeUpdate, configs)

	}

	return nil
}

func (svc *siteService) getSiteUUID(siteID int64) (string, error) {
	var uuid string
	if err := svc.repo.DB.Table("site").
		Where("id = ?", siteID).
		Select("uuid").
		Scan(&uuid).Error; err != nil {
		logx.Error("[site]Failed to get site uuid: ", err)
		return "", ecode.InternalServerError
	}
	return uuid, nil
}
