package service

import (
	"errors"
	"github.com/jinzhu/copier"
	"github.com/rs/xid"
	"github.com/spf13/viper"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/event"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/pkg/logx"
	"github.com/xnile/muxwaf/pkg/utils"
	"gorm.io/gorm"
	"strings"
)

const sampleLogAPIURL = "/api/logs/sample"

type INodeService interface {
	Add(payload *model.NodeModel) error
	List(pageNum, pageSize int64) (*model.ListResp, error)
	Sync(nodeID int64) error
	Del(id int64) error
	SwitchSampleLogUpload(id int64) error
	SwitchStatus(id int64) error
}

type nodeService struct {
	gDB      *gorm.DB
	eventBus *event.EventBus
}

func NewINodeService(db *gorm.DB, eventBus *event.EventBus) INodeService {
	return &nodeService{
		gDB:      db,
		eventBus: eventBus,
	}
}

func (svc *nodeService) Add(payload *model.NodeModel) error {
	//node := new(model.NodeModel)
	if err := svc.gDB.Where("ip_or_domain = ? and port = ?", payload.Addr, payload.Port).First(new(model.NodeModel)).Error; err == nil {
		return errors.New("节点已经存在")
	}

	payload.SampleLogUploadAPIToken = strings.ToUpper(xid.New().String())
	if err := svc.gDB.Create(payload).Error; err != nil {
		return ecode.InternalServerError
	}
	svc.Sync(payload.ID)
	return nil
}

func (svc *nodeService) Del(id int64) error {
	if err := svc.gDB.Delete(&model.NodeModel{}, id).Error; err != nil {
		logx.Error("[node] Failed to delete node: ", err)
		return ecode.InternalServerError
	}
	return nil
}

func (svc *nodeService) List(pageNum, pageSize int64) (*model.ListResp, error) {
	rsp := new(model.ListResp)
	entities := make([]*model.NodeModel, 0)
	var count int64
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)

	gDB := svc.gDB.Model(&model.NodeModel{})
	if err := gDB.Count(&count).Error; err != nil {
		logx.Error("[node] get nodes err: ", err)
		return nil, ecode.InternalServerError
	}
	gDB = gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize))
	if err := gDB.Order("created_at DESC").Find(&entities).Error; err != nil {
		logx.Error("[node] get nodes err: ", err)
		return nil, ecode.InternalServerError
	}

	rsp.SetValue(entities)
	rsp.SetMeta(pageSize, pageNum, count)
	return rsp, nil
}

func (svc *nodeService) SwitchSampleLogUpload(id int64) error {
	entity := new(model.NodeModel)

	if err := svc.gDB.Where("id = ?", id).First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("节点不存在")
		}
		return ecode.InternalServerError
	}

	if err := svc.gDB.Model(&model.NodeModel{}).Where("id = ?", id).UpdateColumn("IsSampleLogUpload", gorm.Expr("ABS(is_sample_log_upload - ?)", 1)).Error; err != nil {
		logx.Error("[node] Failed to update sampled log upload status: ", err.Error())
		return ecode.InternalServerError
	}

	// update guard
	{
		sampledLogUploadApi := "http://" + utils.GetOutboundIP().String() + ":" + viper.GetString("port") + sampleLogAPIURL
		cfg := model.SampleLogUploadGuard{
			IsSampleLogUpload:       0,
			SampleLogUploadAPI:      "",
			SampleLogUploadAPIToken: "",
		}
		if entity.IsSampleLogUpload == 0 {
			cfg.IsSampleLogUpload = 1
			//cfg.SampleLogUploadAPI = "http://" + utils.GetOutboundIP().String() + ":" + viper.GetString("port") + "/logs/sampled"
			cfg.SampleLogUploadAPI = sampledLogUploadApi
			cfg.SampleLogUploadAPIToken = entity.SampleLogUploadAPIToken
		}
		svc.eventBus.PushEvent(event.SampleLogUpload, event.OpTypeUpdate, cfg, id)
	}
	return nil
}

