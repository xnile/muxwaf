package repository

import (
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type ICertRepo interface {
	Insert(model *model.CertModel) error
	List(pageNum, pageSize int64) (model []*model.CertModel, count int64, err error)
	UpdateStatus(id int64) error
	Delete(id int64) error
	Get(id int64) (*model.CertModel, error)
	All() ([]*model.CertModel, error)
}

type certRepo struct {
	db *gorm.DB
}

func NewCertRepo(db *gorm.DB) ICertRepo {
	return &certRepo{
		db: db,
	}
}

func (repo *certRepo) Insert(model *model.CertModel) error {
	return repo.db.Create(model).Error
}

func (repo *certRepo) List(pageNum, pageSize int64) (m []*model.CertModel, count int64, err error) {
	gDB := repo.db.Model(&model.CertModel{})
	if err = gDB.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err = gDB.Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).Order("created_at DESC").
		Find(&m).Error
	return
}

func (repo *certRepo) UpdateStatus(id int64) error {
	return repo.db.Model(&model.CertModel{}).
		Where("id = ?", id).
		UpdateColumn("status", gorm.Expr("ABS(status - ?)", 1)).
		Error
}

func (repo *certRepo) Delete(id int64) error {
	return repo.db.Where("id = ?", id).
		Delete(&model.CertModel{}).Error
}

func (repo *certRepo) Get(id int64) (*model.CertModel, error) {
	entity := new(model.CertModel)
	if err := repo.db.Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (repo *certRepo) All() (m []*model.CertModel, err error) {
	err = repo.db.Find(&m).Error
	return
}
