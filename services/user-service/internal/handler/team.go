package handler

import (
	"net/http"

	"github.com/deepwrite/user-service/internal/models"
	"github.com/deepwrite/user-service/internal/service"
	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	team, err := h.teamService.Create(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": team})
}

func (h *TeamHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")

	teams, err := h.teamService.ListByUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": teams})
}

func (h *TeamHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	teamID := c.Param("id")

	team, err := h.teamService.GetByID(c.Request.Context(), teamID, userID.(string))
	if err != nil {
		if err == models.ErrTeamNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "TEAM_NOT_FOUND", "message": "Team not found"}})
			return
		}
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": team})
}

func (h *TeamHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	teamID := c.Param("id")

	var req models.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	team, err := h.teamService.Update(c.Request.Context(), teamID, userID.(string), &req)
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": team})
}

func (h *TeamHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	teamID := c.Param("id")

	if err := h.teamService.Delete(c.Request.Context(), teamID, userID.(string)); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Team deleted"}})
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	teamID := c.Param("id")

	var req models.AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	if err := h.teamService.AddMember(c.Request.Context(), teamID, userID.(string), &req); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Member added"}})
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	teamID := c.Param("id")
	targetUserID := c.Param("user_id")

	if err := h.teamService.RemoveMember(c.Request.Context(), teamID, userID.(string), targetUserID); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Member removed"}})
}

func (h *TeamHandler) GetMembers(c *gin.Context) {
	userID, _ := c.Get("userID")
	teamID := c.Param("id")

	members, err := h.teamService.GetMembers(c.Request.Context(), teamID, userID.(string))
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": members})
}

func (h *TeamHandler) UpdateMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	teamID := c.Param("id")
	targetUserID := c.Param("user_id")

	var req models.UpdateTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	if err := h.teamService.UpdateMemberRole(c.Request.Context(), teamID, userID.(string), targetUserID, &req); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Member role updated"}})
}
