package model

type NodeModel struct {
	Model
	Name                     string `json:"name" gorm:"not null;default:''"`
	IPOrDomain               string `json:"ip_or_domain" gorm:"type:varchar(255)" binding:"required,ipv4|hostname_rfc1123"`
	Port                     int16  `json:"port" binding:"required,min=1,max=65535"`
	Status                   int8   `json:"status" gorm:"not null;default:1"`
	LastSyncAt               int64  `json:"last_sync_at" gorm:"not null;default:0"`
	LastSyncStatus           int8   `json:"last_sync_status" gorm:"not null;default:0"`
	IsSampleLogUpload       int8   `json:"is_sample_log_upload" gorm:"not null;default:0"`
	SampleLogUploadAPI      string `json:"sample_log_upload_api" gorm:"not null;default:''"`
	SampleLogUploadAPIToken string `json:"sample_log_upload_api_token" gorm:"not null;default:''"`
}

func (NodeModel) TableName() string {
	return "node"
}
