package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/event"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/repository"
	"github.com/xnile/muxwaf/pkg/logx"
	"github.com/xnile/muxwaf/pkg/utils"
	"go4.org/netipx"
	"gorm.io/gorm"
	"net/netip"
	"strings"
)

type IWhitelistService interface {
	AddIP(c *gin.Context) error
	AddURL(c *gin.Context, urlModel *model.WhitelistURLModel) error
	ListIP(pageNum, pageSize, startTime, endTime int64, ip string, status int16) (*model.ListResp, error)
	ListURL(pageNum, pageSize, siteID int64, status int16, url string) (*model.ListResp, error)
	UpdateIPStatus(id int64) error
	UpdateURLStatus(id int64) error
	DeleteIP(id int64) error
	DeleteURL(id int64) error
	UpdateURL(id int64, m *model.WhitelistURLModel) error
	UpdateIP(c *gin.Context, id int64, payload *model.WhitelistIPModel) error
	IsIpIncluded(ip string) (bool, error)
}

type whitelistService struct {
	gDB          *gorm.DB
	repo         *repository.Repository
	eventBus     *event.EventBus
	ipSetBuilder *netipx.IPSetBuilder
}

func NewWhitelistService(gDB *gorm.DB, repo *repository.Repository, eventBus *event.EventBus) IWhitelistService {
	var ipSet netipx.IPSetBuilder
	entities := make([]*model.WhitelistIPModel, 0)
	if err := gDB.Select("IP").Find(&entities).Error; err != nil {
		panic(err)
	}
	for _, entity := range entities {
		if ipAddr, err := netip.ParseAddr(entity.IP); err == nil {
			ipSet.Add(ipAddr)
		} else if prefix, err := netip.ParsePrefix(entity.IP); err == nil {
			ipSet.AddPrefix(prefix)
		}
	}
	return &whitelistService{
		gDB:          gDB,
		repo:         repo,
		eventBus:     eventBus,
		ipSetBuilder: &ipSet,
	}
}

func (svc *whitelistService) AddIP(c *gin.Context) error {
	entity := new(model.WhitelistIPModel)
	if err := c.ShouldBindJSON(entity); err != nil {
		return ecode.ErrIPInvalid
	}

	_netIP := utils.ParseIPorCIDR(entity.IP)
	if _netIP.V6 {
		return ecode.ErrIPv6NotSupportedYet
	}

	// ??????IPSet??????IP???CIDR?????????????????????????????????
	if ipSet, err := svc.ipSetBuilder.IPSet(); err != nil {
		logx.Error("[blacklist ip] generate ip set err: ", err)
	} else {
		if _netIP.IP != nil {
			if ipSet.Contains(*_netIP.IP) {
				return ecode.ErrIPAlreadyExisted
			}
		}
		if _netIP.Net != nil {
			if ipSet.ContainsPrefix(*_netIP.Net) {
				return ecode.ErrIPAlreadyExisted
			}
			_prefix := _netIP.Net.Masked()
			entity.IP = _prefix.String()
			_netIP.Net = &_prefix
		}
	}

	//entities := make([]*model.WhitelistIPModel, 0)
	//if err := svc.repo.DB.Where("ip = ?", ip).Find(&entities).Error; err != nil {
	//	logx.Error(fmt.Sprintf("[whitelist] searching whitelist ip %s err: ", ip), err)
	//	return ecode.InternalServerError
	//}
	//if len(entities) > 0 {
	//	return ecode.ErrRecordAlreadyExists
	//}

	entity.ID = 0
	entity.Status = 1
	if err := svc.repo.DB.Create(&entity).Error; err != nil {
		logx.Error("[whitelistIp] insert whitelist url err: ", err)
		return ecode.InternalServerError
	}

	// ??????IPSet
	if _netIP.IP != nil {
		svc.ipSetBuilder.Add(*_netIP.IP)
	} else if _netIP.Net != nil {
		svc.ipSetBuilder.AddPrefix(*_netIP.Net)
	}

	// ??????guard
	configs := make(model.GuardArrayRsp, 0)
	config := map[string]any{
		"id": entity.UUID,
		"ip": entity.IP,
	}
	configs = append(configs, &config)
	svc.eventBus.PushEvent(event.WhitelistIP, event.OpTypeAdd, configs)

	return nil
}

