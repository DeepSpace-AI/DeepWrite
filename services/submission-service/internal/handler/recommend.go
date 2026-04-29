package handler

import (
	"net/http"

	"github.com/deepwrite/submission-service/internal/models"
	"github.com/deepwrite/submission-service/internal/service"
	"github.com/gin-gonic/gin"
)

type RecommendHandler struct {
	journalService    *service.JournalService
	submissionService *service.SubmissionService
}

func NewRecommendHandler(journalService *service.JournalService, submissionService *service.SubmissionService) *RecommendHandler {
	return &RecommendHandler{journalService: journalService, submissionService: submissionService}
}

func (h *RecommendHandler) RecommendJournals(c *gin.Context) {
	var req models.RecommendJournalsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	journals, err := h.journalService.Recommend(req.Title, req.Abstract, req.Keywords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": journals})
}

func (h *RecommendHandler) RecommendForSubmission(c *gin.Context) {
	userID, _ := c.Get("userID")
	submissionID := c.Param("id")

	sub, err := h.submissionService.GetByID(submissionID, userID.(string))
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "NOT_FOUND", "message": "Submission not found"}})
		return
	}

	abstract := ""
	if sub.Abstract.Valid {
		abstract = sub.Abstract.String
	}

	journals, err := h.journalService.Recommend(sub.Title, abstract, sub.Keywords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": journals})
}
