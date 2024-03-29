package model

import "github.com/lib/pq"

type BlacklistIPGuard struct {
	IP   string `json:"ip"`
	UUID string `json:"id"`
}

//type BlacklistRegionGuard struct {
//	SiteID    string   `json:"site_id"`
//	Countries []string `json:"countries"`
//	Regions   []string `json:"regions"`
//	MatchMode int8     `json:"match_mode"`
//}

type WhitelistIPGuard struct {
	UUID string `json:"id"`
	IP   string `json:"ip"`
}

type WhitelistURLGuard struct {
	UUID      string `json:"id"`
	SiteID    string `json:"site_id"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	MatchMode int8   `json:"match_mode"`
}

type RateLimitGuard struct {
	UUID      string `json:"id"`
	SiteID    string `json:"site_id"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	Limit     int64  `json:"limit"`
	Window    int64  `json:"window"`
	MatchMode int8   `json:"match_mode"`
}

type CertificateGuard struct {
	UUID string `json:"id"`
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

type SampleLogUploadGuard struct {
	IsSampleLogUpload       int8   `json:"is_sample_log_upload"`
	SampleLogUploadAPI      string `json:"sample_log_upload_api"`
	SampleLogUploadAPIToken string `json:"sample_log_upload_api_token"`
}

type SiteConfigGuard struct {
	CertID             string              `json:"cert_id"`
	IsHttps            int8                `json:"is_https"`
	IsForceHttps       int8                `json:"is_force_https"`
	IsRealIPFromHeader int8                `json:"is_real_ip_from_header"`
	RealIPHeader       string              `json:"real_ip_header"`
	Origin             *SiteOriginCfgGuard `json:"origin,omitempty"`
}

type SiteOriginGuard struct {
	Addr     string         `json:"addr"`
	Port     int16          `json:"port"`
	Weight   int8           `json:"weight"`
	Kind     OriginType     `json:"kind"`
	Protocol OriginProtocol `json:"protocol"`
}

type SiteGuard struct {
	UUID    string           `json:"id"`
	Host    string           `json:"host"`
	Configs *SiteConfigGuard `json:"config"`
	//Origins []*SiteOriginGuard `json:"origins"`
}

type SiteRegionBlacklistGuard struct {
	SiteID    string         `json:"site_id"`
	Countries pq.StringArray `json:"countries"`
	Regions   pq.StringArray `json:"regions"`
	MatchMode int8           `json:"match_mode"`
}

type RulesGuard struct {
	WhitelistIP     []*WhitelistIPGuard         `json:"whitelist_ip"`
	WhitelistURL    []*WhitelistURLGuard        `json:"whitelist_url"`
	BlacklistIP     []*BlacklistIPGuard         `json:"blacklist_ip"`
	BlacklistRegion []*SiteRegionBlacklistGuard `json:"blacklist_region"`
	RateLimit       []*RateLimitGuard           `json:"rate_limit"`
}

type SiteOriginCfgGuard struct {
	OriginProtocol   OriginProtocol     `json:"origin_protocol"`
	OriginHostHeader string             `json:"origin_host_header"`
	Origins          []*SiteOriginGuard `json:"origins"`
}

// ConfigsSyncGuard Guard全量配置
type ConfigsSyncGuard struct {
	SampleLog    *SampleLogUploadGuard `json:"sample_log"`
	Sites        []*SiteGuard          `json:"sites"`
	Certificates []*CertificateGuard   `json:"certificates"`
	Rules        *RulesGuard           `json:"rules"`
}