func (svc *whitelistService) AddURL(c *gin.Context, urlModel *model.WhitelistURLModel) error {
	//entities := make([]*model.WhitelistURLModel, 0)
	//if err := svc.repo.DB.Where("site_id = ? and path = ?", siteID, path).Find(&entities).Error; err != nil {
	//	logx.Error(fmt.Sprintf("[whitelist] searching whitelist url whith site_id %d and path %s err:", siteID, path), err)
	//	return ecode.InternalServerError
	//}
	//if len(entities) > 0 {
	//	return ecode.ErrRecordAlreadyExists
	//}

	if err := svc.repo.DB.Where("site_id = ? and path = ? and match_mode = ?", urlModel.SiteID, urlModel.Path, urlModel.MatchMode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrRecordAlreadyExists
		}
		return ecode.InternalServerError
	}

	if err := svc.repo.DB.Create(urlModel).Error; err != nil {
		logx.Error("[whitelistUrl] insert whitelist url err: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{
		siteEntity := model.SiteModel{}
		if err := svc.repo.DB.Where("id = ?", urlModel.SiteID).First(&siteEntity).Error; err == nil {
			configs := make(model.GuardArrayRsp, 0)
			config := map[string]any{
				"id":         urlModel.UUID,
				"path":       urlModel.Path,
				"host":       siteEntity.Domain,
				"site_id":    siteEntity.UUID,
				"match_mode": urlModel.MatchMode,
			}
			configs = append(configs, &config)
			svc.eventBus.PushEvent(event.WhitelistURL, event.OpTypeAdd, configs)
		}
	}
	return nil
}

func (svc *whitelistService) ListIP(pageNum, pageSize, startTime, endTime int64, ip string, status int16) (*model.ListResp, error) {
	rsp := new(model.ListResp)
	entities := make([]*model.WhitelistIPModel, 0)
	var count int64
	gDB := svc.repo.DB.Model(&model.WhitelistIPModel{})
	if startTime > 0 && endTime > 0 {
		gDB = gDB.Where("created_at >= ? AND created_at <= ?", startTime, endTime)
	}
	if len(ip) > 0 {
		gDB.Where("ip = ?", ip)
	}
	if status != -1 {
		gDB.Where("status = ?", status)
	}
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)

	if err := gDB.Count(&count).Error; err != nil {
		logx.Error("[whitelist]Failed get ip whitelist record count: ", err.Error())
		return nil, ecode.InternalServerError
	}
	if err := gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).
		Order("created_at DESC").Find(&entities).Error; err != nil {
		logx.Error("[blacklist ip]Failed to obtaining blacklist ip list: ", err.Error())
		return nil, ecode.InternalServerError
	}

	rsp.SetValue(entities)
	rsp.SetMeta(pageSize, pageNum, count)
	return rsp, nil
}

func (svc *whitelistService) ListURL(pageNum, pageSize, siteID int64, status int16, url string) (*model.ListResp, error) {
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)
	rsp := new(model.ListResp)
	entities := make([]*model.WhitelistURLModel, 0)
	var count int64

	gDB := svc.repo.DB.Model(&model.WhitelistURLModel{})
	if siteID > 0 {
		gDB = gDB.Where("site_id = ?", siteID)
	}
	if status != -1 {
		gDB = gDB.Where("status = ?", status)
	}
	if len(url) > 0 {
		gDB = gDB.Where("path LIKE ?", "%"+url+"%")
	}
	if err := gDB.Count(&count).Error; err != nil {
		logx.Error("[whitelist] Failed to count url whitelist: ", err)
		return nil, ecode.InternalServerError
	}
	if err := gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize)).
		Order("created_at DESC").Find(&entities).Error; err != nil {
		logx.Error("[whitelist] Failed to obtaining url whitelist: ", err.Error())
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

func (svc *whitelistService) UpdateIPStatus(id int64) error {
	if err := svc.repo.WhitelistIPRepo.UpdateStatus(id); err != nil {
		return err
	}
	// guard update
	{
		ipEntity := model.WhitelistIPModel{}
		if err := svc.repo.DB.Where("id = ?", id).First(&ipEntity).Error; err != nil {
			logx.Error("[guard_update] get whitelist ip err: ", err)
			return nil
		}
		if ipEntity.Status == 0 {
			configs := make(model.GuardDelArrayRsp, 0)
			configs = append(configs, ipEntity.UUID)
			svc.eventBus.PushEvent(event.WhitelistIP, event.OpTypeDel, configs)
		}
		if ipEntity.Status == 1 {
			configs := make(model.GuardArrayRsp, 0)
			config := map[string]any{
				"id": ipEntity.UUID,
				"ip": ipEntity.IP,
			}
			configs = append(configs, &config)
			svc.eventBus.PushEvent(event.WhitelistIP, event.OpTypeAdd, configs)
		}
	}
	return nil
}

func (svc *whitelistService) UpdateURLStatus(id int64) error {
	if err := svc.repo.WhitelistURLRepo.UpdateStatus(id); err != nil {
		return err
	}
	// guard update
	{
		urlEntity := model.WhitelistURLModel{}
		if err := svc.repo.DB.Where("id = ?", id).First(&urlEntity).Error; err != nil {
			logx.Error("[guard_update] get whitelist url err: ", err)
			return nil
		}
		if urlEntity.Status == 0 {
			configs := make(model.GuardDelArrayRsp, 0)
			configs = append(configs, urlEntity.UUID)
			svc.eventBus.PushEvent(event.WhitelistURL, event.OpTypeDel, configs)
		}
		if urlEntity.Status == 1 {
			siteEntity := model.SiteModel{}
			if err := svc.repo.DB.Where("id = ?", urlEntity.SiteID).
				Select("Domain", "UUID").
				First(&siteEntity).Error; err != nil {
				logx.Error("[guard_update] get site err: ", err)
				return nil
			}

			configs := make(model.GuardArrayRsp, 0)
			config := map[string]any{
				"id":         urlEntity.UUID,
				"path":       urlEntity.Path,
				"host":       siteEntity.Domain,
				"site_id":    siteEntity.UUID,
				"match_mode": urlEntity.MatchMode,
			}
			configs = append(configs, &config)
			svc.eventBus.PushEvent(event.WhitelistURL, event.OpTypeAdd, configs)
		}
	}
	return nil

}

func (svc *whitelistService) DeleteIP(id int64) error {
	entity := model.WhitelistIPModel{}
	if err := svc.repo.DB.Where("id = ?", id).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrIDNotFound
		} else {
			return ecode.InternalServerError
		}
	}

	if err := svc.repo.DB.Where("id = ?", id).Delete(&model.WhitelistIPModel{}).Error; err != nil {
		logx.Error("[whitelist] delete whitelist ip err: ", err)
		return ecode.InternalServerError
	}

	// ??????IPSet
	_netIP := utils.ParseIPorCIDR(entity.IP)
	if _netIP.IP != nil {
		svc.ipSetBuilder.Remove(*_netIP.IP)
	} else if _netIP.Net != nil {
		svc.ipSetBuilder.RemovePrefix(*_netIP.Net)
	}

	// ??????guard
	{
		configs := make(model.GuardDelArrayRsp, 0)
		configs = append(configs, entity.UUID)
		svc.eventBus.PushEvent(event.WhitelistIP, event.OpTypeDel, configs)
	}

	return nil
}

