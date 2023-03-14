package model

// RoleModel 用户角色
type RoleModel struct {
	Model
	Name   string `json:"name" gorm:"uniqueIndex;type:varchar" binding:"required"`
	Remark string `json:"remark" gorm:"type:text"`
}

func (RoleModel) TableName() string {
	return "role"
}
