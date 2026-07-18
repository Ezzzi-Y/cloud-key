package model

import "time"

type LoginStatus string

const (
	LoginStatusSuccess LoginStatus = "success"
	LoginStatusFailed  LoginStatus = "failed"
)

type LoginLog struct {
	ID        uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID   uint64      `gorm:"type:bigint;index;not null" json:"admin_id"`
	IP        string      `gorm:"type:varchar(50);not null" json:"ip"`
	UserAgent string      `gorm:"type:varchar(500)" json:"user_agent"`
	Status    LoginStatus `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt time.Time   `gorm:"autoCreateTime" json:"created_at"`
}

func (LoginLog) TableName() string { return "login_logs" }
