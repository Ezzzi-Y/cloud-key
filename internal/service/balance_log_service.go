package service

import (
	"CloudKey/internal/model"
	"time"

	"gorm.io/gorm"
)

type BalanceLogService struct {
	db *gorm.DB
}

func NewBalanceLogService(db *gorm.DB) *BalanceLogService {
	return &BalanceLogService{db: db}
}

type RecordBalanceParams struct {
	TenantID     uint64
	KeyID        uint64
	KeyAlias     string
	KeySuffix    string
	Delta        int64
	BeforeAmount int64
	AfterAmount  int64
	Operator     string
	Remark       string
}

func (s *BalanceLogService) Record(params RecordBalanceParams) error {
	return s.db.Create(&model.BalanceLog{
		TenantID:     params.TenantID,
		KeyID:        params.KeyID,
		KeyAlias:     params.KeyAlias,
		KeySuffix:    params.KeySuffix,
		Delta:        params.Delta,
		BeforeAmount: params.BeforeAmount,
		AfterAmount:  params.AfterAmount,
		Operator:     params.Operator,
		Remark:       params.Remark,
		CreatedAt:    time.Now(),
	}).Error
}

type BalanceLogQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Alias     string `form:"alias"`      // 别名前缀搜索
	KeySuffix string `form:"key_suffix"` // 后缀精准搜索
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

func (s *BalanceLogService) ListLogs(query BalanceLogQuery, tenantID uint64) ([]model.BalanceLog, int64, error) {
	var logs []model.BalanceLog
	var total int64

	db := s.db.Model(&model.BalanceLog{}).Where("tenant_id = ?", tenantID)
	if query.Alias != "" {
		db = db.Where("key_alias LIKE ?", query.Alias+"%")
	}
	if query.KeySuffix != "" {
		db = db.Where("key_suffix = ?", query.KeySuffix)
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

func (s *BalanceLogService) ExportLogs(query BalanceLogQuery, tenantID uint64) ([]model.BalanceLog, error) {
	var logs []model.BalanceLog
	db := s.db.Model(&model.BalanceLog{}).Where("tenant_id = ?", tenantID)
	if query.Alias != "" {
		db = db.Where("key_alias LIKE ?", query.Alias+"%")
	}
	if query.KeySuffix != "" {
		db = db.Where("key_suffix = ?", query.KeySuffix)
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
