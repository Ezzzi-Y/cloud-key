package model

import "time"

type KeyStatus string

const (
	KeyStatusActive    KeyStatus = "active"
	KeyStatusExhausted KeyStatus = "exhausted"
	KeyStatusDisabled  KeyStatus = "disabled"
	KeyStatusExpired   KeyStatus = "expired"
)

type Key struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID        uint64     `gorm:"type:bigint;index;not null" json:"tenant_id"`
	Alias           string     `gorm:"type:varchar(255);not null" json:"alias"`
	KeyHash         string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	KeyPrefix       string     `gorm:"type:varchar(50);not null" json:"key_prefix"`
	KeySuffix       string     `gorm:"type:varchar(10);not null" json:"key_suffix"`
	RemainingAmount int64      `gorm:"type:bigint;not null" json:"remaining_amount"`
	Version         int64      `gorm:"type:bigint;not null;default:0" json:"-"`
	Status          KeyStatus  `gorm:"type:varchar(20);not null;default:active" json:"status"`
	CreatedBy       string     `gorm:"type:varchar(100);not null" json:"created_by"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	UsedAt          *time.Time `gorm:"default:null" json:"used_at"`
	ExpireAt        *time.Time `gorm:"default:null" json:"expire_at"`
	MaxUsage        *int64     `gorm:"default:null" json:"max_usage"`
	RateLimit       *int       `gorm:"default:null" json:"rate_limit"`        // 每窗口最大请求数，nil=使用租户默认，0=不限速
	RateLimitWindow *int       `gorm:"default:null" json:"rate_limit_window"` // 窗口大小（秒）
}

func (Key) TableName() string { return "keys" }

func (k *Key) IsUsable() bool {
	return k.Status == KeyStatusActive && k.RemainingAmount > 0
}

func (k *Key) CanDeduct(amount int64) bool {
	return k.IsUsable() && k.RemainingAmount >= amount
}
