package repository

import (
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type ISiteRepo interface {
	Insert(model *model.SiteModel) (int64, error)
	List(pageNum, pageSize int64) ([]*model.SiteModel, int64, error)
	UpdateStatus(id int64) error
	Delete(id int64) error
	GetDomain(id int64) (string, error)
	GetAll() ([]*model.SiteModel, error)
}

type siteRepo struct {
	db *gorm.DB
}

func NewSiteRepo(db *gorm.DB) ISiteRepo {
	return &siteRepo{
		db: db,
	}
}

func (repo *siteRepo) Insert(m *model.SiteModel) (int64, error) {
	err := repo.db.Create(m).Error
	return m.ID, err
}

func (repo *siteRepo) List(pageNum, pageSize int64) (m []*model.SiteModel, count int64, err error) {
	gDB := repo.db.Model(&model.SiteModel{})
	if err = gDB.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err = gDB.Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).
		Order("created_at DESC").Find(&m).Error
	return
}

func (repo *siteRepo) UpdateStatus(id int64) error {
	err := repo.db.Model(&model.SiteModel{}).
		Where("id = ?", id).
		UpdateColumn("status", gorm.Expr("ABS(status - ?)", 1)).
		Error
	return err
}

func (repo *siteRepo) Delete(id int64) error {
	return repo.db.Where("id = ?", id).Delete(&model.SiteModel{}).Error
}

func (repo *siteRepo) GetDomain(id int64) (string, error) {
	entity := new(model.SiteModel)
	var err error
	if err = repo.db.Where("id = ?", id).First(entity).Error; err != nil {
		return "", err
	}
	return entity.Domain, err
}

func (repo *siteRepo) GetAll() (entities []*model.SiteModel, err error) {
	err = repo.db.Model(&model.SiteModel{}).Find(&entities).Error
	return
}
