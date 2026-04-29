package handler

import (
	"net/http"
	"strconv"

	"github.com/deepwrite/submission-service/internal/service"
	"github.com/gin-gonic/gin"
)

type JournalHandler struct {
	service *service.JournalService
}

func NewJournalHandler(service *service.JournalService) *JournalHandler {
	return &JournalHandler{service: service}
}

func (h *JournalHandler) List(c *gin.Context) {
	query := c.Query("query")
	category := c.Query("category")
	quartile := c.DefaultQuery("quartile", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	journals, total, err := h.service.List(query, category, quartile, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    journals,
		"meta":    gin.H{"page": page, "limit": limit, "total": total},
	})
}

func (h *JournalHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_ID", "message": "Invalid journal ID"}})
		return
	}

	journal, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "NOT_FOUND", "message": "Journal not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": journal})
}

func (h *JournalHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": categories})
}
