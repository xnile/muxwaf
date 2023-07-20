package repository

import (
	"errors"
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type IBlacklistIPRepo interface {
	Insert(model *model.BlacklistIPModel) (int64, error)
	List(pageNum, pageSize int64) ([]*model.BlacklistIPModel, int64, error)
	IsExist(ip string) (bool, error)
	UpdateStatus(id int64) error
	Delete(id int64) error
	GetStatus(id int64) (int8, error)
}

type blacklistIPRepo struct {
	db *gorm.DB
}

func NewBlacklistIPRepo(db *gorm.DB) IBlacklistIPRepo {
	return &blacklistIPRepo{
		db: db,
	}
}

func (repo *blacklistIPRepo) Insert(m *model.BlacklistIPModel) (int64, error) {
	err := repo.db.Create(m).Error
	return m.ID, err
}

func (repo *blacklistIPRepo) List(pageNum, pageSize int64) (m []*model.BlacklistIPModel, count int64, err error) {
	gDB := repo.db.Model(&model.BlacklistIPModel{})
	if err = gDB.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	gDB = gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize))
	err = gDB.Order("created_at DESC").Find(&m).Error
	return
}

func (repo *blacklistIPRepo) IsExist(ip string) (bool, error) {
	if err := repo.db.Where("ip = ?", ip).First(&model.BlacklistIPModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (repo *blacklistIPRepo) UpdateStatus(id int64) error {
	err := repo.db.Model(&model.BlacklistIPModel{}).
		Where("id = ?", id).UpdateColumn("status", gorm.Expr("ABS(status - ?)", 1)).Error
	return err
}

func (repo *blacklistIPRepo) Delete(id int64) error {
	return repo.db.Where("id = ?", id).Delete(&model.BlacklistIPModel{}).Error
}

func (repo *blacklistIPRepo) GetStatus(id int64) (int8, error) {
	entity := model.BlacklistIPModel{}
	err := repo.db.Where("id = ?", id).Select("Status").First(&entity).Error
	return entity.Status, err
}
