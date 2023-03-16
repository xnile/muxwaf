package model

type BlacklistIPModel struct {
	Model
	IP     string `json:"ip" gorm:"uniqueIndex;type varchar(39);not null" binding:"required,ip|cidr"`
	Status int16  `json:"status" gorm:"type smallint;not null;default:1"`
	Remark string `json:"remark" gorm:"type text"`
}

func (BlacklistIPModel) TableName() string {
	return "blacklist_ip"
}
