package model

import "github.com/lib/pq"

type SiteRegionBlacklistModel struct {
	Model
	SiteID    int64          `json:"site_id" gorm:"not null"`
	Countries pq.StringArray `json:"countries" gorm:"type:varchar(20)[];not null;default:'{}'"`
	Regions   pq.StringArray `json:"regions" gorm:"type:varchar(20)[];not null;default:'{}'"`
	MatchMode int8           `json:"match_mode" gorm:"not null;default:0;comment:0 blacklist mode,1 whitelist mode"`
	Status    int8           `json:"status" gorm:"not null;default:1"`
}

func (SiteRegionBlacklistModel) TableName() string {
	return "site_region_blacklist"
}

type SiteRegionBlacklistRsp struct {
	Countries pq.StringArray `json:"countries"`
	Regions   pq.StringArray `json:"regions"`
	MatchMode int8           `json:"match_mode"`
}
