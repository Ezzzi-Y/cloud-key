package model

import "time"

type TenantStatus string

const (
	TenantStatusActive   TenantStatus = "active"
	TenantStatusExpired  TenantStatus = "expired"
	TenantStatusDisabled TenantStatus = "disabled"
)

type Tenant struct {
	ID              uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Status          TenantStatus `gorm:"type:varchar(20);not null;default:active" json:"status"`
	ExpireAt        *time.Time   `gorm:"default:null" json:"expire_at"`
	KeyPrefix       string       `gorm:"type:varchar(20);not null;default:sk-" json:"key_prefix"`
	KeyLength       int          `gorm:"type:int;not null;default:32" json:"key_length"`
	KeySuffixLength int          `gorm:"type:int;not null;default:4" json:"key_suffix_length"`
	CreatedAt       time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Tenant) TableName() string { return "tenants" }
