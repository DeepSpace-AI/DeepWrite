package handler

import (
	"fmt"
	"strings"
	"time"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type SessionExportHandler struct{}

func (h *SessionExportHandler) ExportMarkdown(c *gin.Context) {
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
		response.Failed(c, 403, "You don't have permission to export this session")
		return
	}

	session, err := agentmodel.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}

	messages, err := agentmodel.GetMessagesBySession(c.Request.Context(), sessionID, 0, 0)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get messages")
		return
	}

	markdown := generateMarkdown(&session, messages)

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.md", sanitizeFilename(session.Title)))
	c.String(200, markdown)
}

func (h *SessionExportHandler) ExportJSON(c *gin.Context) {
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
		response.Failed(c, 403, "You don't have permission to export this session")
		return
	}

	session, err := agentmodel.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		response.Failed(c, 404, "Session not found")
		return
	}

	messages, err := agentmodel.GetMessagesBySession(c.Request.Context(), sessionID, 0, 0)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get messages")
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", sanitizeFilename(session.Title)))
	c.JSON(200, gin.H{
		"session":  toSessionView(&session),
		"messages": messages,
	})
}

func generateMarkdown(session *agentmodel.Session, messages []agentmodel.Message) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", session.Title))
	sb.WriteString(fmt.Sprintf("> Created: %s\n", session.CreatedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("> Last updated: %s\n\n", session.UpdatedAt.Format(time.RFC3339)))

	if session.Summary != "" {
		sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", session.Summary))
	}

	sb.WriteString("---\n\n")

	for _, msg := range messages {
		var role string
		switch msg.Role {
		case "user":
			role = "👤 **User**"
		case "assistant":
			role = "🤖 **Assistant**"
		default:
			role = fmt.Sprintf("**%s**", msg.Role)
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", role))
		sb.WriteString(msg.Content)
		sb.WriteString("\n\n")
		sb.WriteString(fmt.Sprintf("*%s*\n\n", msg.CreatedAt.Format(time.RFC3339)))
		sb.WriteString("---\n\n")
	}

	return sb.String()
}

func sanitizeFilename(name string) string {
	if name == "" {
		return "conversation"
	}
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "'", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	if len(name) > 100 {
		name = name[:100]
	}
	return name
}
