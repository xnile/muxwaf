package model

type SiteConfigModel struct {
	Model
	SiteID             int64          `json:"site_id" gorm:"index;not null"`
	SiteUUID           string         `json:"-" gorm:"index;type:char(20);default:''"`
	HttpPort           int16          `json:"http_port" gorm:"not null;default:80"`
	HttpsPort          int16          `json:"https_port" gorm:"not null;default:443"`
	IsHttps            int8           `json:"is_https" gorm:"not null;default:0"`
	IsForceHttps       int8           `json:"is_force_https" gorm:"not null;default:0"`
	CertID             int64          `json:"cert_id" gorm:"index;not null;default:0"`
	CertUUID           string         `json:"-" gorm:"index;type:char(20);default:''"`
	IsRealIPFromHeader int8           `json:"is_real_ip_from_header" gorm:"not null;default:0"`
	RealIPHeader       string         `json:"real_ip_header" gorm:"type:varchar(64)"`
	OriginHostHeader   string         `json:"origin_host_header" gorm:"type:varchar(255);not null;default:''"`
	OriginProtocol     OriginProtocol `json:"origin_protocol" gorm:"type:origin_protocol"`
}

func (SiteConfigModel) TableName() string {
	return "site_config"
}

type SiteConfigRsp struct {
	CertID             int64          `json:"cert_id"`
	IsHttps            int8           `json:"is_https"`
	OriginProtocol     OriginProtocol `json:"origin_protocol"`
	IsRealIPFromHeader int8           `json:"is_real_ip_from_header"`
	RealIPHeader       string         `json:"real_ip_header"`
}

type SiteBasicConfigRsp struct {
	Host               string `json:"host"`
	IsRealIPFromHeader int8   `json:"is_real_ip_from_header"`
	RealIPHeader       string `json:"real_ip_header"`
}

type SiteHttpsReq struct {
	IsHttps      *int8  `json:"is_https" binding:"required"`
	CertID       *int64 `json:"cert_id" binding:"required"`
	IsForceHttps *int8  `json:"is_force_https" binding:"required"`
}

type SiteBasicCfgReq struct {
	IsRealIPFromHeader *int8   `json:"is_real_ip_from_header" binding:"required"`
	RealIPHeader       *string `json:"real_ip_header" binding:"required"`
}

// SiteHttpsConfigsRsp HTTPS相关配置
type SiteHttpsConfigsRsp struct {
	IsHttps      int8   `json:"is_https"`
	CertID       int64  `json:"cert_id"`
	CertName     string `json:"cert_name"`
	IsForceHttps int8   `json:"is_force_https"`
}
