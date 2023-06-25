package event

import (
	"bytes"
	"encoding/json"
	"github.com/xnile/muxwaf/internal/model"
	"github.com/xnile/muxwaf/pkg/logx"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type OpType string
type Target string

const (
	OpTypeAdd    OpType = "POST"
	OpTypeUpdate        = "PUT"
	OpTypeDel           = "DELETE"
	OpTypeSync          = "POST"
)

const (
	BlacklistIP     Target = "/api/blacklist/ip"
	BlacklistRegion        = "/api/blacklist/region"
	WhitelistIP            = "/api/whitelist/ip"
	WhitelistURL           = "/api/whitelist/url"
	RateLimit              = "/api/rate-limit"
	Certificate            = "/api/certificates"
	Site                   = "/api/sites"
	SampleLogUpload        = "/api/sys/configs/sample_log_upload"
	All                    = "/api/sys/configs"
)

var _ Handler = &defaultHandler{}

var guardClient = http.Client{Timeout: 30 * time.Second}

type defaultHandler struct {
	gDB *gorm.DB
}

func NewDefaultHandler(gDB *gorm.DB) *defaultHandler {
	return &defaultHandler{
		gDB: gDB,
	}
}

func (h *defaultHandler) Next(event Event) {
	data, err := json.Marshal(event.Payload)
	if err != nil {
		logx.Error("serialize data err: ", err)
		return
	}

	nodes := make([]*model.NodeModel, 0)
	mapGuards := make(map[int64]string)
	if err := h.gDB.Select("id", "Addr", "port").Where("status = ?", 1).Find(&nodes).Error; err != nil {
		logx.Error("[event] Failed to get nodes: ", err.Error())
		return
	}
	for _, node := range nodes {
		mapGuards[node.ID] = node.Addr + ":" + strconv.Itoa(int(node.Port))
	}

	for guardID, guard := range mapGuards {
		if len(event.WorkNodes) > 0 {
			for _, nodeID := range event.WorkNodes {
				if guardID == nodeID {
					goto SYNC
				}
			}
			continue
		}

	SYNC:
		apiURL := "http://" + guard + string(event.Target)
		request, _ := http.NewRequest(string(event.OpType), apiURL, bytes.NewReader(data))
		request.Header.Set("Content-Type", "application/json")
		rsp, err := guardClient.Do(request)
		if err != nil {
			logx.Error("[guard]Failed to sync configs: ", err)
			return
		}
		rspBody := make(map[string]any)
		if err := json.NewDecoder(rsp.Body).Decode(&rspBody); err != nil {
			logx.Error("【guard]Failed to decode json: ", err)
		}

		// debug
		{
			logx.Info("[guard]Sync api: ", apiURL)
			logx.Debug("[guard]Sync data: ", string(data))
			logx.Info("[guard]Sync result: ", rspBody)
			logx.Info("[guard]Sync status: ", rsp.Status)
		}

		// 更新node状态
		{
			var rspStatus int8 = -1
			if rsp.StatusCode == 200 {
				rspStatus = 1
			}
			if err := h.gDB.Select("LastSyncAt", "LastSyncStatus").Where("id = ?", guardID).Updates(model.NodeModel{
				LastSyncAt:     time.Now().Unix(),
				LastSyncStatus: rspStatus,
			}).Error; err != nil {
				logx.Error("[event] Failed to update node sync status: ", err)
			}
		}
	}
}
