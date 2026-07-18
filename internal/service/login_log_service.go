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

func (s *LoginLogService) RecordLogin(adminID uint64, ip, userAgent string, success bool) error {
	status := model.LoginStatusFailed
	if success {
		status = model.LoginStatusSuccess
	}
	return s.db.Create(&model.LoginLog{
		AdminID: adminID, IP: ip, UserAgent: userAgent, Status: status,
	}).Error
}

func (s *LoginLogService) ListLoginLogs(page, pageSize int) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64

	s.db.Model(&model.LoginLog{}).Count(&total)
	offset := (page - 1) * pageSize
	s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)

	return logs, total, nil
}
