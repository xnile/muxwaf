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

type IBlacklistIPService interface {
	Add(c *gin.Context) error
	List(c *gin.Context, pageNum, pageSize, start_time, end_time int64, ip string, status int16) (*model.ListResp, error)
	Update(c *gin.Context, id int64, payload *model.BlacklistIPModel) error
	UpdateStatus(id int64) error
	Delete(id int64) error
	IsIncluded(ip string) (bool, error)
	BatchAdd(c *gin.Context, payload *model.BlacklistBatchAddReq) error
}

type blacklistIPService struct {
	gDB          *gorm.DB
	repo         *repository.Repository
	eventBus     *event.EventBus
	ipSetBuilder *netipx.IPSetBuilder
}

func NewBlacklistIPService(gDB *gorm.DB, repo *repository.Repository, eventBus *event.EventBus) IBlacklistIPService {
	var ipSet netipx.IPSetBuilder
	entities := make([]*model.BlacklistIPModel, 0)
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

	return &blacklistIPService{
		gDB:          gDB,
		repo:         repo,
		eventBus:     eventBus,
		ipSetBuilder: &ipSet,
	}
}

func (svc *blacklistIPService) Add(c *gin.Context) error {
	entity := new(model.BlacklistIPModel)
	if err := c.ShouldBindJSON(entity); err != nil {
		return ecode.ErrIPInvalid
	}

	// 暂不支持IPv6
	_netIP := utils.ParseIPorCIDR(entity.IP)
	if _netIP.V6 {
		return ecode.ErrIPv6NotSupportedYet
	}

	// 通过IPSet检测IP或CIDR是否已经包含在数据库中
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

			entity.IP = _prefix.String() // 10.0.0.1/8 -> 10.1.0.0/8
			_netIP.Net = &_prefix        // 处理后的新Net
		}
	}

	//entities := make([]*model.BlacklistIPModel, 0)
	//if err := svc.repo.DB.Where("ip = ?", entity.IP).Find(&entities).Error; err != nil {
	//	logx.Error(fmt.Sprintf("[blacklist ip] searching %s err: ", entity.IP), err)
	//	return ecode.InternalServerError
	//}
	//if len(entities) > 0 {
	//	return ecode.ErrRecordAlreadyExists
	//}

	entity.Status = 1
	entity.ID = 0
	err := svc.repo.DB.Create(&entity).Error
	if err != nil {
		logx.Error("[whitelistIp] insert whitelist ip err: ", err)
		return ecode.InternalServerError
	}

	// 更新IPSet
	if _netIP.IP != nil {
		svc.ipSetBuilder.Add(*_netIP.IP)
	} else if _netIP.Net != nil {
		svc.ipSetBuilder.AddPrefix(*_netIP.Net)
	}

	// update guard
	configs := make(model.GuardArrayRsp, 0)
	config := map[string]any{
		"id": entity.UUID,
		"ip": entity.IP,
	}
	configs = append(configs, &config)
	svc.eventBus.PushEvent(event.BlacklistIP, event.OpTypeAdd, configs)

	return nil
}

func (svc *blacklistIPService) List(c *gin.Context, pageNum, pageSize, start_time, end_time int64, ip string, status int16) (*model.ListResp, error) {
	rsp := new(model.ListResp)
	entities := make([]*model.BlacklistIPModel, 0)
	var count int64
	gDB := svc.repo.DB.Model(&model.BlacklistIPModel{})
	if start_time > 0 && end_time > 0 {
		gDB = gDB.Where("created_at >= ? AND created_at <= ?", start_time, end_time)
	}
	if len(ip) > 0 {
		gDB.Where("ip = ?", ip)
	}
	if status != -1 {
		gDB.Where("status = ?", status)
	}
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)

	if err := gDB.Count(&count).Error; err != nil {
		logx.Error("[blacklist ip]Failed get record count: ", err.Error())
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

func (svc *blacklistIPService) UpdateStatus(id int64) error {

	if err := svc.repo.BlacklistIP.UpdateStatus(id); err != nil {
		return err
	}

	// update guard
	{
		entity := new(model.BlacklistIPModel)
		if err := svc.gDB.Where("id = ?", id).First(entity).Error; err != nil {
			logx.Error("get blacklist ip err: ", err)
			return nil
		}

		if entity.Status == 0 {
			configs := make(model.GuardDelArrayRsp, 0)
			configs = append(configs, entity.UUID)
			svc.eventBus.PushEvent(event.BlacklistIP, event.OpTypeDel, configs)
		}
		if entity.Status == 1 {
			configs := make(model.GuardArrayRsp, 0)
			config := map[string]any{
				"id": entity.UUID,
				"ip": entity.IP,
			}
			configs = append(configs, &config)
			svc.eventBus.PushEvent(event.BlacklistIP, event.OpTypeAdd, configs)
		}
	}

	return nil
}

func (svc *blacklistIPService) Delete(id int64) error {
	entity := model.BlacklistIPModel{}
	if err := svc.repo.DB.Where("id = ?", id).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ecode.ErrIDNotFound
		} else {
			logx.Error("[blacklistIp] obtaining blacklist ip err: ", err)
			return ecode.InternalServerError
		}
	}

	if err := svc.repo.DB.Where("id = ?", id).Delete(&model.BlacklistIPModel{}).Error; err != nil {
		logx.Error("[blacklistIp] delete blacklist ip err: ", err)
		return ecode.InternalServerError
	}

	// 更新IPSet
	_netIP := utils.ParseIPorCIDR(entity.IP)
	if _netIP.IP != nil {
		svc.ipSetBuilder.Remove(*_netIP.IP)
	} else if _netIP.Net != nil {
		svc.ipSetBuilder.RemovePrefix(*_netIP.Net)
	}

	// 更新guard
	{
		configs := make(model.GuardDelArrayRsp, 0)
		configs = append(configs, entity.UUID)
		svc.eventBus.PushEvent(event.BlacklistIP, event.OpTypeDel, configs)
	}

	return nil
}