func (svc *nodeService) Sync(id int64) error {
	// k,v
	sitesCache := make(map[int64]*model.SiteModel)
	OriginsCache := make(map[int64][]*model.SiteOriginModel)
	siteConfigsCache := make(map[int64]*model.SiteConfigModel)
	certsCache := make(map[int64]*model.CertModel)

	// 实体
	// Guard节点
	nodeEntity := new(model.NodeModel)
	// IP黑名单
	blacklistIPEntities := make([]*model.BlacklistIPModel, 0)
	// IP白名单
	whitelistIPEntities := make([]*model.WhitelistIPModel, 0)
	// URL白名单
	whitelistURLEntities := make([]*model.WhitelistURLModel, 0)
	// 频率限制
	rateLimitEntities := make([]*model.RateLimitModel, 0)
	// 证书
	certificateEntities := make([]*model.CertModel, 0)
	// 网站
	siteEntities := make([]*model.SiteModel, 0)
	// 网站地域级IP黑名单
	siteRegionBlacklistEntities := make([]*model.SiteRegionBlacklistModel, 0)

	// 网站配置
	siteConfigEntities := make([]*model.SiteConfigModel, 0)
	// 网站源站
	siteOriginEntities := make([]*model.SiteOriginModel, 0)

	// 数据库查询
	{
		if err := svc.gDB.Where("id = ?", id).First(nodeEntity).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("节点不存在")
			}
			return ecode.InternalServerError
		}

		if nodeEntity.Status == 0 {
			return errors.New("节点当前处于禁用状态，请先启用节点后再同步")
		}

		if err := svc.gDB.Where("status = ?", 1).Find(&blacklistIPEntities).Error; err != nil {
			logx.Error("[node] Failed to get ip blacklist: ", err.Error())
			return ecode.InternalServerError
		}

		if err := svc.gDB.Where("status = ?", 1).Find(&siteRegionBlacklistEntities).Error; err != nil {
			logx.Error("[node] Failed to get region blacklist: ", err.Error())
			return ecode.InternalServerError
		}

		if err := svc.gDB.Where("status = ?", 1).Find(&whitelistIPEntities).Error; err != nil {
			logx.Error("[node] Failed to get ip whitelist: ", err.Error())
			return ecode.InternalServerError
		}

		if err := svc.gDB.Where("status = ?", 1).Find(&whitelistURLEntities).Error; err != nil {
			logx.Error("[node] Failed to get url whitelist: ", err.Error())
			return ecode.InternalServerError
		}

		if err := svc.gDB.Where("status = ?", 1).Find(&rateLimitEntities).Error; err != nil {
			logx.Error("[node] Failed to get rate limit: ", err.Error())
			return ecode.InternalServerError
		}

		if err := svc.gDB.Find(&certificateEntities).Error; err != nil {
			logx.Error("[node] Failed to get certificates: ", err.Error())
			return ecode.InternalServerError
		}

		if err := svc.gDB.Where("status = ?", 1).Find(&siteEntities).Error; err != nil {
			logx.Error("[node] Failed to get sites: ", err.Error())
			return ecode.InternalServerError
		}

		if err := svc.gDB.Find(&siteConfigEntities).Error; err != nil {
			logx.Error("[node] Failed to get site config entities: ", err.Error())
			return ecode.InternalServerError
		}

		if err := svc.gDB.Find(&siteOriginEntities).Error; err != nil {
			logx.Error("[node] Failed to get site origin entities: ", err.Error())
			return ecode.InternalServerError
		}
	}

	// k,v处理，方便使用
	{
		for _, site := range siteEntities {
			sitesCache[site.ID] = site
		}

		for _, cert := range certificateEntities {
			certsCache[cert.ID] = cert
		}

		for _, origin := range siteOriginEntities {
			siteID := origin.SiteID
			if OriginsCache[siteID] == nil {
				OriginsCache[siteID] = make([]*model.SiteOriginModel, 0)
			}
			OriginsCache[siteID] = append(OriginsCache[siteID], origin)
		}

		for _, cfg := range siteConfigEntities {
			siteConfigsCache[cfg.SiteID] = cfg
		}
	}

	// 准备组装ConfigsSyncGuard需要的数据类型
	sampleLogCfgGuard := new(model.SampleLogUploadGuard)
	arrayBlacklistIPGuard := make([]*model.BlacklistIPGuard, 0)
	arrayWhitelistIPGuard := make([]*model.WhitelistIPGuard, 0)
	arrayCertificateGuard := make([]*model.CertificateGuard, 0)

	//_arraySiteOriginGuard := make([]*model.SiteOriginGuard, 0)

	// 类型转换
	{
		// 拦截日志
		if err := copier.Copy(sampleLogCfgGuard, nodeEntity); err != nil {
			logx.Error("[node] Failed to copy sampledLogUploadGuard: ", 0)
			return ecode.InternalServerError
		}

		// IP黑名单
		if err := copier.Copy(&arrayBlacklistIPGuard, &blacklistIPEntities); err != nil {
			logx.Error("[node] Failed to copy arrayBlacklistIPGuard: ", 0)
			return ecode.InternalServerError
		}
		//if err := copier.Copy(&arrayBlacklistRegionGuard, &blacklistRegionEntities); err != nil {
		//	logx.Error("[node] Failed to copy arrayBlacklistRegionGuard: ", 0)
		//	return ecode.InternalServerError
		//}
		// IP白名单
		if err := copier.Copy(&arrayWhitelistIPGuard, &whitelistIPEntities); err != nil {
			logx.Error("[node] Failed to copy arrayWhitelistIPGuard: ", 0)
			return ecode.InternalServerError
		}

		// 证书，只同步在使用的证书
		{
			inUseCertificates := make([]*model.CertModel, 0)
			for _, site := range siteEntities {
				cfg := siteConfigsCache[site.ID]
				if cfg == nil {
					logx.Error("[node]Site configuration not found")
					return ecode.InternalServerError
				}

				certID := cfg.CertID
				if certID < 1 {
					// 站点未启用HTTPS
					continue
				}
				candidate := certsCache[certID]
				if candidate == nil {
					logx.Error("[node]Certificate not found")
					continue
				}
				inUseCertificates = append(inUseCertificates, candidate)
			}
			if err := copier.Copy(&arrayCertificateGuard, &inUseCertificates); err != nil {
				logx.Error("[node] Failed to copy arrayCertificateGuard: ", 0)
				return ecode.InternalServerError
			}
		}
	}

	arrayWhitelistURLGuard := make([]*model.WhitelistURLGuard, 0)
	arrayRateLimitGuard := make([]*model.RateLimitGuard, 0)
	arraySiteRegionBlacklistGuard := make([]*model.SiteRegionBlacklistGuard, 0)
	{
		// URL白名单列表
		{
			for _, whitelistURL := range whitelistURLEntities {
				whitelistGuard := model.WhitelistURLGuard{
					UUID:      whitelistURL.UUID,
					SiteID:    sitesCache[whitelistURL.SiteID].UUID,
					Host:      sitesCache[whitelistURL.SiteID].Domain,
					Path:      whitelistURL.Path,
					MatchMode: whitelistURL.MatchMode,
				}
				arrayWhitelistURLGuard = append(arrayWhitelistURLGuard, &whitelistGuard)
			}
		}

		// 频率限制列表
		{
			for _, rateLimit := range rateLimitEntities {
				rateLimitGuard := model.RateLimitGuard{
					UUID:      rateLimit.UUID,
					SiteID:    sitesCache[rateLimit.SiteID].UUID,
					Host:      sitesCache[rateLimit.SiteID].Domain,
					Path:      rateLimit.Path,
					Limit:     rateLimit.Limit,
					Window:    rateLimit.Window,
					MatchMode: rateLimit.MatchMode,
				}
				arrayRateLimitGuard = append(arrayRateLimitGuard, &rateLimitGuard)
			}
		}

		// 网站地域级IP黑名单
		{
			for _, siteRegionBlacklist := range siteRegionBlacklistEntities {
				siteRegionBlacklistGuard := model.SiteRegionBlacklistGuard{
					SiteID:    sitesCache[siteRegionBlacklist.SiteID].UUID,
					Countries: siteRegionBlacklist.Countries,
					Regions:   siteRegionBlacklist.Regions,
					MatchMode: siteRegionBlacklist.Status,
				}
				arraySiteRegionBlacklistGuard = append(arraySiteRegionBlacklistGuard, &siteRegionBlacklistGuard)
			}
		}
	}

	arraySiteGuard := make([]*model.SiteGuard, 0)
	for _, site := range siteEntities {
		siteID := site.ID
		cfg := siteConfigsCache[siteID]

		var siteGuard model.SiteGuard
		{
			var configsGuard model.SiteConfigGuard
			{
				var originCfgGuard model.SiteOriginCfgGuard
				{
					var arraySiteOriginGuard []*model.SiteOriginGuard
					{
						arraySiteOrigin := OriginsCache[siteID]
						arraySiteOriginGuard = make([]*model.SiteOriginGuard, 0)
						if err := copier.Copy(&arraySiteOriginGuard, &arraySiteOrigin); err != nil {
							logx.Error("[node] Failed to copy arraySiteOriginGuard: ", err)
							return ecode.InternalServerError
						}
					}
					originCfgGuard = model.SiteOriginCfgGuard{
						OriginProtocol:   cfg.OriginProtocol,
						OriginHostHeader: cfg.OriginHostHeader,
						Origins:          arraySiteOriginGuard,
					}
				}

				{
					cert, ok := certsCache[cfg.CertID]
					if !ok {
						logx.Error("[node]Certificate not found")
						return ecode.InternalServerError
					}

					configsGuard = model.SiteConfigGuard{
						CertID:             cert.UUID,
						IsHttps:            cfg.IsHttps,
						IsRealIPFromHeader: cfg.IsRealIPFromHeader,
						RealIPHeader:       cfg.RealIPHeader,
						Origin:             &originCfgGuard,
					}
				}
			}

			siteGuard = model.SiteGuard{
				UUID:    site.UUID,
				Host:    site.Domain,
				Configs: &configsGuard,
			}
		}

		arraySiteGuard = append(arraySiteGuard, &siteGuard)
	}

	//
	//{
	//	sampledLogUploadApi := "http://" + utils.GetOutboundIP().String() + ":" + viper.GetString("port") + sampleLogAPIURL
	//	sampleLogCfgGuard.SampleLogUploadAPI = sampledLogUploadApi
	//
	//}

	var guardConfigs model.ConfigsSyncGuard
	{
		var rules model.RulesGuard
		{
			rules = model.RulesGuard{
				WhitelistIP:     arrayWhitelistIPGuard,
				WhitelistURL:    arrayWhitelistURLGuard,
				BlacklistIP:     arrayBlacklistIPGuard,
				BlacklistRegion: arraySiteRegionBlacklistGuard,
				RateLimit:       arrayRateLimitGuard,
			}
		}

		guardConfigs = model.ConfigsSyncGuard{
			SampleLog:    sampleLogCfgGuard,
			Sites:        arraySiteGuard,
			Certificates: arrayCertificateGuard,
			Rules:        &rules,
		}
	}

	// update guard
	{
		//go func() {
		//	logx.Info("push event: ", id)
		//	svc.eventBus.PushEvent(event.All, event.OpTypeSync, guardConfigs, id)
		//}()
		svc.eventBus.PushEvent(event.All, event.OpTypeSync, guardConfigs, id)
	}
	return nil
}

func (svc *nodeService) SwitchStatus(id int64) error {
	if err := svc.gDB.Model(&model.NodeModel{}).Where("id = ?", id).UpdateColumn("status", gorm.Expr("ABS(status - ?)", 1)).Error; err != nil {
		logx.Error("[node] Failed to update node status: ", err.Error())
		return ecode.InternalServerError
	}
	return nil
}
