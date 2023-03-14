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
	Add(entity *model.RateLimitModel) error
	List(pageNum, pageSize, siteID int64, status, matchMode int16, url string) (*model.ListResp, error)
	UpdateStatus(id int64) error
	Delete(id int64) error
	Update(id int64, m *model.RateLimitModel) error
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

func (svc *rateLimitService) Add(entity *model.RateLimitModel) error {
	//entity := model.RateLimitModel{
	//	SiteID: siteID,
	//	Path:   path,
	//	Limit:  limit,
	//	Window: window,
	//	Status: 1,
	//	Remark: remark,
	//}
	entity.ID = 0
	if err := svc.repo.DB.Create(entity).Error; err != nil {
		return err
	}

	// update guard
	{
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

	for _, entity := range entities {
		domain, _ := svc.repo.Site.GetDomain(entity.SiteID)
		entity.Domain = domain
	}

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

	if err := svc.repo.DB.Where("id = ?", id).
		Select("SiteID", "Path", "Limit", "Window", "MatchMode", "Remark").
		Updates(m).
		Error; err != nil {
		return err
	}

	// update guard
	{
		_ = svc.repo.DB.Where("id = ?", id).First(entity).Error
		siteEntity := model.SiteModel{}
		if err := svc.repo.DB.Where("id = ?", m.SiteID).First(&siteEntity).Error; err == nil {
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
			svc.eventBus.PushEvent(event.RateLimit, event.OpTypeUpdate, configs)
		}
	}

	return nil
}
