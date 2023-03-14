package model

// UserRoleModel 用户角色表
type UserRoleModel struct {
	Model
	UID    int64 `json:"uid" gorm:"index;not null" binding:"required"`
	RoleID int64 `json:"role_id" gorm:"index;not null" binding:"required"`
}

func (UserRoleModel) TableName() string {
	return "user_role"
}
