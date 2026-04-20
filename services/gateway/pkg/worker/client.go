package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	authToken  string
}

type ClientOption func(*Client)

func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithAuthToken(token string) ClientOption {
	return func(c *Client) {
		c.authToken = token
	}
}

func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "http://localhost:8001",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type ExtractPdfRequest struct {
	FileID      string `json:"file_id"`
	ReferenceID string `json:"reference_id,omitempty"`
	Content     string `json:"content"`
}

type TaskResponse struct {
	TaskID   string `json:"task_id"`
	TaskName string `json:"task_name"`
	Queue    string `json:"queue"`
}

type TaskStatusResponse struct {
	TaskID string          `json:"task_id"`
	Status string          `json:"status"`
	Ready  bool            `json:"ready"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type PdfExtractResult struct {
	FileID         string            `json:"file_id"`
	Status         string            `json:"status"`
	ExtractedAt    string            `json:"extracted_at"`
	TextLength     int               `json:"text_length"`
	Metadata       map[string]any    `json:"metadata"`
	RawTextPreview string            `json:"raw_text_preview,omitempty"`
	DOI            string            `json:"doi,omitempty"`
	Crossref       map[string]any    `json:"crossref,omitempty"`
	Reference      *ReferenceFromPdf `json:"reference,omitempty"`
}

type ReferenceFromPdf struct {
	DOI       string      `json:"doi"`
	Title     string      `json:"title"`
	Authors   []PdfAuthor `json:"authors"`
	Year      int         `json:"year,omitempty"`
	Source    string      `json:"source,omitempty"`
	Type      string      `json:"type"`
	Abstract  string      `json:"abstract,omitempty"`
	Volume    string      `json:"volume,omitempty"`
	Issue     string      `json:"issue,omitempty"`
	Pages     string      `json:"pages,omitempty"`
	Publisher string      `json:"publisher,omitempty"`
	URL       string      `json:"url,omitempty"`
}

type PdfAuthor struct {
	Family  string `json:"family,omitempty"`
	Given   string `json:"given,omitempty"`
	Literal string `json:"literal,omitempty"`
	ORCID   string `json:"orcid,omitempty"`
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("worker API error: %d - %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) ExtractPdfMetadata(ctx context.Context, fileID, contentBase64 string) (*TaskResponse, error) {
	req := ExtractPdfRequest{
		FileID:  fileID,
		Content: contentBase64,
	}

	var result TaskResponse
	if err := c.doRequest(ctx, http.MethodPost, "/tasks/pdf/extract", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) ExtractAndLookupPdf(ctx context.Context, fileID, contentBase64 string) (*TaskResponse, error) {
	req := ExtractPdfRequest{
		FileID:  fileID,
		Content: contentBase64,
	}

	var result TaskResponse
	if err := c.doRequest(ctx, http.MethodPost, "/tasks/pdf/extract-and-lookup", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) ExtractAndLookupPdfWithRef(ctx context.Context, fileID, referenceID, contentBase64 string) (*TaskResponse, error) {
	req := ExtractPdfRequest{
		FileID:      fileID,
		ReferenceID: referenceID,
		Content:     contentBase64,
	}

	var result TaskResponse
	if err := c.doRequest(ctx, http.MethodPost, "/tasks/pdf/extract-and-lookup", req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatusResponse, error) {
	var result TaskStatusResponse
	if err := c.doRequest(ctx, http.MethodGet, "/tasks/"+taskID+"/status", nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) WaitForTask(ctx context.Context, taskID string, pollInterval time.Duration, maxWait time.Duration) (*TaskStatusResponse, error) {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	timeout := time.NewTimer(maxWait)
	defer timeout.Stop()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			return nil, fmt.Errorf("task %s timed out after %v", taskID, maxWait)
		case <-ticker.C:
			status, err := c.GetTaskStatus(ctx, taskID)
			if err != nil {
				return nil, err
			}

			if status.Ready {
				return status, nil
			}
		}
	}
}
