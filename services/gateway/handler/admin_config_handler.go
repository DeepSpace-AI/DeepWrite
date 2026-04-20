package handler

import (
	"strings"

	systemmodel "github.com/deepwrite/serivces/gateway/models/system"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminConfigHandler struct{}

func (h *AdminConfigHandler) List(c *gin.Context) {
	configs, err := systemmodel.GetAllConfigs(c.Request.Context())
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get configs")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"configs":     configs,
		"definitions": systemmodel.ConfigDefinitions,
	})
}

func (h *AdminConfigHandler) Get(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		response.Failed(c, response.ErrorBadRequestCode, "key is required")
		return
	}

	value, err := systemmodel.GetConfig(c.Request.Context(), key)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get config")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"key":   key,
		"value": value,
	})
}

func (h *AdminConfigHandler) Update(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		response.Failed(c, response.ErrorBadRequestCode, "key is required")
		return
	}

	var req UpdateConfigRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	userID := c.GetString("user_id")

	var description string
	for _, def := range systemmodel.ConfigDefinitions {
		if def.Key == key {
			description = def.Description
			break
		}
	}

	if err := systemmodel.SetConfig(c.Request.Context(), key, req.Value, description, &userID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update config")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"key":   key,
		"value": req.Value,
	})
}

func (h *AdminConfigHandler) BatchUpdate(c *gin.Context) {
	var req BatchUpdateConfigRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	userID := c.GetString("user_id")

	for key, value := range req.Configs {
		var description string
		for _, def := range systemmodel.ConfigDefinitions {
			if def.Key == key {
				description = def.Description
				break
			}
		}

		if err := systemmodel.SetConfig(c.Request.Context(), key, value, description, &userID); err != nil {
			response.Failed(c, response.ErrorUnknownCode, "Failed to update config: "+key)
			return
		}
	}

	configs, _ := systemmodel.GetAllConfigs(c.Request.Context())
	response.Success(c, response.SuccessCode, gin.H{"configs": configs})
}

type UpdateConfigRequest struct {
	Value string `json:"value"`
}

type BatchUpdateConfigRequest struct {
	Configs map[string]string `json:"configs" binding:"required"`
}
