package model

type AttackLogModel struct {
	Model
	SiteID        string `json:"site_id" gorm:"index;not null;default:0"`
	Host          string `json:"host"`
	RemoteAddr    string `json:"remote_addr"`
	RealClientIp  string `json:"real_client_ip"`
	RequestID     string `json:"request_id" gorm:"index;not null;default:''"`
	RequestPath   string `json:"request_path"`
	RequestMethod string `json:"request_method"`
	RequestTime   int64  `json:"request_time"`
	ProcessTime   int64  `json:"process_time"`
	RuleType      string `json:"rule_type"`
	Action        int8   `json:"action" gorm:"index:not null;default:0"`
	WorkerID      int64  `json:"worker_id"`
}

func (AttackLogModel) TableName() string {
	return "sample_log"
}