func (svc *whitelistService) DeleteURL(id int64) error {
	entity := model.WhitelistURLModel{}
	if err := svc.repo.DB.Where("id = ?", id).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrIDNotFound
		} else {
			return ecode.InternalServerError
		}
	}

	if err := svc.repo.DB.Where("id = ?", id).Delete(&model.WhitelistURLModel{}).Error; err != nil {
		logx.Error("[whitelist] delete whitelist url err: ", err)
		return ecode.InternalServerError
	}

	// update guard
	{
		configs := make(model.GuardDelArrayRsp, 0)
		configs = append(configs, entity.UUID)
		svc.eventBus.PushEvent(event.WhitelistURL, event.OpTypeDel, configs)
	}
	return nil
}

func (svc *whitelistService) UpdateURL(id int64, m *model.WhitelistURLModel) error {

	if err := svc.repo.DB.Where("id = ?", id).
		Select("SiteID", "Path", "MatchMode", "Remark").
		Updates(m).Error; err != nil {
		return err
	}

	// update guard
	entity := model.WhitelistURLModel{}
	_ = svc.repo.DB.Where("id = ?", id).First(&entity).Error
	{
		siteEntity := model.SiteModel{}
		if err := svc.repo.DB.Where("id = ?", m.SiteID).First(&siteEntity).Error; err == nil {
			configs := make(model.GuardArrayRsp, 0)
			config := map[string]any{
				"id":         entity.UUID,
				"path":       m.Path,
				"host":       siteEntity.Domain,
				"site_id":    siteEntity.UUID,
				"match_mode": m.MatchMode,
			}
			configs = append(configs, &config)
			svc.eventBus.PushEvent(event.WhitelistURL, event.OpTypeUpdate, configs)
		}
	}

	return nil
}

//func (svc *whitelistService) UpdateIP(id int64, ip, remark string) error {
//	field := make(map[string]any)
//	field["ip"] = ip
//	field["remark"] = remark
//	return svc.repo.WhitelistIPRepo.Update(id, field)
//}

func (svc *whitelistService) IsIpIncluded(ip string) (bool, error) {
	ip = strings.TrimSpace(ip)
	if len(ip) == 0 {
		return false, ecode.ErrIPorCIDREmpty
	}

	netIP := utils.ParseIPorCIDR(ip)

	if !netIP.V4 && !netIP.V6 {
		return false, ecode.ErrIPInvalid
	}

	if ipSet, err := svc.ipSetBuilder.IPSet(); err != nil {
		logx.Error("[whitelist] Failed to generate ip set: ", err)
		return false, ecode.InternalServerError
	} else {
		if netIP.IP != nil {
			if ipSet.Contains(*netIP.IP) {
				return true, nil
			}
		}
		if netIP.Net != nil {
			if ipSet.ContainsPrefix(*netIP.Net) {
				return true, nil
			}
		}
		return false, nil
	}
}

func (svc *whitelistService) UpdateIP(c *gin.Context, id int64, payload *model.WhitelistIPModel) error {
	if _, err := svc.isExist(id); err != nil {
		return err
	}
	if err := svc.repo.DB.Where("id = ?", id).Select("remark").Updates(payload).Error; err != nil {
		logx.Error("[whitelist]Failed to update: ", err.Error())
		return ecode.InternalServerError
	}
	return nil
}

func (svc *whitelistService) isExist(id int64) (bool, error) {
	if err := svc.repo.DB.Where("id = ?", id).First(&model.WhitelistIPModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ecode.ErrIDNotFound
		}
		return false, err
	}
	return true, nil
}