func (svc *blacklistIPService) IsIncluded(ip string) (bool, error) {
	ip = strings.TrimSpace(ip)
	if len(ip) == 0 {
		return false, ecode.ErrIPorCIDREmpty
	}

	netIP := utils.ParseIPorCIDR(ip)

	if !netIP.V4 && !netIP.V6 {
		return false, ecode.ErrIPInvalid
	}

	if ipSet, err := svc.ipSetBuilder.IPSet(); err != nil {
		logx.Error("[blacklist ip] generate ip set err: ", err)
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

func (svc *blacklistIPService) Update(c *gin.Context, id int64, payload *model.BlacklistIPModel) error {
	if _, err := svc.isExist(id); err != nil {
		return err
	}
	if err := svc.repo.DB.Where("id = ?", id).Select("remark").Updates(payload).Error; err != nil {
		logx.Error("[blacklist]Failed to update: ", err.Error())
		return ecode.InternalServerError
	}
	return nil
}

func (svc *blacklistIPService) isExist(id int64) (bool, error) {
	if err := svc.repo.DB.Where("id = ?", id).First(&model.BlacklistIPModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ecode.ErrIDNotFound
		}
		return false, err
	}
	return true, nil
}

func ipValidator(ipSet *netipx.IPSet, ip string) (*utils.IPNet, error) {
	ipNet := utils.ParseIPorCIDR(ip)
	if !ipNet.V4 && !ipNet.V6 {
		return nil, ecode.ErrIPInvalid
	}

	if ipNet.V6 {
		return nil, ecode.ErrIPv6NotSupportedYet
	}
	if ipNet.IP != nil {
		if ipSet.Contains(*ipNet.IP) {
			return nil, ecode.ErrIPAlreadyExisted
		}
	}
	if ipNet.Net != nil {
		if ipSet.ContainsPrefix(*ipNet.Net) {
			return nil, ecode.ErrIPAlreadyExisted
		}
		_prefix := ipNet.Net.Masked()
		ipNet.Net = &_prefix
	}
	return &ipNet, nil
}

func (svc *blacklistIPService) BatchAdd(c *gin.Context, payload *model.BlacklistBatchAddReq) error {
	var ipSet *netipx.IPSet
	var err error

	ipSet, err = svc.ipSetBuilder.IPSet()
	if err != nil {
		logx.Error("[blacklist ip] failed to create ip set err: ", err)
		return ecode.InternalServerError
	}

	ipNetList := make([]*utils.IPNet, 0)
	for _, ip := range payload.IPList {
		ipNet, err := ipValidator(ipSet, ip)
		if err != nil {
			return err
		}
		ipNetList = append(ipNetList, ipNet)
	}

	tx := svc.repo.DB.Begin()
	for _, ipNet := range ipNetList {
		var ip string
		if ipNet.IP != nil {
			ip = ipNet.IP.String()
		} else if ipNet.Net != nil {
			ip = ipNet.Net.Masked().String()
		}

		if err := tx.Create(&model.BlacklistIPModel{
			IP:     ip,
			Status: 1,
			Remark: payload.Remark,
		}).Error; err != nil {
			logx.Error("[blacklist ip] failed to batch add: ", err)
			tx.Rollback()
			return ecode.InternalServerError
		}

		//// 更新IPSet
		//if ipNet.IP != nil {
		//	svc.ipSetBuilder.Add(*ipNet.IP)
		//} else if ipNet.Net != nil {
		//	svc.ipSetBuilder.AddPrefix(*ipNet.Net)
		//}
	}
	if err := tx.Commit().Error; err != nil {
		logx.Error("[blacklist ip] failed to commit: ", err)
	}
	return nil
}
