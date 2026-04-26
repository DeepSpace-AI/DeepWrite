package handler

import (
	"net/http"

	"github.com/deepwrite/writing-service/internal/models"
	"github.com/deepwrite/writing-service/internal/service"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	docService *service.DocumentService
}

func NewDocumentHandler(docService *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{docService: docService}
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var req models.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	doc, err := h.docService.CreateDocument(c.Request.Context(), &req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")
	includeContent := c.Query("include_content") == "true"

	doc, err := h.docService.GetDocument(c.Request.Context(), id, includeContent)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	doc, err := h.docService.UpdateDocument(c.Request.Context(), id, &req, userID.(string))
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) UpdateContent(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateDocumentContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	content, err := h.docService.UpdateContent(c.Request.Context(), id, &req, userID.(string))
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	if err := h.docService.DeleteDocument(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	var req models.ListDocumentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	docs, total, err := h.docService.ListDocuments(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": docs,
		"total":     total,
		"page":      req.Page,
		"limit":     req.Limit,
	})
}

func (h *DocumentHandler) CreateVersion(c *gin.Context) {
	documentID := c.Param("id")
	var req models.CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	version, err := h.docService.CreateVersion(c.Request.Context(), documentID, &req, userID.(string))
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, version)
}

func (h *DocumentHandler) ListVersions(c *gin.Context) {
	documentID := c.Param("id")
	var req models.ListVersionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	versions, total, err := h.docService.ListVersions(c.Request.Context(), documentID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"versions": versions,
		"total":    total,
		"page":     req.Page,
		"limit":    req.Limit,
	})
}

func (h *DocumentHandler) AddCollaborator(c *gin.Context) {
	documentID := c.Param("id")
	var req models.AddCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.docService.AddCollaborator(c.Request.Context(), documentID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "collaborator added successfully"})
}

func (h *DocumentHandler) RemoveCollaborator(c *gin.Context) {
	documentID := c.Param("id")
	userID := c.Param("uid")

	if err := h.docService.RemoveCollaborator(c.Request.Context(), documentID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "collaborator removed successfully"})
}

func (h *DocumentHandler) GetCollaborators(c *gin.Context) {
	documentID := c.Param("id")
	collaborators, err := h.docService.GetCollaborators(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"collaborators": collaborators})
}
