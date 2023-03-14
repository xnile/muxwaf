package repository

import (
	"errors"
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type IWhitelistIPRepo interface {
	Insert(model *model.WhitelistIPModel) (int64, error)
	List(pageNum, pageSize int64) ([]*model.WhitelistIPModel, int64, error)
	IsExist(ip string) (bool, error)
	UpdateStatus(id int64) error
	Delete(id int64) error
	Update(id int64, field map[string]any) error
}

type whitelistIPRepo struct {
	db *gorm.DB
}

func NewWhitelistIPRepo(db *gorm.DB) IWhitelistIPRepo {
	return &whitelistIPRepo{db: db}
}

func (repo *whitelistIPRepo) Insert(m *model.WhitelistIPModel) (int64, error) {
	err := repo.db.Create(m).Error
	return m.ID, err
}

func (repo *whitelistIPRepo) List(pageNum, pageSize int64) (m []*model.WhitelistIPModel, count int64, err error) {
	gDB := repo.db.Model(&model.WhitelistIPModel{})
	if err = gDB.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	gDB = gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize))
	err = gDB.Order("created_at DESC").Find(&m).Error
	return
}

func (repo *whitelistIPRepo) IsExist(ip string) (bool, error) {
	if err := repo.db.Where("ip = ?", ip).First(&model.WhitelistIPModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (repo *whitelistIPRepo) UpdateStatus(id int64) error {
	err := repo.db.Model(&model.WhitelistIPModel{}).
		Where("id = ?", id).UpdateColumn("status", gorm.Expr("ABS(status - ?)", 1)).Error
	return err
}

func (repo *whitelistIPRepo) Delete(id int64) error {
	return repo.db.Where("id = ?", id).Delete(&model.WhitelistIPModel{}).Error
}

func (repo *whitelistIPRepo) Update(id int64, field map[string]any) error {
	return repo.db.Model(&model.WhitelistIPModel{}).Where("id = ?", id).Updates(field).Error
}
