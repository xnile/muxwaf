package model

type WhitelistURLModel struct {
	Model
	SiteID    int64  `json:"site_id" gorm:"index;type:bigint;not null" binding:"required"`
	SiteUUID  string `json:"-" gorm:"index;type:char(20);default:''"`
	Host      string `json:"host" gorm:"uniqueIndex;type:varchar(255);not null"`
	Path      string `json:"path" gorm:"type:text;not null" binding:"required,uri"`
	MatchMode int8   `json:"match_mode" gorm:"type:smallint;not null;default:1;comment:url match mode,1 as prefix,2 as exact" binding:"required"`
	//Method    int8   `json:"method" gorm:"type:smallint;not null;default:0;comment:not used yet"`
	Status int8   `json:"status" gorm:"type:smallint;not null;default:1"`
	Remark string `json:"remark" gorm:"type:text"`
	Domain string `json:"domain" gorm:"-"`
}

func (WhitelistURLModel) TableName() string {
	return "whitelist_url"
}
