package model

import (
	"github.com/rs/xid"
	"github.com/xnile/muxwaf/pkg/utils"
	"gorm.io/gorm"
)

// Model base model
type Model struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID      string `json:"uuid" gorm:"index;type:char(20)" `
	CreatedAt int64  `json:"created_at" gorm:"type:bigint;not null"`
	UpdatedAt int64  `json:"updatedAt" gorm:"type:bigint"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if len(m.UUID) == 0 {
		m.UUID = xid.New().String()
	}
	return nil
}

type ListResp struct {
	List interface{} `json:"list"`
	Meta Meta        `json:"meta"`
}

type Meta struct {
	PageSize int64 `json:"page_size"`
	PageNum  int64 `json:"page_num"`
	Pages    int64 `json:"pages"`
	Total    int64 `json:"total"`
}

func (l *ListResp) SetValue(v interface{}) {
	l.List = v
}

func (l *ListResp) SetMeta(pageSize, pageNum, count int64) {
	meta := Meta{
		PageSize: pageSize,
		PageNum:  pageNum,
		Pages:    utils.CalPage(count, pageSize),
		Total:    count,
	}
	l.Meta = meta
}

// TransferResp antdv 穿梭框返回
type TransferResp struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	Disabled bool   `json:"disabled"`
	Remark   string `json:"remark"`
	//前四个为规定字段
	CreatedAt int64 `json:"created_at"`
}

type ListNewResp struct {
	List interface{} `json:"list"`
	Meta Meta        `json:"meta"`
}

func (l *ListNewResp) SetValue(v interface{}) {
	l.List = v
}

func (l *ListNewResp) SetMeta(pageSize, pageNum, count int64) {
	meta := Meta{
		PageSize: pageSize,
		PageNum:  pageNum,
		Pages:    utils.CalPage(count, pageSize),
		Total:    count,
	}
	l.Meta = meta
}

type GuardArrayRsp []any

type GuardDelArrayRsp []any
