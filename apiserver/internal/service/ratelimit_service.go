package service

import (
	"errors"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/event"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/repository"
	"github.com/xnile/muxwaf/pkg/logx"
	"github.com/xnile/muxwaf/pkg/utils"
	"gorm.io/gorm"
)

type IRateLimitService interface {
	Add(payload *model.RateLimitReq) error
	List(pageNum, pageSize, siteID int64, status, matchMode int16, url string) (*model.ListResp, error)
	UpdateStatus(id int64) error
	Delete(id int64) error
	Update(id int64, m *model.RateLimitModel) error
	BatchAdd(payload []*model.RateLimitReq) error
}

type rateLimitService struct {
	repo     *repository.Repository
	eventBus *event.EventBus
}

func NewRateLimitService(repo *repository.Repository, eventBus *event.EventBus) IRateLimitService {
	return &rateLimitService{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (svc *rateLimitService) Add(payload *model.RateLimitReq) error {
	var siteEty model.SiteModel
	{
		if err := svc.repo.DB.Select("ID", "UUID", "Domain").
			Where("id = ?", payload.SiteID).
			First(&siteEty).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ecode.ErrSiteNotFound
			}
			logx.Error("[RateLimit]Failed to get site info: ", err)
			return ecode.InternalServerError
		}
	}

	//payload.ID = 0
	//payload.Status = 1
	//payload.SiteUUID = siteEty.UUID
	//payload.Host = siteEty.Domain
	entity := model.RateLimitModel{
		SiteUUID:  siteEty.UUID,
		Host:      siteEty.Domain,
		Path:      payload.Path,
		Limit:     payload.Limit,
		Window:    payload.Window,
		MatchMode: payload.MatchMode,
		Status:    1,
		Remark:    payload.Remark,
	}

	if err := svc.repo.DB.Create(&entity).Error; err != nil {
		logx.Error("[RateLimit]Failed to create rate limit: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{

		configs := make(model.GuardArrayRsp, 0)
		{
			rateLimitGuard := model.RateLimitGuard{
				UUID:      entity.UUID,
				SiteID:    entity.SiteUUID,
				Host:      entity.Host,
				Path:      payload.Path,
				Limit:     payload.Limit,
				Window:    payload.Window,
				MatchMode: payload.MatchMode,
			}
			configs = append(configs, &rateLimitGuard)
		}

		//siteEntity := model.SiteModel{}
		//if err := svc.repo.DB.Where("id = ?", entity.SiteID).First(&siteEntity).Error; err == nil {
		//	configs := make(model.GuardArrayRsp, 0)
		//	config := map[string]any{
		//		"id":         entity.UUID,
		//		"path":       entity.Path,
		//		"limit":      entity.Limit,
		//		"window":     entity.Window,
		//		"match_mode": entity.MatchMode,
		//		"host":       siteEntity.Domain,
		//		"site_id":    siteEntity.UUID,
		//	}
		//	configs = append(configs, &config)
		svc.eventBus.PushEvent(event.RateLimit, event.OpTypeAdd, configs)
	}

	return nil

}

func (svc *rateLimitService) List(pageNum, pageSize, siteID int64, status, matchMode int16, url string) (*model.ListResp, error) {
	rsp := new(model.ListResp)
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)
	entities := make([]*model.RateLimitModel, 0)
	var count int64

	gDB := svc.repo.DB.Model(&model.RateLimitModel{})
	if siteID > 0 {
		gDB.Where("site_id = ?", siteID)
	}
	if status != -1 {
		gDB.Where("status = ?", status)
	}
	if matchMode > 0 {
		gDB.Where("match_mode = ?", matchMode)
	}
	if len(url) > 0 {
		gDB.Where("path LIKE ?", "%"+url+"%")
	}

	if err := gDB.Count(&count).Error; err != nil {
		logx.Error("[rateLimit] Failed to count rate limit: ", err)
		return nil, ecode.InternalServerError
	}
	if err := gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).
		Order("created_at DESC").Find(&entities).Error; err != nil {
		logx.Error("[rateLimit] Failed to get rate limit: ", err)
		return nil, ecode.InternalServerError
	}

	//for _, entity := range entities {
	//	domain, _ := svc.repo.Site.GetDomain(entity.SiteID)
	//	entity.Domain = domain
	//}

	rsp.SetValue(entities)
	rsp.SetMeta(pageSize, pageNum, count)
	return rsp, nil
}

func (svc *rateLimitService) UpdateStatus(id int64) error {
	if err := svc.repo.RateLimitRepo.UpdateStatus(id); err != nil {
		return err
	}
	// update guard
	{
		entity := model.RateLimitModel{}
		if err := svc.repo.DB.Where("id = ?", id).First(&entity).Error; err != nil {
			logx.Error("[guard_update] get rate limit err: ", err)
			return nil
		}

		if entity.Status == 0 {
			configs := make(model.GuardDelArrayRsp, 0)
			configs = append(configs, entity.UUID)
			svc.eventBus.PushEvent(event.RateLimit, event.OpTypeDel, configs)
		}

		if entity.Status == 1 {
			siteEntity := model.SiteModel{}
			if err := svc.repo.DB.Where("id = ?", entity.SiteID).First(&siteEntity).Error; err == nil {
				configs := make(model.GuardArrayRsp, 0)
				config := map[string]any{
					"id":         entity.UUID,
					"path":       entity.Path,
					"limit":      entity.Limit,
					"window":     entity.Window,
					"match_mode": entity.MatchMode,
					"host":       siteEntity.Domain,
					"site_id":    siteEntity.UUID,
				}
				configs = append(configs, &config)
				svc.eventBus.PushEvent(event.RateLimit, event.OpTypeAdd, configs)
			}
		}
	}
	return nil
}

