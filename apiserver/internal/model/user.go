package model

type UserModel struct {
	Model
	Username     string   `json:"username" gorm:"uniqueIndex;type:varchar(64);not null" binding:"required"`
	Password     string   `json:"password" gorm:"index;type:varchar" binding:"required,gte=8"`
	Name         string   `json:"name" gorm:"uniqueIndex;type:varchar(64);not null"`
	Email        string   `json:"email" gorm:"type:varchar;not null" binding:"omitempty,email"` // omitempty 如果为空则跳过校验
	Phone        string   `json:"phone" gorm:"uniqueIndex;type:varchar(20)"`
	Avatar       string   `json:"avatar" gorm:"type:varchar"`
	UpdatedAt    int64    `json:"updated_at"`
	BlockedAt    int64    `json:"blocked_at"`
	LastSignInIP string   `json:"last_sign_in_ip" gorm:"type:varchar"`
	LastSignInAt int64    `json:"last_sign_in_at"`
	Role         []string `json:"role" gorm:"-"`
}

func (UserModel) TableName() string {
	return "user"
}

type UserLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserUpdateReq struct {
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone"`
}

type UserPasswordResetReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
