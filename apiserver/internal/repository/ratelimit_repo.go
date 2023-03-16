package repository

import (
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type IRateLimitRepo interface {
	Insert(model *model.RateLimitModel) error
	List(pageNum, pageSize int64) (model []*model.RateLimitModel, count int64, err error)
	UpdateStatus(id int64) error
	Delete(id int64) error
	Update(id int64, field map[string]any) error
}

type rateLimitRepo struct {
	db *gorm.DB
}

func NewRateLimitRepo(db *gorm.DB) IRateLimitRepo {
	return &rateLimitRepo{
		db: db,
	}
}

func (repo *rateLimitRepo) Insert(model *model.RateLimitModel) error {
	return repo.db.Create(model).Error
}

func (repo *rateLimitRepo) List(pageNum, pageSize int64) (m []*model.RateLimitModel, count int64, err error) {
	gDB := repo.db.Model(&model.RateLimitModel{})
	if err = gDB.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err = gDB.Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).Order("created_at DESC").
		Find(&m).Error
	return
}

func (repo *rateLimitRepo) UpdateStatus(id int64) error {
	return repo.db.Model(&model.RateLimitModel{}).
		Where("id = ?", id).
		UpdateColumn("status", gorm.Expr("ABS(status - ?)", 1)).
		Error
}

func (repo *rateLimitRepo) Delete(id int64) error {
	return repo.db.Where("id = ?", id).
		Delete(&model.RateLimitModel{}).Error
}

func (repo *rateLimitRepo) Update(id int64, field map[string]any) error {
	return repo.db.Model(&model.RateLimitModel{}).Where("id = ?", id).Updates(field).Error
}
