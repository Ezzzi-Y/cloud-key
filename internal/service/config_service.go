package service

import (
	"CloudKey/internal/model"

	"gorm.io/gorm"
)

type ConfigService struct {
	db *gorm.DB
}

func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{db: db}
}

func (s *ConfigService) GetConfig(key string) (string, error) {
	var cfg model.SysConfig
	if err := s.db.Where("`key` = ?", key).First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return cfg.Value, nil
}

func (s *ConfigService) GetAllConfigs() ([]model.SysConfig, error) {
	var configs []model.SysConfig
	if err := s.db.Order("id ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *ConfigService) SetConfig(key, value, description string) error {
	var cfg model.SysConfig
	result := s.db.Where("`key` = ?", key).First(&cfg)
	if result.Error == gorm.ErrRecordNotFound {
		return s.db.Create(&model.SysConfig{Key: key, Value: value, Description: description}).Error
	}
	if result.Error != nil {
		return result.Error
	}
	updates := map[string]interface{}{"value": value}
	if description != "" {
		updates["description"] = description
	}
	return s.db.Model(&cfg).Updates(updates).Error
}

func (s *ConfigService) InitDefaultConfigs() error {
	defaults := []model.SysConfig{
		{Key: "key_prefix", Value: "sk-", Description: "卡密默认前缀"},
		{Key: "key_length", Value: "32", Description: "卡密随机部分长度"},
		{Key: "key_suffix_length", Value: "4", Description: "卡密后缀长度"},
		{Key: "record_request_params", Value: "false", Description: "是否记录请求参数"},
		{Key: "jwt_expire_hours", Value: "24", Description: "JWT 过期时间（小时）"},
	}

	for _, d := range defaults {
		var count int64
		s.db.Model(&model.SysConfig{}).Where("`key` = ?", d.Key).Count(&count)
		if count == 0 {
			if err := s.db.Create(&d).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
