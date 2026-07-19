package model

import "time"

type Admin struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	TotpSecret   string    `gorm:"type:varchar(255)" json:"-"`
	TotpSetup    bool      `gorm:"default:false" json:"totp_setup"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Admin) TableName() string { return "admins" }
