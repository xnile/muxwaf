package model

type WhitelistIPModel struct {
	Model
	IP     string `json:"ip" gorm:"uniqueIndex;type:varchar(39)" binding:"required,ip4_addr|cidrv4"`
	Status int16  `json:"status" gorm:"type:smallint;not null;default 1"`
	Remark string `json:"remark" gorm:"type:text"`
}

func (WhitelistIPModel) TableName() string {
	return "whitelist_ip"
}
