package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/repository"
	"github.com/xnile/muxwaf/pkg/auth"
	"github.com/xnile/muxwaf/pkg/logx"
	"github.com/xnile/muxwaf/pkg/token"
	"github.com/xnile/muxwaf/pkg/utils"
	"gorm.io/gorm"
	"time"
)

type UserService interface {
	Insert(username, password, name, email, phone, avatar string) error
	List(pageNum, pageSize int64) (*model.ListResp, error)
	Update(uid int64, toUpdate *model.UserUpdateReq) error
	Delete(uid int64) error
	Block(uid int64) error
	Login(username, password string) (map[string]interface{}, error)
	GetByID(uid int64) (*model.UserModel, error)
	Info(uid int64) (*model.UserModel, error)
	ResetPassword(c *gin.Context, payload *model.UserPasswordResetReq) error
}

type userService struct {
	gDB  *gorm.DB
	repo *repository.Repository
}

func NewUserService(gDB *gorm.DB, repo *repository.Repository) UserService {
	if err := gDB.Select("ID").Where("username = ?", "admin").First(&model.UserModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := gDB.Create(&model.UserModel{
				Username: "admin",
				Password: auth.HashPassword(utils.MD5("admin@123")),
				Name:     "超级管理员",
				Email:    "admin@muxwaf.com",
				Phone:    "13800000000",
			}); err != nil {
				logx.Error("[site]Failed to init admin user: ", err)
			}
		}
	}
	return &userService{
		gDB:  gDB,
		repo: repo,
	}
}

func (svc *userService) Insert(username, password, name, email, phone, avatar string) error {
	password = auth.HashPassword(utils.MD5(password))
	err := svc.repo.User.Insert(&model.UserModel{
		Username:     username,
		Password:     password,
		Name:         name,
		Email:        email,
		Phone:        phone,
		Avatar:       avatar,
		UpdatedAt:    0,
		BlockedAt:    0,
		LastSignInIP: "",
		LastSignInAt: 0,
	})
	return err
}

func (svc *userService) List(pageNum, pageSize int64) (*model.ListResp, error) {
	rsp := &model.ListResp{}
	pageNum, pageSize = utils.CheckPageSizeNum(pageNum, pageSize)
	entities, count, err := svc.repo.User.List(pageNum, pageSize)
	if err != nil {
		return rsp, err
	}

	if count == 0 {
		return rsp, nil
	}

	v := make([]map[string]interface{}, 0)
	for _, entity := range entities {
		dto := make(map[string]interface{})
		dto["id"] = entity.ID
		dto["username"] = entity.Username
		dto["name"] = entity.Name
		dto["email"] = entity.Email
		dto["phone"] = entity.Phone
		dto["created_at"] = entity.CreatedAt
		dto["blocked_at"] = entity.BlockedAt
		dto["last_sign_in_at"] = entity.LastSignInAt

		v = append(v, dto)
	}

	//meta := model.Meta{
	//	PageSize: pageSize,
	//	PageNum:  pageNum,
	//	Pages:    utils.CalPage(count, pageSize),
	//	Total:    count,
	//}
	rsp.SetMeta(pageSize, pageNum, count)
	rsp.SetValue(v)

	return rsp, nil
}

func (svc *userService) Update(uid int64, toUpdate *model.UserUpdateReq) error {
	user := new(model.UserModel)
	if err := copier.Copy(user, toUpdate); err != nil {
		logx.Error("[user] Failed to copy user: ", err)
		return ecode.InternalServerError
	}

	user.ID = uid
	gDB := svc.repo.DB.Where("id = ?", uid)
	if len(user.Password) > 0 {
		user.Password = auth.HashPassword(utils.MD5(user.Password))
		gDB.Select("Password", "Name", "Email", "Phone")
	} else {
		gDB.Select("Name", "Email", "Phone")
	}
	if err := gDB.Updates(user).Error; err != nil {
		logx.Error("[user] Failed to update user: ", err.Error())
		return ecode.InternalServerError
	}

	return nil
}

func (svc *userService) Delete(uid int64) error {
	//
	//if err := svc.repo.UserRole.DeleteByUID(uid); err != nil {
	//	return err
	//}
	err := svc.repo.User.Delete(uid)
	return err
}

func (svc *userService) Block(uid int64) error {
	field := make(map[string]interface{})
	user, err := svc.repo.User.GetUserByID(uid)
	if err != nil {
		return err
	}
	if user.BlockedAt == 0 {
		field["blocked_at"] = time.Now().Unix()
	} else {
		field["blocked_at"] = 0
	}

	return svc.repo.User.Update(uid, field)
}

func (svc *userService) Login(username, password string) (map[string]interface{}, error) {
	userEntity := new(model.UserModel)
	if err := svc.gDB.Where("username = ?", username).First(userEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUsernameOrPwdIncorrect
		}
		return nil, ecode.InternalServerError
	}

	if err := auth.CheckPassword(password, userEntity.Password); err != nil {
		return nil, ecode.ErrUsernameOrPwdIncorrect
	}

	userEntity.LastSignInAt = time.Now().Unix()
	if err := svc.repo.DB.Where("id = ?", userEntity.ID).Select("last_sign_in_at").Updates(userEntity).Error; err != nil {
		logx.Warnf("Failed to update last_sign_in_at: ", err)
	}

	tokenStr := token.Encode(userEntity.ID)

	rsp := make(map[string]interface{})
	rsp["token"] = tokenStr
	//if err := svc.cache.User.AddUserToken(userEntity.ID, tokenStr); err != nil {
	//
	//}
	return rsp, nil
}

func (svc *userService) GetByID(uid int64) (*model.UserModel, error) {
	user, err := svc.repo.User.GetUserByID(uid)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (svc *userService) Info(uid int64) (*model.UserModel, error) {
	user, err := svc.repo.User.GetUserByID(uid)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	user.Role = make([]string, 0)
	// 默认角色
	user.Role = append(user.Role, "登陆用户")
	if user.Username == "admin" {
		user.Role = append(user.Role, "超级管理员")
	}

	return user, nil
}

func (svc *userService) ResetPassword(c *gin.Context, payload *model.UserPasswordResetReq) error {
	uid := c.GetInt64("uid")
	if uid < 1 {
		return errors.New("请先登录")
	}
	user := new(model.UserModel)
	if err := svc.repo.DB.Where("id = ?", uid).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		logx.Error("[user]Failed to get user: ", err)
		return ecode.InternalServerError
	}

	if err := auth.CheckPassword(payload.OldPassword, user.Password); err != nil {
		logx.Error("ERR: ", err.Error())
		return errors.New("当前密码不正确")
	}

	newPass := auth.HashPassword(payload.NewPassword)
	if err := svc.repo.DB.Where("id = ?", uid).
		Select("Password").
		Updates(&model.UserModel{Password: newPass}).Error; err != nil {
		logx.Error("[user] Failed to reset password: ", err.Error())
		return ecode.InternalServerError
	}

	return nil
}
