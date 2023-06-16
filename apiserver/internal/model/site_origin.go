package model

import (
	"database/sql/driver"
)

type OriginType string

const (
	IPOrigin     OriginType = "ip"
	DomainOrigin OriginType = "domain"
)

func (p *OriginType) Scan(value interface{}) error {
	*p = OriginType(value.(string))
	return nil
}

func (p OriginType) Value() (driver.Value, error) {
	return string(p), nil
}

type OriginProtocol string

const (
	HTTPOriginProtocol  OriginProtocol = "http"
	HTTPSOriginProtocol OriginProtocol = "https"

	//暂不支持
	//FollowOriginProtocol OriginProtocol = "follow"
)

func (p *OriginProtocol) Scan(value interface{}) error {
	*p = OriginProtocol(value.(string))
	//*p = OriginProtocol(value.([]byte))
	return nil
}

func (p OriginProtocol) Value() (driver.Value, error) {
	return string(p), nil
}

type SiteOriginModel struct {
	Model
	SiteID int64 `json:"site_id" gorm:"index;not null"`
	//HttpPort  int16 `json:"http_port"  gorm:"type:smallint;not null;default:80" binding:"required,numeric"`
	//HttpsPort int16 `json:"https_port"  gorm:"type:smallint;not null;default:443" `
	//
	//Type int16  `json:"-" gorm:"type:smallint;not null;default:1;comment:1 as ip,2 as domain"` // not use
	//Host string `json:"host" gorm:"type:varchar(253);not null;" binding:"required,ip"`         // 域名最大长度

	Port     int16          `json:"port"  gorm:"type:smallint;not null;default:80" binding:"gte=1,lte=65535"`
	Addr     string         `json:"addr" gorm:"type:varchar(253);not null;default:127.0.0.1" binding:"required,ipv4|fqdn"`
	Weight   int16          `json:"weight" gorm:"type:smallint;not null;default:100" binding:"gte=0,lte=100"`
	Kind     OriginType     `json:"-"  gorm:"type:origin_type"`
	Protocol OriginProtocol `json:"-" gorm:"type:origin_protocol"`
}

func (SiteOriginModel) TableName() string {
	return "site_origin"
}

// TODO: delete
type SiteOriginRsp struct {
	ID     int64  `json:"id"`
	Addr   string `json:"addr"`
	Port   int16  `json:"port"`
	Weight int16  `json:"weight"`
	//Kind     OriginType     `json:"kind"`
	//Protocol OriginProtocol `json:"protocol"`
}

type SiteHttpsRsp struct {
	Https    bool   `json:"https"`
	CertName string `json:"certName"`
}

type OriginCfgReq struct {
	OriginProtocol   OriginProtocol     `json:"origin_protocol"`
	OriginHostHeader string             `json:"origin_host_header"`
	Origins          []*SiteOriginModel `json:"origins" binding:"gt=0,dive"`
}

type OriginCfgRsp struct {
	OriginProtocol   OriginProtocol     `json:"origin_protocol"`
	OriginHostHeader string             `json:"origin_host_header"`
	Origins          []*SiteOriginModel `json:"origins"`
}
