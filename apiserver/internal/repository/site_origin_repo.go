package repository

import (
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type ISiteOriginRepo interface {
	Insert(model *model.SiteOriginModel) error
	GetBySiteID(siteID int64) ([]*model.SiteOriginModel, error)
	Update(id int64, field map[string]interface{}) error
	Delete(id int64) error
}

type siteOriginRepo struct {
	db *gorm.DB
}

func NewSiteOriginRepo(db *gorm.DB) ISiteOriginRepo {
	return &siteOriginRepo{
		db: db,
	}
}

func (repo *siteOriginRepo) Insert(originModel *model.SiteOriginModel) error {
	return repo.db.Create(originModel).Error
}

func (repo *siteOriginRepo) GetBySiteID(siteID int64) ([]*model.SiteOriginModel, error) {
	entities := make([]*model.SiteOriginModel, 0)
	err := repo.db.Where("site_id = ?", siteID).Find(&entities).Error
	return entities, err
}

func (repo *siteOriginRepo) Update(id int64, field map[string]interface{}) error {
	return repo.db.Model(&model.SiteOriginModel{}).Where("id = ?", id).Updates(field).Error
}

func (repo *siteOriginRepo) Delete(id int64) error {
	return repo.db.Where("id = ?", id).Delete(&model.SiteOriginModel{}).Error
}
