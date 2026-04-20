package handler

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/deepwrite/serivces/gateway/models/reference"
	"github.com/deepwrite/serivces/gateway/pkg/crossref"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/response"
	"github.com/deepwrite/serivces/gateway/pkg/semanticscholar"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	crossrefClient        *crossref.Client
	semanticscholarClient *semanticscholar.Client
}

func NewSearchHandler() *SearchHandler {
	ssApiKey := os.Getenv("SEMANTIC_SCHOLAR_API_KEY")

	return &SearchHandler{
		crossrefClient: crossref.NewClient(
			crossref.WithMailto("support@deepwrite.work"),
		),
		semanticscholarClient: semanticscholar.NewClient(
			semanticscholar.WithAPIKey(ssApiKey),
		),
	}
}

type AISearchResult struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Authors       []Author `json:"authors"`
	Year          int      `json:"year,omitempty"`
	Source        string   `json:"source,omitempty"`
	DOI           string   `json:"doi,omitempty"`
	URL           string   `json:"url,omitempty"`
	Abstract      string   `json:"abstract,omitempty"`
	Type          string   `json:"type"`
	CitationCount int      `json:"citation_count,omitempty"`
	OpenAccessURL string   `json:"open_access_url,omitempty"`
	Provider      string   `json:"provider"`
}

type Author struct {
	Family  string `json:"family,omitempty"`
	Given   string `json:"given,omitempty"`
	Literal string `json:"literal,omitempty"`
	ORCID   string `json:"orcid,omitempty"`
}

type AISearchResponse struct {
	Items    []AISearchResult `json:"items"`
	Total    int              `json:"total"`
	Provider string           `json:"provider"`
}

func (h *SearchHandler) Search(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		response.Failed(c, response.ErrorBadRequestCode, "query is required")
		return
	}

	provider := strings.ToLower(strings.TrimSpace(c.Query("provider")))
	if provider == "" {
		provider = "all"
	}

	limit := 20
	offset := 0

	var items []AISearchResult
	var total int

	switch provider {
	case "semanticscholar", "ss":
		items, total = h.searchSemanticScholar(c.Request.Context(), query, limit, offset)
	case "crossref", "cr":
		items, total = h.searchCrossref(c.Request.Context(), query, limit, offset)
	default:
		items, total = h.searchAll(c.Request.Context(), query, limit, offset)
	}

	response.Success(c, response.SuccessCode, AISearchResponse{
		Items:    items,
		Total:    total,
		Provider: provider,
	})
}

func (h *SearchHandler) searchAll(ctx interface{}, query string, limit, offset int) ([]AISearchResult, int) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	ssResults := make([]AISearchResult, 0)
	crResults := make([]AISearchResult, 0)

	wg.Add(2)

	go func() {
		defer wg.Done()
		results, _ := h.searchSemanticScholar(ctx, query, limit/2, offset)
		mu.Lock()
		ssResults = results
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		results, _ := h.searchCrossref(ctx, query, limit/2, offset)
		mu.Lock()
		crResults = results
		mu.Unlock()
	}()

	wg.Wait()

	seen := make(map[string]bool)
	merged := make([]AISearchResult, 0)

	for _, item := range ssResults {
		key := item.DOI
		if key == "" {
			key = item.Title
		}
		if !seen[key] {
			seen[key] = true
			merged = append(merged, item)
		}
	}

	for _, item := range crResults {
		key := item.DOI
		if key == "" {
			key = item.Title
		}
		if !seen[key] {
			seen[key] = true
			merged = append(merged, item)
		}
	}

	return merged, len(merged)
}

func (h *SearchHandler) searchSemanticScholar(ctx interface{}, query string, limit, offset int) ([]AISearchResult, int) {
	var reqCtx context.Context
	switch v := ctx.(type) {
	case context.Context:
		reqCtx = v
	default:
		reqCtx = context.Background()
	}

	result, err := h.semanticscholarClient.Search(reqCtx, query, limit, offset)
	if err != nil {
		return nil, 0
	}

	items := make([]AISearchResult, len(result.Data))
	for i, paper := range result.Data {
		authors := make([]Author, len(paper.Authors))
		for j, a := range paper.Authors {
			parts := strings.SplitN(a.Name, " ", 2)
			authors[j] = Author{
				Literal: a.Name,
			}
			if len(parts) == 2 {
				authors[j].Given = parts[0]
				authors[j].Family = parts[1]
			}
		}

		source := paper.Venue
		if paper.Journal != nil && paper.Journal.Name != "" {
			source = paper.Journal.Name
		}

		paperType := "article"
		if len(paper.PublicationTypes) > 0 {
			pt := paper.PublicationTypes[0]
			switch pt {
			case "Conference":
				paperType = "conference"
			case "Book":
				paperType = "book"
			case "Review":
				paperType = "article"
			case "Thesis":
				paperType = "thesis"
			}
		}

		openAccessURL := ""
		if paper.OpenAccessPDF != nil {
			openAccessURL = paper.OpenAccessPDF.URL
		}

		items[i] = AISearchResult{
			ID:            paper.PaperID,
			Title:         paper.Title,
			Authors:       authors,
			Year:          paper.Year,
			Source:        source,
			DOI:           paper.DOI,
			URL:           paper.URL,
			Abstract:      paper.Abstract,
			Type:          paperType,
			CitationCount: paper.CitationCount,
			OpenAccessURL: openAccessURL,
			Provider:      "semanticscholar",
		}
	}

	return items, result.Total
}

