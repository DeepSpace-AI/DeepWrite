package handler

import (
	"strings"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/config"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/deepwrite/serivces/gateway/pkg/worker"
	"github.com/gin-gonic/gin"
)

type MemoryTriggerHandler struct{}

func (h *MemoryTriggerHandler) TriggerExtraction(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	sessionID := strings.TrimSpace(c.Param("id"))
	if sessionID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "session_id is required")
		return
	}

	isOwner, err := agentmodel.IsSessionOwner(c.Request.Context(), sessionID, userID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to access this session")
		return
	}

	session, err := agentmodel.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}

	cfg := config.GetGlobalConfig()
	workerClient := worker.NewClient(
		worker.WithBaseURL(cfg.Worker.URL),
		worker.WithAuthToken(cfg.Worker.Token),
	)

	taskResp, err := workerClient.ExtractMemories(
		c.Request.Context(),
		sessionID,
		session.AgentID,
		session.UserID,
		session.WorkspaceID,
	)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to trigger memory extraction: "+err.Error())
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"task_id":   taskResp.TaskID,
		"task_name": taskResp.TaskName,
		"queue":     taskResp.Queue,
	})
}

func (h *MemoryTriggerHandler) GetExtractionStatus(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "task_id is required")
		return
	}

	cfg := config.GetGlobalConfig()
	workerClient := worker.NewClient(
		worker.WithBaseURL(cfg.Worker.URL),
		worker.WithAuthToken(cfg.Worker.Token),
	)

	status, err := workerClient.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get task status: "+err.Error())
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"task_id": status.TaskID,
		"status":  status.Status,
		"ready":   status.Ready,
		"result":  status.Result,
		"error":   status.Error,
	})
}
