package repository

import (
	"errors"
	"github.com/xnile/muxwaf/internal/model"
	"gorm.io/gorm"
)

type UserRepo interface {
	Insert(m *model.UserModel) error
	List(pageNum, pageSize int64) (entities []model.UserModel, count int64, err error)
	Update(uid int64, field map[string]interface{}) error
	Delete(uid int64) error
	GetUserByUsername(username string) (*model.UserModel, error)
	GetUserByID(uid int64) (*model.UserModel, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{
		db: db,
	}
}

func (repo *userRepo) Insert(m *model.UserModel) error {
	return repo.db.Create(m).Error
}

func (repo *userRepo) List(pageNum, pageSize int64) (entities []model.UserModel, count int64, err error) {
	gDB := repo.db.Model(&model.UserModel{})
	if err = gDB.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if pageNum > 0 && pageSize > 0 {
		gDB = gDB.Offset(int((pageNum - 1) * pageSize)).Limit(int(pageSize))
	}
	//err = gDB.Order("created_at DESC").Select("id", "username", "name", "email", "phone").Find(&entities).Error
	err = gDB.Order("created_at DESC").Find(&entities).Error
	return
}

func (repo *userRepo) Update(uid int64, field map[string]interface{}) error {
	err := repo.db.Model(&model.UserModel{}).Where("id = ?", uid).Updates(field).Error
	return err
}

func (repo *userRepo) Delete(uid int64) error {
	err := repo.db.Where("id = ?", uid).Delete(&model.UserModel{}).Error
	return err
}

func (repo *userRepo) GetUserByUsername(username string) (*model.UserModel, error) {
	et := new(model.UserModel)
	err := repo.db.Where("username = ?", username).First(et).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return et, nil
}

func (repo *userRepo) GetUserByID(uid int64) (*model.UserModel, error) {
	et := new(model.UserModel)
	err := repo.db.Where("id = ?", uid).First(et).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return et, nil
}
