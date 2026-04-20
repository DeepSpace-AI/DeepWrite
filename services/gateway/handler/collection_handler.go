package handler

import (
	"strings"

	"github.com/deepwrite/serivces/gateway/models/reference"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CollectionHandler struct{}

func NewCollectionHandler() *CollectionHandler {
	return &CollectionHandler{}
}

func (h *CollectionHandler) List(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var params reference.ListCollectionsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid query parameters")
		return
	}

	query := database.DB.Model(&reference.Collection{}).Where("owner_id = ?", userID)

	if params.WorkspaceID != "" {
		query = query.Where("workspace_id = ?", params.WorkspaceID)
	}

	var collections []reference.Collection
	if err := query.Order("sort_order ASC, created_at ASC").Find(&collections).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to list collections")
		return
	}

	tree := h.buildTree(collections, nil)

	response.Success(c, response.SuccessCode, tree)
}

func (h *CollectionHandler) buildTree(collections []reference.Collection, parentID *string) []*reference.CollectionTree {
	var result []*reference.CollectionTree

	for i := range collections {
		c := &collections[i]

		var isMatch bool
		if parentID == nil {
			isMatch = c.ParentID == nil
		} else {
			isMatch = c.ParentID != nil && *c.ParentID == *parentID
		}

		if isMatch {
			node := &reference.CollectionTree{
				Collection: *c,
				Children:   h.buildTree(collections, &c.ID),
			}

			var count int64
			database.DB.Model(&reference.CollectionReference{}).
				Where("collection_id = ?", c.ID).
				Count(&count)
			node.ReferenceCount = int(count)

			result = append(result, node)
		}
	}

	return result
}

func (h *CollectionHandler) GetByID(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Failed(c, response.ErrorBadRequestCode, "id is required")
		return
	}

	var col reference.Collection
	if err := database.DB.Where("id = ? AND owner_id = ?", id, userID).First(&col).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Failed(c, response.ErrorNotFoundCode, "collection not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "failed to get collection")
		return
	}

	response.Success(c, response.SuccessCode, col)
}

func (h *CollectionHandler) Create(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var input reference.CreateCollectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	col := reference.Collection{
		OwnerID:     userID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		ParentID:    input.ParentID,
		Color:       input.Color,
		Icon:        input.Icon,
		WorkspaceID: input.WorkspaceID,
		SortOrder:   0,
	}

	if err := database.DB.Create(&col).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to create collection")
		return
	}

	response.Success(c, response.SuccessCreatedCode, col)
}

func (h *CollectionHandler) Update(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Failed(c, response.ErrorBadRequestCode, "id is required")
		return
	}

	var col reference.Collection
	if err := database.DB.Where("id = ? AND owner_id = ?", id, userID).First(&col).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Failed(c, response.ErrorNotFoundCode, "collection not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "failed to get collection")
		return
	}

	var input reference.UpdateCollectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	updates := make(map[string]any)

	if input.Name != nil {
		updates["name"] = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}
	if input.ParentID != nil {
		updates["parent_id"] = input.ParentID
	}
	if input.Color != nil {
		updates["color"] = *input.Color
	}
	if input.Icon != nil {
		updates["icon"] = *input.Icon
	}
	if input.SortOrder != nil {
		updates["sort_order"] = *input.SortOrder
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&col).Updates(updates).Error; err != nil {
			response.Failed(c, response.ErrorUnknownCode, "failed to update collection")
			return
		}
	}

	if err := database.DB.Where("id = ?", id).First(&col).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to get updated collection")
		return
	}

	response.Success(c, response.SuccessCode, col)
}

func (h *CollectionHandler) Delete(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Failed(c, response.ErrorBadRequestCode, "id is required")
		return
	}

	result := database.DB.Where("id = ? AND owner_id = ?", id, userID).Delete(&reference.Collection{})
	if result.Error != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to delete collection")
		return
	}

	if result.RowsAffected == 0 {
		response.Failed(c, response.ErrorNotFoundCode, "collection not found")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{"id": id})
}

func (h *CollectionHandler) AddReference(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	referenceID := strings.TrimSpace(c.Param("rid"))

	if collectionID == "" || referenceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "collection_id and reference_id are required")
		return
	}

	var col reference.Collection
	if err := database.DB.Where("id = ? AND owner_id = ?", collectionID, userID).First(&col).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Failed(c, response.ErrorNotFoundCode, "collection not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "failed to get collection")
		return
	}

	var ref reference.Reference
	if err := database.DB.Where("id = ? AND owner_id = ?", referenceID, userID).First(&ref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Failed(c, response.ErrorNotFoundCode, "reference not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "failed to get reference")
		return
	}

	cr := reference.CollectionReference{
		CollectionID: collectionID,
		ReferenceID:  referenceID,
	}

	if err := database.DB.FirstOrCreate(&cr, cr).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to add reference to collection")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"collection_id": collectionID,
		"reference_id":  referenceID,
	})
}

func (h *CollectionHandler) RemoveReference(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	referenceID := strings.TrimSpace(c.Param("rid"))

	if collectionID == "" || referenceID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "collection_id and reference_id are required")
		return
	}

	result := database.DB.Where("collection_id = ? AND reference_id = ?", collectionID, referenceID).
		Delete(&reference.CollectionReference{})

	if result.Error != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to remove reference from collection")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"collection_id": collectionID,
		"reference_id":  referenceID,
	})
}

func (h *CollectionHandler) SetReferences(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	if collectionID == "" {
		response.Failed(c, response.ErrorBadRequestCode, "collection_id is required")
		return
	}

	var col reference.Collection
	if err := database.DB.Where("id = ? AND owner_id = ?", collectionID, userID).First(&col).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Failed(c, response.ErrorNotFoundCode, "collection not found")
			return
		}
		response.Failed(c, response.ErrorUnknownCode, "failed to get collection")
		return
	}

	var input struct {
		ReferenceIDs []string `json:"reference_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("collection_id = ?", collectionID).Delete(&reference.CollectionReference{}).Error; err != nil {
			return err
		}

		for i, refID := range input.ReferenceIDs {
			cr := reference.CollectionReference{
				CollectionID: collectionID,
				ReferenceID:  refID,
				SortOrder:    i,
			}
			if err := tx.Create(&cr).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to set references")
		return
	}

	response.Success(c, response.SuccessCode, gin.H{
		"collection_id":   collectionID,
		"reference_count": len(input.ReferenceIDs),
	})
}
