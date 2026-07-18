package model

import "time"

type SysConfig struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Key         string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value       string    `gorm:"type:varchar(500);not null" json:"value"`
	Description string    `gorm:"type:varchar(500)" json:"description"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SysConfig) TableName() string { return "configs" }
