package model

import "time"

type UsageLog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       uint64    `gorm:"type:bigint;index;not null" json:"tenant_id"`
	KeyID          uint64    `gorm:"type:bigint;index;not null" json:"key_id"`
	KeyAlias       string    `gorm:"type:varchar(255);not null" json:"key_alias"`
	KeySuffix      string    `gorm:"type:varchar(10);not null" json:"key_suffix"`
	Amount         int64     `gorm:"type:bigint;not null" json:"amount"`
	IP             string    `gorm:"type:varchar(50);not null" json:"ip"`
	UserAgent      string    `gorm:"type:varchar(500)" json:"user_agent"`
	RequestParams  string    `gorm:"type:text" json:"request_params"`
	ResponseStatus int       `gorm:"type:int;not null" json:"response_status"`
	CreatedAt      time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (UsageLog) TableName() string { return "usage_logs" }
