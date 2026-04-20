package handler

import (
	"strings"

	agentmodel "github.com/deepwrite/serivces/gateway/models/agent"
	"github.com/deepwrite/serivces/gateway/pkg/request"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type SessionGroupHandler struct{}

func (h *SessionGroupHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	groups, err := agentmodel.GetSessionGroupsByUser(c.Request.Context(), userID)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to get session groups")
		return
	}

	items := make([]map[string]any, 0, len(groups))
	for _, group := range groups {
		items = append(items, toSessionGroupView(&group))
	}

	response.Success(c, response.SuccessCode, gin.H{"items": items})
}

func (h *SessionGroupHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	var req CreateSessionGroupRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	group, err := agentmodel.CreateSessionGroup(c.Request.Context(), agentmodel.CreateSessionGroupInput{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	})
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to create session group")
		return
	}

	response.Success(c, response.SuccessCreatedCode, toSessionGroupView(&group))
}

func (h *SessionGroupHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	groupID := strings.TrimSpace(c.Param("id"))
	if groupID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "group_id is required")
		return
	}

	isOwner, err := agentmodel.IsSessionGroupOwner(c.Request.Context(), groupID, userID)
	if err != nil {
		response.Failed(c, 404, "Session group not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to update this group")
		return
	}

	var req UpdateSessionGroupRequest
	if ok := request.ValidateStruct(c, &req); !ok {
		return
	}

	input := agentmodel.UpdateSessionGroupInput{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
	}

	group, err := agentmodel.UpdateSessionGroup(c.Request.Context(), groupID, input)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to update session group")
		return
	}

	response.Success(c, response.SuccessCode, toSessionGroupView(&group))
}

func (h *SessionGroupHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Failed(c, 401, "Unauthorized")
		return
	}

	groupID := strings.TrimSpace(c.Param("id"))
	if groupID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "group_id is required")
		return
	}

	isOwner, err := agentmodel.IsSessionGroupOwner(c.Request.Context(), groupID, userID)
	if err != nil {
		response.Failed(c, 404, "Session group not found")
		return
	}
	if !isOwner {
		response.Failed(c, 403, "You don't have permission to delete this group")
		return
	}

	if err := agentmodel.DeleteSessionGroup(c.Request.Context(), groupID); err != nil {
		response.Failed(c, response.ErrorUnknownCode, "Failed to delete session group")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": groupID, "deleted": true})
}

func toSessionGroupView(group *agentmodel.SessionGroup) map[string]any {
	return map[string]any{
		"id":          group.ID,
		"user_id":     group.UserID,
		"name":        group.Name,
		"description": group.Description,
		"color":       group.Color,
		"icon":        group.Icon,
		"sort_order":  group.SortOrder,
		"created_at":  group.CreatedAt,
		"updated_at":  group.UpdatedAt,
	}
}

type CreateSessionGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

type UpdateSessionGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	SortOrder   *int   `json:"sort_order"`
}
