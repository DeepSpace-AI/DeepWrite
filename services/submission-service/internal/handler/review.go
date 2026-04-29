package handler

import (
	"net/http"

	"github.com/deepwrite/submission-service/internal/models"
	"github.com/deepwrite/submission-service/internal/service"
	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	service *service.ReviewService
}

func NewReviewHandler(service *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userID")
	submissionID := c.Param("id")

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	review, err := h.service.Create(submissionID, userID.(string), &req)
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": review})
}

func (h *ReviewHandler) ListBySubmission(c *gin.Context) {
	userID, _ := c.Get("userID")
	submissionID := c.Param("id")

	reviews, err := h.service.ListBySubmission(submissionID, userID.(string))
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": reviews})
}

func (h *ReviewHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	reviewID := c.Param("review_id")

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	review, err := h.service.Update(reviewID, userID.(string), &req)
	if err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": review})
}

func (h *ReviewHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	reviewID := c.Param("review_id")

	if err := h.service.Delete(reviewID, userID.(string)); err != nil {
		if err == models.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "Access denied"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Review deleted"}})
}
