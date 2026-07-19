package service

import (
	"CloudKey/internal/model"

	"gorm.io/gorm"
)

type LoginLogService struct {
	db *gorm.DB
}

func NewLoginLogService(db *gorm.DB) *LoginLogService {
	return &LoginLogService{db: db}
}

func (s *LoginLogService) RecordLogin(userID uint64, tenantID *uint64, ip, userAgent string, success bool) error {
	status := model.LoginStatusFailed
	if success {
		status = model.LoginStatusSuccess
	}
	return s.db.Create(&model.LoginLog{
		UserID: userID, TenantID: tenantID, IP: ip, UserAgent: userAgent, Status: status,
	}).Error
}

func (s *LoginLogService) ListLoginLogs(page, pageSize int, tenantID *uint64) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64

	db := s.db.Model(&model.LoginLog{})
	if tenantID != nil {
		db = db.Where("tenant_id = ?", *tenantID)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
