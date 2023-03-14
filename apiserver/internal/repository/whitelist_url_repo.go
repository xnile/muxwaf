package repository

import (
	"errors"
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type IWhitelistURLRepo interface {
	Insert(m *model.WhitelistURLModel) (int64, error)
	List(pageNum, pageSize int64) ([]*model.WhitelistURLModel, int64, error)
	IsExist(ip string) (bool, error)
	UpdateStatus(id int64) error
	Delete(id int64) error
	Update(id int64, field map[string]any) error
}

type whitelistURLRepo struct {
	db *gorm.DB
}

func NewWhitelistURLRepo(db *gorm.DB) IWhitelistURLRepo {
	return &whitelistURLRepo{db: db}
}

func (repo *whitelistURLRepo) Insert(m *model.WhitelistURLModel) (int64, error) {
	err := repo.db.Create(m).Error
	return m.ID, err
}

func (repo *whitelistURLRepo) List(pageNum, pageSize int64) (m []*model.WhitelistURLModel, count int64, err error) {
	gDB := repo.db.Model(&model.WhitelistURLModel{})
	if err = gDB.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	gDB = gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize))
	err = gDB.Order("created_at DESC").Find(&m).Error
	return
}

func (repo *whitelistURLRepo) IsExist(ip string) (bool, error) {
	if err := repo.db.Where("ip = ?", ip).First(&model.WhitelistURLModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (repo *whitelistURLRepo) UpdateStatus(id int64) error {
	err := repo.db.Model(&model.WhitelistURLModel{}).
		Where("id = ?", id).UpdateColumn("status", gorm.Expr("ABS(status - ?)", 1)).Error
	return err
}

func (repo *whitelistURLRepo) Delete(id int64) error {
	return repo.db.Where("id = ?", id).Delete(&model.WhitelistURLModel{}).Error
}

func (repo *whitelistURLRepo) Update(id int64, field map[string]any) error {
	return repo.db.Model(&model.WhitelistURLModel{}).Where("id = ?", id).Updates(field).Error
}