func (svc *rateLimitService) Delete(id int64) error {
	entity := model.RateLimitModel{}
	if err := svc.repo.DB.Where("id = ?", id).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrIDNotFound
		} else {
			return ecode.InternalServerError
		}
	}

	if err := svc.repo.DB.Where("id = ?", id).Delete(&model.RateLimitModel{}).Error; err != nil {
		logx.Error("[rate_limit] delete rate limit err: ", err)
		return ecode.InternalServerError
	}
	// update guard
	{
		configs := make(model.GuardDelArrayRsp, 0)
		configs = append(configs, entity.UUID)
		svc.eventBus.PushEvent(event.RateLimit, event.OpTypeDel, configs)
	}
	return nil
}

func (svc *rateLimitService) Update(id int64, m *model.RateLimitModel) error {
	entity := new(model.RateLimitModel)
	if err := svc.repo.DB.Where("id = ?", id).First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrIDNotFound
		}
		logx.Error("[rate limit] searching rate limit err: ", err)
		return ecode.InternalServerError
	}

	siteEty, err := svc.getSiteDomainAndUUID(m.SiteID)
	if err != nil {
		logx.Error("[RateLimit]Failed to get site info: ", err)
		return err
	}

	// 同时更新SiteUUID和Host
	m.SiteUUID = siteEty.UUID
	m.Host = siteEty.Domain
	if err := svc.repo.DB.Where("id = ?", id).
		Select("SiteID", "Path", "Limit", "Window", "MatchMode", "Remark", "SiteUUID", "Host").
		Updates(m).
		Error; err != nil {
		logx.Error("[RateLimit]Failed to update rate limit: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{

		configs := make(model.GuardArrayRsp, 0)

		rateLimitGuard := model.RateLimitGuard{
			UUID:      entity.UUID,
			SiteID:    m.SiteUUID,
			Host:      m.Host,
			Path:      m.Path,
			Limit:     m.Limit,
			Window:    m.Window,
			MatchMode: m.MatchMode,
		}

		//_ = svc.repo.DB.Where("id = ?", id).First(entity).Error
		//siteEntity := model.SiteModel{}
		//if err := svc.repo.DB.Where("id = ?", m.SiteID).First(&siteEntity).Error; err == nil {
		//	configs := make(model.GuardArrayRsp, 0)
		//	config := map[string]any{
		//		"id":         entity.UUID,
		//		"path":       entity.Path,
		//		"limit":      entity.Limit,
		//		"window":     entity.Window,
		//		"match_mode": entity.MatchMode,
		//		"host":       siteEntity.Domain,
		//		"site_id":    siteEntity.UUID,
		//	}
		configs = append(configs, &rateLimitGuard)
		svc.eventBus.PushEvent(event.RateLimit, event.OpTypeUpdate, configs)

	}

	return nil
}

func (svc *rateLimitService) BatchAdd(payload []*model.RateLimitReq) error {
	entities := make([]*model.RateLimitModel, 0)
	guardCfg := make([]*model.RateLimitGuard, 0)

	tx := svc.repo.DB.Begin()

	for _, v := range payload {
		var siteEty model.SiteModel
		{
			if err := svc.repo.DB.Select("ID", "UUID", "Domain").
				Where("id = ?", v.SiteID).
				First(&siteEty).
				Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ecode.ErrSiteNotFound
				}
				logx.Error("[RateLimit]Failed to get site info: ", err)
				return ecode.InternalServerError
			}
		}

		entity := model.RateLimitModel{
			SiteID:    v.SiteID,
			SiteUUID:  siteEty.UUID,
			Host:      siteEty.Domain,
			Path:      v.Path,
			Limit:     v.Limit,
			Window:    v.Window,
			MatchMode: v.MatchMode,
			Status:    1,
			Remark:    v.Remark,
		}
		if err := tx.Create(&entity).Error; err != nil {
			logx.Error("[Rate Limit] Failed to insert : ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}
		entities = append(entities, &entity)
	}
	if err := tx.Commit().Error; err != nil {
		logx.Error("[Rate Limit] Failed to batch insert: ", err)
		return ecode.InternalServerError
	}

	// update guard
	for _, entity := range entities {
		siteEntity := model.SiteModel{}
		if err := svc.repo.DB.Where("id = ?", entity.SiteID).First(&siteEntity).Error; err != nil {
			logx.Error("[Rate Limit] Failed to fetching site info : ", err)
			return ecode.InternalServerError

		}
		guardCfg = append(guardCfg, &model.RateLimitGuard{
			UUID:      entity.UUID,
			SiteID:    siteEntity.UUID,
			Host:      siteEntity.Domain,
			Path:      entity.Path,
			Limit:     entity.Limit,
			Window:    entity.Window,
			MatchMode: entity.MatchMode,
		})
	}

	svc.eventBus.PushEvent(event.RateLimit, event.OpTypeAdd, guardCfg)

	return nil
}

func (svc *rateLimitService) getSiteDomainAndUUID(siteID int64) (*model.SiteModel, error) {
	var siteEty model.SiteModel
	{
		if err := svc.repo.DB.Select("ID", "UUID", "Domain").
			Where("id = ?", siteID).
			First(&siteEty).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ecode.ErrSiteNotFound
			}
			return nil, ecode.InternalServerError
		}
	}
	return &siteEty, nil
}
