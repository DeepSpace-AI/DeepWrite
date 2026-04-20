package semanticscholar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

type ClientOption func(*Client)

func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.semanticscholar.org/graph/v1",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type Paper struct {
	PaperID          string         `json:"paperId"`
	Title            string         `json:"title"`
	Abstract         string         `json:"abstract"`
	Year             int            `json:"year"`
	Authors          []Author       `json:"authors"`
	Venue            string         `json:"venue"`
	Journal          *Journal       `json:"journal"`
	DOI              string         `json:"doi"`
	URL              string         `json:"url"`
	CitationCount    int            `json:"citationCount"`
	ReferenceCount   int            `json:"referenceCount"`
	PublicationDate  string         `json:"publicationDate"`
	PublicationTypes []string       `json:"publicationTypes"`
	FieldsOfStudy    []string       `json:"fieldsOfStudy"`
	OpenAccessPDF    *OpenAccessPDF `json:"openAccessPdf"`
}

type Author struct {
	AuthorID string `json:"authorId"`
	Name     string `json:"name"`
}

type Journal struct {
	Name   string `json:"name"`
	Volume string `json:"volume"`
	Pages  string `json:"pages"`
}

type OpenAccessPDF struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

type SearchResponse struct {
	Total  int     `json:"total"`
	Offset int     `json:"offset"`
	Next   int     `json:"next"`
	Data   []Paper `json:"data"`
}

type PaperBulkResponse struct {
	Next string  `json:"next"`
	Data []Paper `json:"data"`
}

func (c *Client) Search(ctx context.Context, query string, limit, offset int) (*SearchResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("fields", "paperId,title,abstract,year,authors,venue,journal,doi,url,citationCount,referenceCount,publicationDate,publicationTypes,fieldsOfStudy,openAccessPdf")

	urlStr := fmt.Sprintf("%s/paper/search?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Semantic Scholar API error: %d - %s", resp.StatusCode, string(body))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetPaper(ctx context.Context, paperID string) (*Paper, error) {
	fields := "paperId,title,abstract,year,authors,venue,journal,doi,url,citationCount,referenceCount,publicationDate,publicationTypes,fieldsOfStudy,openAccessPdf"
	urlStr := fmt.Sprintf("%s/paper/%s?fields=%s", c.baseURL, url.PathEscape(paperID), fields)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get paper: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("paper not found: %s", paperID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Semantic Scholar API error: %d - %s", resp.StatusCode, string(body))
	}

	var paper Paper
	if err := json.NewDecoder(resp.Body).Decode(&paper); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &paper, nil
}

func (c *Client) GetPapersByDOI(ctx context.Context, dois []string) ([]Paper, error) {
	if len(dois) == 0 {
		return nil, nil
	}

	doiList := make([]string, len(dois))
	for i, doi := range dois {
		doiList[i] = fmt.Sprintf("DOI:%s", strings.TrimSpace(doi))
	}

	params := url.Values{}
	params.Set("fields", "paperId,title,abstract,year,authors,venue,journal,doi,url,citationCount,referenceCount,publicationDate,publicationTypes,fieldsOfStudy,openAccessPdf")

	urlStr := fmt.Sprintf("%s/paper/batch?%s", c.baseURL, params.Encode())

	body := map[string][]string{"ids": doiList}
	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get papers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Semantic Scholar API error: %d - %s", resp.StatusCode, string(body))
	}

	var papers []Paper
	if err := json.NewDecoder(resp.Body).Decode(&papers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return papers, nil
}

func (c *Client) SearchByAuthor(ctx context.Context, authorName string, limit int) ([]Paper, error) {
	if limit <= 0 {
		limit = 20
	}

	params := url.Values{}
	params.Set("query", authorName)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("fields", "paperId,title,abstract,year,authors,venue,journal,doi,url,citationCount,referenceCount,publicationDate,publicationTypes,fieldsOfStudy,openAccessPdf")

	urlStr := fmt.Sprintf("%s/paper/search/author?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search by author: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Semantic Scholar API error: %d - %s", resp.StatusCode, string(body))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}
