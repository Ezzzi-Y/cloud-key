package model

import "time"

type BalanceLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64    `gorm:"type:bigint;index;not null" json:"tenant_id"`
	KeyID        uint64    `gorm:"type:bigint;index;not null" json:"key_id"`
	KeyAlias     string    `gorm:"type:varchar(255);not null" json:"key_alias"`
	Delta        int64     `gorm:"type:bigint;not null" json:"delta"`
	BeforeAmount int64     `gorm:"type:bigint;not null" json:"before_amount"`
	AfterAmount  int64     `gorm:"type:bigint;not null" json:"after_amount"`
	Operator     string    `gorm:"type:varchar(100);not null" json:"operator"`
	Remark       string    `gorm:"type:varchar(500);not null;default:''" json:"remark"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (BalanceLog) TableName() string { return "balance_logs" }
