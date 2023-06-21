package model

type SiteModel struct {
	Model
	Domain    string `json:"domain" gorm:"uniqueIndex;type:varchar(255);not null" binding:"required,fqdn"`
	Status    int16  `json:"status" gorm:"type:smallint;not null;default:0"`
	Remark    string `json:"remark" gorm:"type:text"`
	DeletedAt int64  `json:"-" gorm:"not null;default:0"`
}

func (SiteModel) TableName() string {
	return "site"
}

type SiteReq struct {
	Domain         string             `json:"domain" binding:"required,fqdn"`
	OriginProtocol OriginProtocol     `json:"origin_protocol" binding:"required,oneof=http https"`
	Origins        []*SiteOriginModel `json:"origins"  binding:"gt=0,dive"`
}

type SiteRsp struct {
	Model
	Domain  string           `json:"domain"`
	Status  int16            `json:"status"`
	Remark  string           `json:"remark"`
	Config  *SiteConfigRsp   `json:"config"`
	Origins []*SiteOriginRsp `json:"origins"`
}

// SiteGuardRsp guard sync entity
//type SiteGuardRsp struct {
//	ID      string             `json:"id"`
//	Host    string             `json:"host"`
//	Config  *SiteConfigGuard   `json:"config"`
//	Origins []*SiteOriginGuard `json:"origins"`
//}
