package model

type SiteConfigModel struct {
	Model
	SiteID             int64  `json:"site_id" gorm:"index;not null"`
	HttpPort           int16  `json:"http_port" gorm:"not null;default:80"`
	HttpsPort          int16  `json:"https_port" gorm:"not null;default:443"`
	IsHttps            int16  `json:"is_https" gorm:"not null;default:0"`
	CertID             int64  `json:"cert_id" gorm:"index;not null;default:0"`
	OriginProtocol     int16  `json:"origin_protocol" gorm:"not null;default:1"`
	IsRealIPFromHeader int16  `json:"is_real_ip_from_header" gorm:"not null;default:0"`
	RealIPHeader       string `json:"real_ip_header" gorm:"type:varchar(64)"`
	IsHttpsForce       int8   `json:"is_https_force" gorm:"not null;default:0"`
	OriginHost         string `json:"origin_host" gorm:"type:varchar(255);not null;default:''"`
}

func (SiteConfigModel) TableName() string {
	return "site_config"
}

type SiteConfigRsp struct {
	CertID             int64  `json:"cert_id"`
	IsHttps            int16  `json:"is_https"`
	OriginProtocol     int16  `json:"origin_protocol"`
	IsRealIPFromHeader int16  `json:"is_real_ip_from_header"`
	RealIPHeader       string `json:"real_ip_header"`
}

type SiteConfigReq struct {
	IsRealIPFromHeader *int16  `json:"is_real_ip_from_header" binding:"required"`
	RealIPHeader       *string `json:"real_ip_header" binding:"required"`
	OriginProtocol     *int16  `json:"origin_protocol" binding:"required"`
}

type SiteHttpsReq struct {
	IsHttps *int16 `json:"is_https" binding:"required"`
	CertID  *int64 `json:"cert_id" binding:"required"`
}
