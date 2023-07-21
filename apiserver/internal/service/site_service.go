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

	//GetConfigs(siteID int64) (*model.SiteConfigRsp, error)
	GetCertificates(siteID int64, domain string) ([]*model.CertCandidateResp, error)
	GetRegionBlacklist(id int64) (*model.SiteRegionBlacklistRsp, error)
	UpdateRegionBlacklist(id int64, region *model.SiteRegionBlacklistModel) error
	//GetSiteDomain(id int64) (string, error)
	// TODO
	//UpdateStatus(id int64) error

	UpdateOriginCfg(id int64, payload model.OriginCfgReq) error
	GetOriginCfg(siteID int64) (*model.OriginCfgRsp, error)
	UpdateBasicConfigs(siteID int64, payload *model.SiteBasicCfgReq) error
	UpdateHttpsConfigs(siteID int64, payload *model.SiteHttpsReq) error
	GetHttpsConfigs(siteID int64) (*model.SiteHttpsConfigsRsp, error)
	GetBasicHttps(siteID int64) (*model.SiteBasicConfigRsp, error)
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

func (svc *siteService) Add(payload *model.SiteReq) error {
	// 检查域名是否已经存在
	{
		sites := make([]*model.SiteModel, 0)
		if err := svc.repo.DB.Where("domain = ?", payload.Domain).Select("ID").Find(&sites).Error; err != nil {
			return ecode.InternalServerError
		}
		if len(sites) > 0 {
			return errors.New("域名已经存在")
		}
	}

	tx := svc.repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		logx.Error("[site]Failed to begin a transaction")
		return ecode.InternalServerError
	}

	var siteEty model.SiteModel

	{
		siteEty = model.SiteModel{

			Domain: payload.Domain,
			Status: 1,
		}

		if err := tx.Create(&siteEty).Error; err != nil {
			logx.Error("[site] add site err: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}
	}

	// 配置
	{
		siteCfgEty := model.SiteConfigModel{
			SiteID:             siteEty.ID,
			HttpPort:           80,
			HttpsPort:          443,
			IsHttps:            0,
			CertID:             0,
			IsRealIPFromHeader: 0,
			RealIPHeader:       "",
			IsForceHttps:       0,
			OriginHostHeader:   siteEty.Domain,
			OriginProtocol:     payload.OriginProtocol,
			SiteUUID:           siteEty.UUID,
			CertUUID:           "",
		}

		if err := tx.Create(&siteCfgEty).Error; err != nil {
			logx.Error("[site]Failed to create site config: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}
	}

	// 源站
	{
		var originEntities []model.SiteOriginModel
		{
			originEntities = make([]model.SiteOriginModel, 0)
			for _, origin := range payload.Origins {
				// 判断地址类型
				kind := model.IPOrigin
				if !utils.IsIPv4(origin.Addr) {
					kind = model.DomainOrigin
				}
				originEty := model.SiteOriginModel{
					SiteID:   siteEty.ID,
					Port:     origin.Port,
					Addr:     origin.Addr,
					Weight:   origin.Weight,
					Kind:     kind,
					Protocol: payload.OriginProtocol,
					SiteUUID: siteEty.UUID,
				}

				originEntities = append(originEntities, originEty)
			}
		}

		if err := tx.Create(&originEntities).Error; err != nil {
			logx.Error("[site]Failed to create site origins: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}

	}
	// 地域封禁
	{
		regionBlacklist := model.SiteRegionBlacklistModel{
			SiteID:    siteEty.ID,
			Countries: nil,
			Regions:   nil,
			MatchMode: 0,
			Status:    1,
			SiteUUID:  siteEty.UUID,
		}
		if err := tx.Create(&regionBlacklist).Error; err != nil {
			logx.Error("[site]Failed to add site region blacklist: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}
	}

	if err := tx.Commit().Error; err != nil {
		logx.Error("[site]Failed to commit create site: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(siteEty.ID, false)
		if err != nil {
			logx.Error("[GUARD]Failed to update guard: ", err)
			return nil
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeAdd, configs)
	}

	return nil
}

//func (svc *siteService) UpdateConfig(siteID int64, req *model.SiteConfigReq) error {
//	fields := make([]string, 0)
//	configEntity := new(model.SiteConfigModel)
//	if req.IsRealIPFromHeader != nil {
//		fields = append(fields, "IsRealIPFromHeader")
//		configEntity.IsRealIPFromHeader = *req.IsRealIPFromHeader
//	}
//	if req.RealIPHeader != nil {
//		fields = append(fields, "RealIPHeader")
//		configEntity.RealIPHeader = *req.RealIPHeader
//	}
//	if req.OriginProtocol != nil {
//		fields = append(fields, "OriginProtocol")
//		configEntity.OriginProtocol = *req.OriginProtocol
//	}
//
//	if err := svc.repo.DB.Model(&model.SiteConfigModel{}).Where("site_id = ?", siteID).
//		Select(fields).
//		Updates(configEntity).Error; err != nil {
//		logx.Error("[site] update site config err: ", err)
//		return ecode.InternalServerError
//	}
//
//	// guard update
//	{
//		configs := make(model.GuardArrayRsp, 0)
//		config, err := svc.assembleSiteGuardRsp(siteID)
//		if err != nil {
//			return err
//		}
//		configs = append(configs, &config)
//		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
//	}
//
//	return nil
//}

func (svc *siteService) UpdateBasicConfigs(siteID int64, payload *model.SiteBasicCfgReq) error {
	cfg := model.SiteConfigModel{
		IsRealIPFromHeader: *payload.IsRealIPFromHeader,
		RealIPHeader:       *payload.RealIPHeader,
	}
	if *payload.IsRealIPFromHeader == 1 && *payload.RealIPHeader == "" {
		return errors.New("参数有误，header不能为空")
	}

	if err := svc.repo.DB.Model(&model.SiteConfigModel{}).
		Select("IsRealIPFromHeader", "RealIPHeader").
		Where("site_id = ?", siteID).
		Updates(&cfg).
		Error; err != nil {
		logx.Error("[site]Failed to update site basic configs: ", err)
		return ecode.InternalServerError
	}

	// guard update
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(siteID, true)
		if err != nil {
			return err
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
	}
	return nil

}

func (svc *siteService) UpdateHttpsConfigs(siteID int64, payload *model.SiteHttpsReq) error {
	//configEntity := new(model.SiteConfigModel)
	//configEntity.IsHttps = *payload.IsHttps
	//if *payload.IsHttps == 0 {
	//	configEntity.CertID = 0
	//	configEntity.IsForceHttps = 0
	//} else {
	//	configEntity.CertID = *payload.CertID
	//	configEntity.IsForceHttps = *payload.IsForceHttps
	//}

	var (
		isHttps      int8  = 0
		certID       int64 = 0
		isForceHttps int8  = 0
		certUUID           = ""
	)

	if *payload.IsHttps != 0 {
		if *payload.CertID < 1 {
			return errors.New("请选择证书")
		}
		isHttps = *payload.IsHttps
		certID = *payload.CertID
		isForceHttps = *payload.IsForceHttps

		// 更新UUID
		if err := svc.repo.DB.Table("cert").Where("id = ?", certID).Select("UUID").Scan(&certUUID).Error; err != nil {
			logx.Error("[site]Failed to get cert uuid: ", err)
			return ecode.InternalServerError
		}
	}

	configEntity := model.SiteConfigModel{
		IsHttps:      isHttps,
		CertID:       certID,
		IsForceHttps: isForceHttps,
		CertUUID:     certUUID,
	}

	if err := svc.repo.DB.Model(&model.SiteConfigModel{}).Where("site_id = ?", siteID).
		Select("CertID", "IsHttps", "IsForceHttps", "CertUUID").
		Updates(&configEntity).
		Error; err != nil {
		logx.Error("[site] update https config err: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(siteID, true)
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

//func (svc *siteService) GetOrigins(id int64) []*model.SiteOriginRsp {
//	r := make([]*model.SiteOriginRsp, 0)
//	origins, err := svc.repo.SiteOriginRepo.GetBySiteID(id)
//	copier.Copy(&r, &origins)
//	if err != nil {
//		logx.Error("get site origins err: ", err)
//		return nil
//	}
//	return r
//
//}

//func (svc *siteService) AddSiteOrigins(id int64, origins []*model.SiteOriginModel) error {
//	for _, origin := range origins {
//		if err := svc.repo.SiteOriginRepo.Insert(&model.SiteOriginModel{
//			SiteID:   id,
//			Port:     origin.Port,
//			Weight:   origin.Weight,
//			Addr:     origin.Addr,
//			Protocol: model.HTTPOriginProtocol,
//		}); err != nil {
//			logx.Error("add site origins err: ", err)
//			return ecode.InternalServerError
//		}
//	}
//
//	// update guard
//	{
//		configs := make(model.GuardArrayRsp, 0)
//		config, err := svc.assembleSiteGuardRsp(id)
//		if err != nil {
//			return err
//		}
//		configs = append(configs, &config)
//		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
//	}
//	return nil
//}

//func (svc *siteService) UpdateOrigin(id int64, originModel *model.SiteOriginModel) error {
//	if err := svc.repo.DB.Model(&model.SiteOriginModel{}).Where("id = ?", id).
//		Select("HttpPort", "Weight", "Host").
//		Updates(originModel).Error; err != nil {
//		//	TODO log
//		return ecode.InternalServerError
//	}
//
//	originEntity := new(model.SiteOriginModel)
//	_ = svc.repo.DB.Where("id = ?", id).First(originEntity).Error
//
//	// update guard
//	{
//		configs := make(model.GuardArrayRsp, 0)
//		config, err := svc.assembleSiteGuardRsp(originEntity.SiteID)
//		if err != nil {
//			return err
//		}
//		configs = append(configs, &config)
//		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
//	}
//	return nil
//}

//func (svc *siteService) DelOrigin(id int64) error {
//	origin := new(model.SiteOriginModel)
//	if err := svc.repo.DB.Where("id = ?", id).Select("SiteID", "UUID").First(origin).Error; err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return ecode.ErrIDNotFound
//		} else {
//			logx.Error("[site] fetching site origin err: ", err)
//			return ecode.InternalServerError
//		}
//	}
//
//	origins := make([]*model.SiteOriginModel, 0)
//	if err := svc.repo.DB.Where("site_id = ?", origin.SiteID).Select("ID").Find(&origins).Error; err != nil {
//		logx.Error("[site] searching site origins err: ", err)
//		return ecode.InternalServerError
//	}
//	if len(origins) < 2 {
//		return ecode.ErrAtLeastOneOrigin
//	}
//
//	if err := svc.repo.DB.Where("id = ?", id).Delete(&model.SiteOriginModel{}).Error; err != nil {
//		//	TODO log
//		return ecode.InternalServerError
//	}
//
//	// update guard
//	{
//		configs := make(model.GuardArrayRsp, 0)
//		config, err := svc.assembleSiteGuardRsp(origin.SiteID)
//		if err != nil {
//			return err
//		}
//		configs = append(configs, &config)
//		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
//	}
//	return nil
//}

func (svc *siteService) GetHttpsConfigs(siteID int64) (*model.SiteHttpsConfigsRsp, error) {
	if err := svc.isSiteExist(siteID); err != nil {
		return nil, err
	}

	var configsEty model.SiteConfigModel
	{
		if err := svc.repo.DB.Where("site_id = ?", siteID).
			Select("IsHttps", "CertID", "IsForceHttps").
			First(&configsEty).Error; err != nil {
			logx.Error("[site]Failed to get site configs: ", err)
			return nil, ecode.InternalServerError
		}
	}

	var certEty model.CertModel
	{
		if configsEty.CertID > 0 {
			if err := svc.repo.DB.
				Where("id = ?", configsEty.CertID).
				Select("Name").
				First(&certEty).
				Error; err != nil {
				logx.Error("[site]Failed to get certificate: ", err)
				return nil, ecode.InternalServerError
			}
		}
	}

	rsp := model.SiteHttpsConfigsRsp{
		IsHttps:      configsEty.IsHttps,
		CertName:     certEty.Name,
		IsForceHttps: configsEty.IsForceHttps,
		CertID:       configsEty.CertID,
	}
	return &rsp, nil
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

	// 删除关联资源
	tx := svc.repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		logx.Error("[site]Failed to create transaction: ", err)
		return ecode.InternalServerError
	}
	{
		// 删除站点
		if err = tx.Where("id = ?", id).Delete(&model.SiteModel{}).Error; err != nil {
			logx.Error("[site] delete site err: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}
		//删除站点配置
		if err = tx.Where("site_id = ?", id).Delete(&model.SiteConfigModel{}).Error; err != nil {
			logx.Error("[site] delete site config err: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}
		//删除站点源站
		if err = tx.Where("site_id = ?", id).Delete(&model.SiteOriginModel{}).Error; err != nil {
			logx.Error("[site] delete site origins err: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}

		//删除站点地域级IP黑名单
		if err = tx.Where("site_id = ?", id).Delete(&model.SiteRegionBlacklistModel{}).Error; err != nil {
			logx.Error("[site]Failed to delete site region IP blacklist: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}
	}

	if err := tx.Commit().Error; err != nil {
		logx.Error("[site]Failed to commit transaction: ", err)
		tx.Rollback()
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

func (svc *siteService) assembleSiteGuardRsp(siteID int64, ignoreOriginCfg bool) (*model.SiteGuard, error) {
	siteEntity := new(model.SiteModel)
	siteConfigEntity := new(model.SiteConfigModel)
	//certEntity := new(model.CertModel)

	//var siteOriginCfgGuar model.SiteOriginCfgGuard

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

	// 源站配置
	//siteOriginCfgGuard := model.SiteOriginCfgGuard{
	//	OriginProtocol:   siteConfigEntity.OriginProtocol,
	//	OriginHostHeader: siteConfigEntity.OriginHostHeader,
	//	Origins:          _siteOriginsGuard,
	//}

	// 站点
	siteGuardRsp := new(model.SiteGuard)
	siteGuardRsp.UUID = siteEntity.UUID
	siteGuardRsp.Host = siteEntity.Domain
	siteGuardRsp.Configs = nil

	// 站点配置
	_siteConfigGuard := new(model.SiteConfigGuard)
	if err := copier.Copy(_siteConfigGuard, siteConfigEntity); err != nil {
		logx.Error("[site] Failed to copy site guard config: ", err)
		return nil, err
	}
	_siteConfigGuard.CertID = siteConfigEntity.CertUUID

	_siteConfigGuard.Origin = nil
	if !ignoreOriginCfg {
		siteOriginEntities := make([]*model.SiteOriginModel, 0)
		if err := svc.repo.DB.Where("site_id = ?", siteID).Find(&siteOriginEntities).Error; err != nil {
			return nil, ecode.InternalServerError
		}

		// 源站
		_siteOriginsGuard := make([]*model.SiteOriginGuard, 0)
		if err := copier.Copy(&_siteOriginsGuard, &siteOriginEntities); err != nil {
			logx.Error("[site]Failed to copy site origins")
			return nil, ecode.InternalServerError
		}
		// 源站配置
		siteOriginCfgGuard := model.SiteOriginCfgGuard{
			OriginProtocol:   siteConfigEntity.OriginProtocol,
			OriginHostHeader: siteConfigEntity.OriginHostHeader,
			Origins:          _siteOriginsGuard,
		}
		_siteConfigGuard.Origin = &siteOriginCfgGuard
	}

	//if siteConfigEntity.CertID == 0 {
	//	_siteConfigGuard.CertID = ""
	//} else {
	//	if err := svc.repo.DB.Where("id =?", siteConfigEntity.CertID).First(certEntity).Error; err != nil {
	//		logx.Error("[site] get cert uuid err: ", err)
	//		_siteConfigGuard.CertID = ""
	//	}
	//	_siteConfigGuard.CertID = certEntity.UUID
	//}

	siteGuardRsp.Configs = _siteConfigGuard

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
			logx.Error("[site]Failed to get site UUID: ", err)
			return ecode.InternalServerError
		} else {
			cfg.SiteID = siteUUID
		}

		configs := make(model.GuardArrayRsp, 0)
		configs = append(configs, cfg)
		svc.eventBus.PushEvent(event.BlacklistRegion, event.OpTypeUpdate, configs)

	}

	return nil
}

//func (svc *siteService) getSiteUUID(siteID int64) (string, error) {
//	var uuid string
//	if err := svc.repo.DB.Table("site").
//		Where("id = ?", siteID).
//		Select("uuid").
//		Scan(&uuid).Error; err != nil {
//		logx.Error("[site]Failed to get site uuid: ", err)
//		return "", ecode.InternalServerError
//	}
//	return uuid, nil
//}

func (svc *siteService) GetSiteDomain(id int64) (string, error) {
	var domain string
	if err := svc.repo.DB.Model(&model.SiteModel{}).Where("id = ?", id).Select("domain").Scan(&domain).Error; err != nil {
		logx.Error("[site]Failed to get site domain: ", err)
		return "", ecode.InternalServerError
	}
	return domain, nil
}

func (svc *siteService) isSiteExist(siteID int64) error {
	if err := svc.repo.DB.Where("id = ?", siteID).Select("id").First(&model.SiteModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrSiteNotFound
		}
		logx.Error("[site]Failed to get site: ", err)
		return ecode.InternalServerError
	}
	return nil
}

func (svc *siteService) UpdateOriginCfg(siteID int64, payload model.OriginCfgReq) error {
	if err := svc.isSiteExist(siteID); err != nil {
		return err
	}
	oldOrigins := make([]*model.SiteOriginModel, 0)
	if err := svc.repo.DB.Where("site_id = ?", siteID).Select("id").Find(&oldOrigins).Error; err != nil {
		logx.Error("[site]Failed to get site origins: ", err)
		return ecode.InternalServerError
	}

	siteCfg := model.SiteConfigModel{
		OriginHostHeader: payload.OriginHostHeader,
		OriginProtocol:   payload.OriginProtocol,
	}

	tx := svc.repo.DB.Begin()

	if err := tx.Model(&model.SiteConfigModel{}).
		Where("site_id = ?", siteID).
		Select("OriginHostHeader", "OriginProtocol").
		Updates(siteCfg).
		Error; err != nil {
		tx.Rollback()
		logx.Error("[site]Failed to update site config: ", err)
		return ecode.InternalServerError
	}

	//	删除
nextDel:
	for _, oldOrigin := range oldOrigins {
		for _, newOrigin := range payload.Origins {
			if oldOrigin.ID == newOrigin.ID {
				continue nextDel
			}
		}
		if err := tx.Where("id = ?", oldOrigin.ID).Delete(&model.SiteOriginModel{}).Error; err != nil {
			logx.Error("[site]Failed to delete origin: ", err)
			return ecode.InternalServerError
		}
	}

nextAddOrUpdate:
	for _, newOrigin := range payload.Origins {
		kind := model.IPOrigin
		if !utils.IsIPv4(newOrigin.Addr) {
			kind = model.DomainOrigin
		}

		siteUUID, err := svc.getSiteUUID(siteID)
		if err != nil {
			logx.Error("[site]Failed to get site UUID: ", err)
			return ecode.InternalServerError
		}

		o := model.SiteOriginModel{
			SiteID:   siteID,
			Port:     newOrigin.Port,
			Addr:     newOrigin.Addr,
			Weight:   newOrigin.Weight,
			Kind:     kind,
			Protocol: payload.OriginProtocol,
			SiteUUID: siteUUID,
		}
		for _, oldOrigin := range oldOrigins {
			if newOrigin.ID == oldOrigin.ID {
				if err := tx.Where("id = ?", oldOrigin.ID).
					Select("Port", "Addr", "Weight", "Kind", "Protocol").
					Updates(&o).Error; err != nil {
					tx.Rollback()
					logx.Error("[site]Failed to update origin: ", err)
					return ecode.InternalServerError
				}
				continue nextAddOrUpdate
			}
		}

		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			logx.Error("[site]Failed to insert origin: ", err)
			return ecode.InternalServerError
		}
	}
	if err := tx.Commit().Error; err != nil {
		logx.Error("[site]Failed to commit transaction : ", err)
		return ecode.InternalServerError
	}

	// update guard
	{
		configs := make(model.GuardArrayRsp, 0)
		config, err := svc.assembleSiteGuardRsp(siteID, false)
		if err != nil {
			return err
		}
		configs = append(configs, &config)
		svc.eventBus.PushEvent(event.Site, event.OpTypeUpdate, configs)
	}
	return nil
}

func (svc *siteService) GetOriginCfg(siteID int64) (*model.OriginCfgRsp, error) {
	if err := svc.isSiteExist(siteID); err != nil {
		return nil, err
	}

	siteCfg := new(model.SiteConfigModel)
	origins := make([]*model.SiteOriginModel, 0)

	if err := svc.repo.DB.Where("site_id = ?", siteID).
		Select("OriginProtocol", "OriginHostHeader").First(siteCfg).Error; err != nil {
		logx.Error("[site]Failed to get site configs: ", err)
		return nil, ecode.InternalServerError
	}

	if err := svc.repo.DB.Where("site_id = ?", siteID).
		Omit("Kind", "Protocol").
		Find(&origins).
		Error; err != nil {
		logx.Error("[site]Failed to get site origins: ", err)
		return nil, ecode.InternalServerError
	}

	rsp := model.OriginCfgRsp{
		OriginProtocol:   siteCfg.OriginProtocol,
		OriginHostHeader: siteCfg.OriginHostHeader,
		Origins:          origins,
	}
	return &rsp, nil
}

func (svc *siteService) GetBasicHttps(siteID int64) (*model.SiteBasicConfigRsp, error) {
	var siteEntity model.SiteModel
	if err := svc.repo.DB.Where("id = ?", siteID).Select("Domain").First(&siteEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrSiteNotFound
		}
		logx.Error("[site]Failed to get site: ", err)
		return nil, ecode.InternalServerError
	}

	var configEntity model.SiteConfigModel
	if err := svc.repo.DB.Where("site_id = ?", siteID).Select("IsRealIPFromHeader", "RealIPHeader").First(&configEntity).Error; err != nil {
		logx.Error("[ste] get site config err: ", err)
		return nil, ecode.InternalServerError
	}

	return &model.SiteBasicConfigRsp{
		Host:               siteEntity.Domain,
		IsRealIPFromHeader: configEntity.IsRealIPFromHeader,
		RealIPHeader:       configEntity.RealIPHeader,
	}, nil
}

func (svc *siteService) getSiteUUID(siteID int64) (string, error) {
	var uuid string
	if err := svc.repo.DB.Table("site").Where("id = ?", siteID).Select("UUID").Scan(&uuid).Error; err != nil {
		return "", err
	}
	return uuid, nil
}
