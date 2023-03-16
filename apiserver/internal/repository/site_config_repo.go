package repository

import (
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type ISiteConfigRepo interface {
	Insert(configModel *model.SiteConfigModel) error
	GetBySiteID(siteID int64) (*model.SiteConfigModel, error)
	Update(id int64, field map[string]interface{}) error
	Delete(id int64) error
}

type siteConfigRepo struct {
	db *gorm.DB
}

func NewSiteConfigRepo(db *gorm.DB) ISiteConfigRepo {
	return &siteConfigRepo{
		db: db,
	}
}

func (repo *siteConfigRepo) Insert(configModel *model.SiteConfigModel) error {
	return repo.db.Create(configModel).Error

}

func (repo *siteConfigRepo) GetBySiteID(siteID int64) (*model.SiteConfigModel, error) {
	entity := new(model.SiteConfigModel)
	err := repo.db.Where("site_id = ?", siteID).
		First(entity).Error
	return entity, err
}

func (repo *siteConfigRepo) Update(id int64, field map[string]interface{}) error {
	return repo.db.Model(&model.SiteConfigModel{}).Where("id = ?", id).Updates(field).Error
}

func (repo *siteConfigRepo) Delete(id int64) error {
	return repo.db.Where("id = ?", id).Delete(&model.SiteConfigModel{}).Error
}
