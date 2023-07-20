package model

type RateLimitModel struct {
	Model
	SiteID    int64  `json:"site_id" gorm:"index;not null" binding:"required,numeric"`
	SiteUUID  string `json:"-" gorm:"index;type:char(20);default:''"`
	Host      string `json:"host" gorm:"index;type:varchar(255);not null"`
	Path      string `json:"path" gorm:"type:text;not null" binding:"required,uri"`
	Limit     int64  `json:"limit" gorm:"type:bigint;not null" binding:"required,numeric"`
	Window    int64  `json:"window" gorm:"type:bigint;not null" binding:"required,numeric"`
	MatchMode int8   `json:"match_mode" gorm:"type:smallint;not null;default:1;comment:1 prefix match,2 exact match" binding:"required,oneof=1 2"`
	Status    int8   `json:"status" gorm:"type:smallint;not null;default:1"`
	Remark    string `json:"remark" gorm:"type:text"`
	//Domain    string `json:"domain" gorm:"-"`
}

func (RateLimitModel) TableName() string {
	return "rate_limit"
}

type RateLimitReq struct {
	SiteID    int64  `json:"site_id"  binding:"required,numeric"`
	Path      string `json:"path" binding:"required,uri"`
	Limit     int64  `json:"limit" binding:"required,numeric"`
	Window    int64  `json:"window" binding:"required,numeric"`
	MatchMode int8   `json:"match_mode" binding:"required,oneof=1 2"`
	Remark    string `json:"remark"`
}
