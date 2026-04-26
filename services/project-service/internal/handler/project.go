package handler

import (
	"net/http"
	"strconv"

	"github.com/deepwrite/project-service/internal/models"
	"github.com/deepwrite/project-service/internal/service"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	project, err := h.service.Create(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": project})
}

func (h *ProjectHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	projects, total, err := h.service.ListByUser(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    projects,
		"meta":    gin.H{"limit": limit, "offset": offset, "total": total},
	})
}

func (h *ProjectHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("id")

	project, err := h.service.GetByID(c.Request.Context(), projectID, userID.(string))
	if err != nil {
		if err == models.ErrProjectNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "PROJECT_NOT_FOUND", "message": "Project not found"}})
			return
		}
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": project})
}

func (h *ProjectHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("id")

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	project, err := h.service.Update(c.Request.Context(), projectID, userID.(string), &req)
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": project})
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), projectID, userID.(string)); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Project archived"}})
}

func (h *ProjectHandler) Archive(c *gin.Context) {
	h.Delete(c)
}

func (h *ProjectHandler) Unarchive(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("id")

	if err := h.service.Unarchive(c.Request.Context(), projectID, userID.(string)); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Project unarchived"}})
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("id")

	var req models.AddProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	if err := h.service.AddMember(c.Request.Context(), projectID, userID.(string), &req); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Member added"}})
}

func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("id")
	targetUserID := c.Param("user_id")

	if err := h.service.RemoveMember(c.Request.Context(), projectID, userID.(string), targetUserID); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Member removed"}})
}

func (h *ProjectHandler) GetMembers(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("id")

	members, err := h.service.GetMembers(c.Request.Context(), projectID, userID.(string))
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

func (h *ProjectHandler) UpdateMember(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("id")
	targetUserID := c.Param("user_id")

	var req models.UpdateProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	if err := h.service.UpdateMemberRole(c.Request.Context(), projectID, userID.(string), targetUserID, &req); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Member role updated"}})
}
