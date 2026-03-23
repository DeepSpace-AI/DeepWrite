package workspace

import (
	"context"
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreateAnnotationInput struct {
	FileID      string
	UserID      string
	WorkspaceID string
	Type        AnnotationType
	Page        int
	RectX       float64
	RectY       float64
	RectWidth   float64
	RectHeight  float64
	Color       string
	Content     string
	Paths       string
}

type UpdateAnnotationInput struct {
	RectX      *float64
	RectY      *float64
	RectWidth  *float64
	RectHeight *float64
	Color      *string
	Content    *string
	Paths      *string
}

func CreateAnnotation(ctx context.Context, input CreateAnnotationInput) (FileAnnotation, error) {
	annotation := FileAnnotation{
		FileID:      strings.TrimSpace(input.FileID),
		UserID:      strings.TrimSpace(input.UserID),
		WorkspaceID: strings.TrimSpace(input.WorkspaceID),
		Type:        input.Type,
		Page:        input.Page,
		RectX:       input.RectX,
		RectY:       input.RectY,
		RectWidth:   input.RectWidth,
		RectHeight:  input.RectHeight,
		Color:       strings.TrimSpace(input.Color),
		Content:     strings.TrimSpace(input.Content),
		Paths:       strings.TrimSpace(input.Paths),
	}

	err := database.DB.WithContext(ctx).Create(&annotation).Error
	return annotation, err
}

func GetAnnotationByID(ctx context.Context, annotationID string) (FileAnnotation, error) {
	var annotation FileAnnotation
	err := database.DB.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(annotationID)).
		First(&annotation).Error
	return annotation, err
}

func ListAnnotationsByFile(ctx context.Context, fileID string) ([]FileAnnotation, error) {
	var annotations []FileAnnotation
	err := database.DB.WithContext(ctx).
		Where("file_id = ?", strings.TrimSpace(fileID)).
		Order("page asc, rect_y asc").
		Find(&annotations).Error
	return annotations, err
}

func ListAnnotationsByFileAndPage(ctx context.Context, fileID string, page int) ([]FileAnnotation, error) {
	var annotations []FileAnnotation
	err := database.DB.WithContext(ctx).
		Where("file_id = ? AND page = ?", strings.TrimSpace(fileID), page).
		Order("rect_y asc").
		Find(&annotations).Error
	return annotations, err
}

func UpdateAnnotation(ctx context.Context, annotationID string, input UpdateAnnotationInput) (FileAnnotation, error) {
	var annotation FileAnnotation
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", strings.TrimSpace(annotationID)).
			First(&annotation).Error; err != nil {
			return err
		}

		updates := map[string]any{}

		if input.RectX != nil {
			updates["rect_x"] = *input.RectX
		}
		if input.RectY != nil {
			updates["rect_y"] = *input.RectY
		}
		if input.RectWidth != nil {
			updates["rect_width"] = *input.RectWidth
		}
		if input.RectHeight != nil {
			updates["rect_height"] = *input.RectHeight
		}
		if input.Color != nil {
			updates["color"] = strings.TrimSpace(*input.Color)
		}
		if input.Content != nil {
			updates["content"] = strings.TrimSpace(*input.Content)
		}
		if input.Paths != nil {
			updates["paths"] = strings.TrimSpace(*input.Paths)
		}

		if len(updates) > 0 {
			if err := tx.Model(&FileAnnotation{}).Where("id = ?", annotation.ID).Updates(updates).Error; err != nil {
				return err
			}
		}

		return tx.Where("id = ?", annotation.ID).First(&annotation).Error
	})

	return annotation, err
}

func DeleteAnnotation(ctx context.Context, annotationID string) error {
	return database.DB.WithContext(ctx).
		Delete(&FileAnnotation{}, "id = ?", strings.TrimSpace(annotationID)).Error
}

func DeleteAnnotationsByFile(ctx context.Context, fileID string) error {
	return database.DB.WithContext(ctx).
		Delete(&FileAnnotation{}, "file_id = ?", strings.TrimSpace(fileID)).Error
}
