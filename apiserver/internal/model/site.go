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
	SiteModel
	Origins []*SiteOriginModel `json:"origins" gorm:"-" binding:"required"`
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
type SiteGuardRsp struct {
	ID      string             `json:"id"`
	Host    string             `json:"host"`
	Config  *SiteConfigGuard   `json:"config"`
	Origins []*SiteOriginGuard `json:"origins"`
}

//type SiteConfigGuard struct {
//	IsHttps            int16  `json:"is_https"`
//	IsRealIpFromHeader int16  `json:"is_real_ip_from_header"`
//	RealIpHeader       string `json:"real_ip_header"`
//	OriginProtocol     int16  `json:"origin_protocol"`
//	CertID             string `json:"cert_id"`
//}

//type SiteOriginGuard struct {
//	Host      string `json:"host"`
//	HttpPort  int16  `json:"http_port"`
//	HttpsPort int16  `json:"https_port"`
//	Weight    int16  `json:"weight"`
//}