func (h *SearchHandler) searchCrossref(ctx interface{}, query string, limit, offset int) ([]AISearchResult, int) {
	var reqCtx context.Context
	switch v := ctx.(type) {
	case context.Context:
		reqCtx = v
	default:
		reqCtx = context.Background()
	}

	refs, total, err := h.crossrefClient.Search(reqCtx, query, limit, offset)
	if err != nil {
		return nil, 0
	}

	items := make([]AISearchResult, len(refs))
	for i, ref := range refs {
		var authors []Author
		if len(ref.Authors) > 0 {
			var authorList []reference.Author
			json.Unmarshal(ref.Authors, &authorList)
			authors = make([]Author, len(authorList))
			for j, a := range authorList {
				authors[j] = Author{
					Family:  a.Family,
					Given:   a.Given,
					Literal: a.Literal,
					ORCID:   a.ORCID,
				}
			}
		}

		year := 0
		if ref.Year != nil {
			year = *ref.Year
		}

		items[i] = AISearchResult{
			ID:       ref.DOI,
			Title:    ref.Title,
			Authors:  authors,
			Year:     year,
			Source:   ref.Source,
			DOI:      ref.DOI,
			URL:      ref.URL,
			Abstract: ref.Abstract,
			Type:     string(ref.Type),
			Provider: "crossref",
		}
	}

	return items, total
}

type BatchImportRequest struct {
	Items []AISearchResult `json:"items"`
}

type BatchImportResponse struct {
	Imported int                   `json:"imported"`
	Failed   int                   `json:"failed"`
	Items    []reference.Reference `json:"items"`
	Errors   []string              `json:"errors,omitempty"`
}

func (h *SearchHandler) BatchImport(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req BatchImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	if len(req.Items) == 0 {
		response.Failed(c, response.ErrorBadRequestCode, "items are required")
		return
	}

	result := BatchImportResponse{
		Imported: 0,
		Failed:   0,
		Items:    make([]reference.Reference, 0),
		Errors:   make([]string, 0),
	}

	for _, item := range req.Items {
		ref := &reference.Reference{
			OwnerID:  userID,
			Title:    item.Title,
			DOI:      item.DOI,
			URL:      item.URL,
			Abstract: item.Abstract,
			Source:   item.Source,
			Type:     reference.ReferenceType(item.Type),
		}

		if item.Year > 0 {
			ref.Year = &item.Year
		}

		if len(item.Authors) > 0 {
			authors := make([]reference.Author, len(item.Authors))
			for i, a := range item.Authors {
				authors[i] = reference.Author{
					Family:  a.Family,
					Given:   a.Given,
					Literal: a.Literal,
					ORCID:   a.ORCID,
				}
			}
			authorsJSON, _ := json.Marshal(authors)
			ref.Authors = authorsJSON
		}

		if ref.DOI != "" {
			var existing reference.Reference
			if err := database.DB.Where("doi = ? AND owner_id = ?", ref.DOI, userID).First(&existing).Error; err == nil {
				result.Failed++
				result.Errors = append(result.Errors, "duplicate DOI: "+ref.DOI)
				continue
			}
		}

		if err := database.DB.Create(ref).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "failed to create: "+ref.Title)
			continue
		}

		result.Imported++
		result.Items = append(result.Items, *ref)
	}

	response.Success(c, response.SuccessCode, result)
}

type SearchByDOIRequest struct {
	DOIs []string `json:"dois"`
}

func (h *SearchHandler) SearchByDOI(c *gin.Context) {
	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		response.Failed(c, response.ErrorUnauthorizedCode, "unauthorized")
		return
	}

	var req SearchByDOIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failed(c, response.ErrorBadRequestCode, "invalid request: "+err.Error())
		return
	}

	if len(req.DOIs) == 0 {
		response.Failed(c, response.ErrorBadRequestCode, "dois are required")
		return
	}

	papers, err := h.semanticscholarClient.GetPapersByDOI(c.Request.Context(), req.DOIs)
	if err != nil {
		response.Failed(c, response.ErrorUnknownCode, "failed to search by DOI: "+err.Error())
		return
	}

	items := make([]AISearchResult, len(papers))
	for i, paper := range papers {
		authors := make([]Author, len(paper.Authors))
		for j, a := range paper.Authors {
			parts := strings.SplitN(a.Name, " ", 2)
			authors[j] = Author{
				Literal: a.Name,
			}
			if len(parts) == 2 {
				authors[j].Given = parts[0]
				authors[j].Family = parts[1]
			}
		}

		source := paper.Venue
		if paper.Journal != nil && paper.Journal.Name != "" {
			source = paper.Journal.Name
		}

		openAccessURL := ""
		if paper.OpenAccessPDF != nil {
			openAccessURL = paper.OpenAccessPDF.URL
		}

		items[i] = AISearchResult{
			ID:            paper.PaperID,
			Title:         paper.Title,
			Authors:       authors,
			Year:          paper.Year,
			Source:        source,
			DOI:           paper.DOI,
			URL:           paper.URL,
			Abstract:      paper.Abstract,
			Type:          "article",
			CitationCount: paper.CitationCount,
			OpenAccessURL: openAccessURL,
			Provider:      "semanticscholar",
		}
	}

	response.Success(c, response.SuccessCode, AISearchResponse{
		Items:    items,
		Total:    len(items),
		Provider: "semanticscholar",
	})
}
