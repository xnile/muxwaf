package model

import "github.com/lib/pq"

type CertModel struct {
	Model
	Name      string         `json:"name" gorm:"uniqueIndex;type:varchar(128);not null" binding:"required"`
	Cert      string         `json:"cert" gorm:"type:text;not null" binding:"required"`
	Key       string         `json:"key" gorm:"type:text;not null" binding:"required"`
	CN        string         `json:"cn" gorm:"type:varchar(255);not null;default:''"`
	Sans      pq.StringArray `json:"sans" gorm:"type:varchar(255)[];not null;default:'{}'"`
	BeginTime int64          `json:"begin_time" gorm:"type:bigint;not null;default:0"`
	EndTime   int64          `json:"end_time" gorm:"type:bigint;not null;default:0"`
}

func (CertModel) TableName() string {
	return "cert"
}

// CertResp 返回证书列表
type CertResp struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	CN        string         `json:"cn"`
	Sans      []string       `json:"sans"`
	BeginTime int64          `json:"begin_time"`
	EndTime   int64          `json:"end_time"`
	Sites     []CertBindSite `json:"sites"`
}

type CertBindSite struct {
	ID     int64  `json:"id"`
	Domain string `json:"domain"`
}

// CertCandidateResp 返回所有证书
type CertCandidateResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
