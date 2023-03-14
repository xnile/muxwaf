package model

type SiteOriginModel struct {
	Model
	SiteID    int64  `json:"site_id" gorm:"index;not null"`
	HttpPort  int16  `json:"http_port"  gorm:"type:smallint;not null;default:80" binding:"required,numeric"`
	HttpsPort int16  `json:"https_port"  gorm:"type:smallint;not null;default:443" `
	Weight    int16  `json:"weight" gorm:"type:smallint;not null;default:100" binding:"required,numeric"`
	Type      int16  `json:"-" gorm:"type:smallint;not null;default:1;comment:1 as ip,2 as domain"` // not use
	Host      string `json:"host" gorm:"type:varchar(253);not null;" binding:"required,ip"`         // 域名最大长度
}

func (SiteOriginModel) TableName() string {
	return "site_origin"
}

type SiteOriginRsp struct {
	ID       int64  `json:"id"`
	HttpPort int16  `json:"http_port"`
	Weight   int16  `json:"weight"`
	Host     string `json:"host"`
}

type SiteHttpsRsp struct {
	Https    bool   `json:"https"`
	CertName string `json:"certName"`
}
