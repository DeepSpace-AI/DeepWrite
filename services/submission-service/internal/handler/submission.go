package handler

import (
	"net/http"
	"strconv"

	"github.com/deepwrite/submission-service/internal/models"
	"github.com/deepwrite/submission-service/internal/service"
	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	service *service.SubmissionService
}

func NewSubmissionHandler(service *service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{service: service}
}

func (h *SubmissionHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	submission, err := h.service.Create(userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": submission})
}

func (h *SubmissionHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	submission, err := h.service.GetByID(id, userID.(string))
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		if err == models.ErrSubmissionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "NOT_FOUND", "message": "Submission not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": submission})
}

func (h *SubmissionHandler) ListByProject(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID := c.Param("project_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	submissions, total, err := h.service.ListByProject(projectID, userID.(string), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    submissions,
		"meta":    gin.H{"page": page, "limit": limit, "total": total},
	})
}

func (h *SubmissionHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	var req models.UpdateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	submission, err := h.service.Update(id, userID.(string), &req)
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": submission})
}

func (h *SubmissionHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	if err := h.service.Delete(id, userID.(string)); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Submission deleted"}})
}

func (h *SubmissionHandler) SubmitToJournal(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	var req models.SubmitToJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	submission, err := h.service.SubmitToJournal(id, userID.(string), req.JournalID)
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": submission})
}

func (h *SubmissionHandler) UpdateStatus(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	submission, err := h.service.UpdateStatus(id, userID.(string), req.Status, req.Note)
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": submission})
}

func (h *SubmissionHandler) AddHistory(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	var req models.AddHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	if err := h.service.AddHistory(id, userID.(string), &req); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "History added"}})
}
