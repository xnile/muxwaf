package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/handler"
	"github.com/xnile/muxwaf/internal/ecode"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/internal/service"
	"github.com/xnile/muxwaf/pkg/logx"
	"strconv"
)

func Login(c *gin.Context) {
	payload := new(model.UserLoginReq)
	if err := c.ShouldBindJSON(&payload); err != nil {
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	svc := service.SVC.User
	data, err := svc.Login(payload.Username, payload.Password)
	if err != nil {
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, nil, data)
}

func Logout(c *gin.Context) {
	handler.ResponseBuilder(c, nil, nil)
}

func UserInfo(c *gin.Context) {
	//uid := c.GetString("uid")
	uid := c.GetInt64("uid")

	//uid, _ := strconv.ParseInt(c.GetString("uid"), 10, 64)
	// TODO uid为空
	svc := service.SVC.User
	user, err := svc.Info(uid)
	if err != nil {
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, nil, user)
}

func InsertUser(c *gin.Context) {
	var payload model.UserModel

	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	svc := service.SVC.User
	err := svc.Insert(payload.Username, payload.Password, payload.Name, payload.Email, payload.Phone, payload.Avatar)
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, nil, nil)
}

func ListUsers(c *gin.Context) {
	pageNum, _ := strconv.ParseInt(c.Query("page_num"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)

	svc := service.SVC.User
	data, err := svc.List(pageNum, pageSize)
	if err != nil {
		logx.Warn(err)
		handler.ResponseBuilder(c, err, nil)
		return
	}
	handler.ResponseBuilder(c, nil, data)
}

func UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id < 1 {
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}

	var payload model.UserUpdateReq

	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	err := service.SVC.User.Update(id, &payload)
	handler.ResponseBuilder(c, err, nil)
}

func ResetPassword(c *gin.Context) {
	var payload model.UserPasswordResetReq
	if err := c.ShouldBindJSON(&payload); err != nil {
		logx.Warnf("request parameter error: %v", err)
		handler.ResponseBuilder(c, ecode.ErrParam, nil)
		return
	}
	err := service.SVC.User.ResetPassword(c, &payload)
	handler.ResponseBuilder(c, err, nil)
}
