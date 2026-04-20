package handler

import (
	"encoding/json"
	"strings"

	"github.com/deepwrite/serivces/gateway/models/reference"
	"github.com/deepwrite/serivces/gateway/pkg/bibtex"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/deepwrite/serivces/gateway/pkg/ris"
	"github.com/gin-gonic/gin"
)

type ImportHandler struct{}

func NewImportHandler() *ImportHandler {
	return &ImportHandler{}
}

type BibtexImportRequest struct {
	Content string `json:"content"`
}

type BibtexImportResponse struct {
	Imported int                   `json:"imported"`
	Skipped  int                   `json:"skipped"`
	Errors   []ImportError         `json:"errors,omitempty"`
	Items    []reference.Reference `json:"items"`
}

type ImportError struct {
	Line    int    `json:"line,omitempty"`
	Key     string `json:"key,omitempty"`
	Message string `json:"message"`
}

func (h *ImportHandler) ImportBibtex(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req BibtexImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	if req.Content == "" {
		response.Failed(c, response.ErrorBadRequestCode, "content is required")
		return
	}

	parser := bibtex.NewParser()
	entries := parser.Parse(req.Content)

	result := BibtexImportResponse{
		Imported: 0,
		Skipped:  0,
		Errors:   make([]ImportError, 0),
		Items:    make([]reference.Reference, 0),
	}

	for _, entry := range entries {
		ref := parser.ToReference(entry)
		if ref.Title == "" {
			result.Skipped++
			result.Errors = append(result.Errors, ImportError{
				Key:     entry.Key,
				Message: "missing title",
			})
			continue
		}

		ref.OwnerID = userID

		if ref.DOI != "" {
			var existing reference.Reference
			if err := database.DB.Where("doi = ? AND owner_id = ?", ref.DOI, userID).First(&existing).Error; err == nil {
				result.Skipped++
				result.Errors = append(result.Errors, ImportError{
					Key:     entry.Key,
					Message: "duplicate DOI: " + ref.DOI,
				})
				continue
			}
		}

		if err := database.DB.Create(ref).Error; err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, ImportError{
				Key:     entry.Key,
				Message: "failed to create: " + err.Error(),
			})
			continue
		}

		result.Imported++
		result.Items = append(result.Items, *ref)
	}

	response.Success(c, response.SuccessCode, result)
}

type RISImportRequest struct {
	Content string `json:"content"`
}

type RISImportResponse struct {
	Imported int                   `json:"imported"`
	Skipped  int                   `json:"skipped"`
	Errors   []ImportError         `json:"errors,omitempty"`
	Items    []reference.Reference `json:"items"`
}

func (h *ImportHandler) ImportRIS(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req RISImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	if req.Content == "" {
		response.Failed(c, response.ErrorBadRequestCode, "content is required")
		return
	}

	parser := ris.NewRISParser()
	entries := parser.Parse(req.Content)

	result := RISImportResponse{
		Imported: 0,
		Skipped:  0,
		Errors:   make([]ImportError, 0),
		Items:    make([]reference.Reference, 0),
	}

	for i, entry := range entries {
		ref := parser.ToReference(entry)
		if ref.Title == "" {
			result.Skipped++
			result.Errors = append(result.Errors, ImportError{
				Line:    i + 1,
				Message: "missing title",
			})
			continue
		}

		ref.OwnerID = userID

		if ref.DOI != "" {
			var existing reference.Reference
			if err := database.DB.Where("doi = ? AND owner_id = ?", ref.DOI, userID).First(&existing).Error; err == nil {
				result.Skipped++
				result.Errors = append(result.Errors, ImportError{
					Line:    i + 1,
					Message: "duplicate DOI: " + ref.DOI,
				})
				continue
			}
		}

		if err := database.DB.Create(ref).Error; err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, ImportError{
				Line:    i + 1,
				Message: "failed to create: " + err.Error(),
			})
			continue
		}

		result.Imported++
		result.Items = append(result.Items, *ref)
	}

	response.Success(c, response.SuccessCode, result)
}

