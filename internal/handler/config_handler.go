package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configSvc *service.ConfigService
}

func NewConfigHandler(svc *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configSvc: svc}
}

func (h *ConfigHandler) GetConfigs(c *gin.Context) {
	configs, err := h.configSvc.GetAllConfigs()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, configs)
}

func (h *ConfigHandler) UpdateConfigs(c *gin.Context) {
	var req []struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}

	for _, item := range req {
		if err := h.configSvc.SetConfig(item.Key, item.Value, item.Description); err != nil {
			InternalError(c)
			return
		}
	}
	Success(c, nil)
}
