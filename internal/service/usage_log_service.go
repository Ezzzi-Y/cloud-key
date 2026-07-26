package service

import (
	"CloudKey/internal/model"
	"time"

	"gorm.io/gorm"
)

type UsageLogService struct {
	db *gorm.DB
}

func NewUsageLogService(db *gorm.DB) *UsageLogService {
	return &UsageLogService{db: db}
}

type RecordUsageParams struct {
	TenantID       uint64
	KeyID          uint64
	KeyAlias       string
	Amount         int64
	IP             string
	UserAgent      string
	RequestParams  string
	ResponseStatus int
}

func (s *UsageLogService) Record(params RecordUsageParams) error {
	return s.db.Create(&model.UsageLog{
		TenantID:       params.TenantID,
		KeyID:          params.KeyID,
		KeyAlias:       params.KeyAlias,
		Amount:         params.Amount,
		IP:             params.IP,
		UserAgent:      params.UserAgent,
		RequestParams:  params.RequestParams,
		ResponseStatus: params.ResponseStatus,
		CreatedAt:      time.Now(),
	}).Error
}

type UsageLogQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	KeyAlias  string `form:"key_alias"`
	IP        string `form:"ip"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

func (s *UsageLogService) ListLogs(query UsageLogQuery, tenantID uint64) ([]model.UsageLog, int64, error) {
	var logs []model.UsageLog
	var total int64

	db := s.db.Model(&model.UsageLog{}).Where("tenant_id = ?", tenantID)
	if query.KeyAlias != "" {
		db = db.Where("key_alias = ?", query.KeyAlias)
	}
	if query.IP != "" {
		db = db.Where("ip = ?", query.IP)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (s *UsageLogService) ExportLogs(query UsageLogQuery, tenantID uint64) ([]model.UsageLog, error) {
	var logs []model.UsageLog
	db := s.db.Model(&model.UsageLog{}).Where("tenant_id = ?", tenantID)
	if query.KeyAlias != "" {
		db = db.Where("key_alias = ?", query.KeyAlias)
	}
	if query.IP != "" {
		db = db.Where("ip = ?", query.IP)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}
	if err := db.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
