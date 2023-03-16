package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/pkg/logx"
	"github.com/xnile/muxwaf/pkg/utils"
	"gorm.io/gorm"
)

type IAttackLogService interface {
	Add(c *gin.Context, entity *model.AttackLogModel)
	List(pageNum, pageSize, startTime, endTime, siteID int64, action int8, content string) (*model.ListResp, error)
}

type attackLogService struct {
	db *gorm.DB
}

func NewAttackLogService(db *gorm.DB) IAttackLogService {
	return &attackLogService{db: db}
}

func (svc *attackLogService) Add(c *gin.Context, entity *model.AttackLogModel) {
	token := c.GetHeader("Token")

	if err := svc.db.Where("sampled_log_upload_api_token = ?", token).First(new(model.NodeModel)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(403, gin.H{"code": 403, "message": "Permission denied"})
			c.Abort()
			return
		}
		logx.Error("[attack log] failed to get node token: ", err)
		c.JSON(500, gin.H{"code": 500, "message": "系统错误"})
		c.Abort()
		return
	}

	if err := svc.db.Create(&entity).Error; err != nil {
		logx.Error("[attack log] add attack log err: ", err)
		c.JSON(500, gin.H{"code": 500, "message": "系统错误"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "OK"})
}

func (svc *attackLogService) List(pageNum, pageSize, startTime, endTime, siteID int64, action int8, content string) (*model.ListResp, error) {
	rsp := new(model.ListResp)
	entities := make([]*model.AttackLogModel, 0)
	var count int64
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)

	gDB := svc.db.Model(&model.AttackLogModel{})
	if startTime > 0 && endTime > 0 {
		gDB = gDB.Where("created_at >= ? AND created_at <= ?", startTime, endTime)
	}
	if action != -1 {
		gDB = gDB.Where("action = ?", action)
	}
	if siteID > 0 {
		var siteUUID string
		if err := svc.db.Table("site").Select("uuid").Where("id = ?", siteID).Scan(&siteUUID).Error; err != nil {
			logx.Error("[sampled_log] Failed to get site uuid: ", err)
		} else {
			gDB = gDB.Where("site_id = ?", siteUUID)
		}
	}
	if len(content) > 0 {
		gDB.Where("request_path LIKE ? or real_client_ip = ? or request_id = ?", content+"%", content, content)
	}

	if err := gDB.Count(&count).Error; err != nil {
		logx.Error("[attack log] get attack log err: ", err)
		return nil, ecode.InternalServerError
	}
	gDB = gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize))
	if err := gDB.Order("created_at DESC").Find(&entities).Error; err != nil {
		logx.Error("[attack log] get attack log err: ", err)
		return nil, ecode.InternalServerError
	}

	rsp.SetValue(entities)
	rsp.SetMeta(pageSize, pageNum, count)
	return rsp, nil
}