func (h *ImportHandler) ExportBibtex(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	ids := c.QueryArray("ids")

	var refs []reference.Reference
	query := database.DB.Where("owner_id = ?", userID)

	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}

	if err := query.Find(&refs).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to fetch references")
		return
	}

	exporter := bibtex.NewExporter()
	content := exporter.Export(refs)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=references.bib")
	c.String(200, content)
}

func (h *ImportHandler) ExportRIS(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	ids := c.QueryArray("ids")

	var refs []reference.Reference
	query := database.DB.Where("owner_id = ?", userID)

	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}

	if err := query.Find(&refs).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to fetch references")
		return
	}

	exporter := ris.NewRISExporter()
	content := exporter.Export(refs)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=references.ris")
	c.String(200, content)
}

func (h *ImportHandler) ExportCSLJSON(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	ids := c.QueryArray("ids")

	var refs []reference.Reference
	query := database.DB.Where("owner_id = ?", userID)

	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}

	if err := query.Find(&refs).Error; err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to fetch references")
		return
	}

	cslItems := make([]CSLItem, len(refs))
	for i, ref := range refs {
		cslItems[i] = referenceToCSL(&ref)
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=references.json")
	json.NewEncoder(c.Writer).Encode(cslItems)
}

type CSLItem struct {
	ID             string      `json:"id"`
	Type           string      `json:"type"`
	Title          string      `json:"title"`
	Author         []CSLAuthor `json:"author,omitempty"`
	Issued         CSLDate     `json:"issued,omitempty"`
	ContainerTitle string      `json:"container-title,omitempty"`
	Volume         string      `json:"volume,omitempty"`
	Issue          string      `json:"issue,omitempty"`
	Page           string      `json:"page,omitempty"`
	DOI            string      `json:"DOI,omitempty"`
	URL            string      `json:"URL,omitempty"`
	Publisher      string      `json:"publisher,omitempty"`
	Abstract       string      `json:"abstract,omitempty"`
	Note           string      `json:"note,omitempty"`
}

type CSLAuthor struct {
	Family  string `json:"family,omitempty"`
	Given   string `json:"given,omitempty"`
	Literal string `json:"literal,omitempty"`
}

type CSLDate struct {
	DateParts [][]int `json:"date-parts"`
}

func referenceToCSL(ref *reference.Reference) CSLItem {
	item := CSLItem{
		ID:             ref.ID,
		Type:           mapToCSLType(ref.Type),
		Title:          ref.Title,
		DOI:            ref.DOI,
		URL:            ref.URL,
		Abstract:       ref.Abstract,
		ContainerTitle: ref.Source,
		Volume:         ref.Volume,
		Issue:          ref.Issue,
		Page:           ref.Pages,
		Publisher:      ref.Publisher,
	}

	if len(ref.Authors) > 0 {
		var authors []reference.Author
		json.Unmarshal(ref.Authors, &authors)
		item.Author = make([]CSLAuthor, len(authors))
		for i, a := range authors {
			item.Author[i] = CSLAuthor{
				Family:  a.Family,
				Given:   a.Given,
				Literal: a.Literal,
			}
		}
	}

	if ref.Year != nil {
		item.Issued = CSLDate{
			DateParts: [][]int{{*ref.Year}},
		}
	}

	return item
}

func mapToCSLType(refType reference.ReferenceType) string {
	switch refType {
	case reference.ReferenceTypeArticle:
		return "article-journal"
	case reference.ReferenceTypeBook:
		return "book"
	case reference.ReferenceTypeBookChapter:
		return "chapter"
	case reference.ReferenceTypeConference:
		return "paper-conference"
	case reference.ReferenceTypeThesis:
		return "thesis"
	case reference.ReferenceTypeReport:
		return "report"
	case reference.ReferenceTypeWeb:
		return "webpage"
	case reference.ReferenceTypePreprint:
		return "article"
	default:
		return "document"
	}
}
