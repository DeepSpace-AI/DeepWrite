package crossref

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/reference"
)

type Client struct {
	httpClient  *http.Client
	baseURL     string
	mailto      string
	userAgent   string
	rateLimit   time.Duration
	lastRequest time.Time
}

type ClientOption func(*Client)

func WithMailto(email string) ClientOption {
	return func(c *Client) {
		c.mailto = email
	}
}

func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithRateLimit(d time.Duration) ClientOption {
	return func(c *Client) {
		c.rateLimit = d
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.crossref.org",
		rateLimit: 50 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type crossrefWork struct {
	DOI             string           `json:"DOI"`
	Title           []string         `json:"title"`
	Author          []crossrefAuthor `json:"author"`
	PublishedPrint  *crossrefDate    `json:"published-print"`
	PublishedOnline *crossrefDate    `json:"published-online"`
	ContainerTitle  []string         `json:"container-title"`
	Type            string           `json:"type"`
	Abstract        string           `json:"abstract"`
	Volume          string           `json:"volume"`
	Issue           string           `json:"issue"`
	Page            string           `json:"page"`
	Publisher       string           `json:"publisher"`
	Language        string           `json:"language"`
	URL             string           `json:"URL"`
	ISBN            []string         `json:"ISBN"`
	Subject         []string         `json:"subject"`
}

type crossrefAuthor struct {
	Family      string `json:"family"`
	Given       string `json:"given"`
	Suffix      string `json:"suffix"`
	Name        string `json:"name"`
	ORCID       string `json:"ORCID"`
	Affiliation []struct {
		Name string `json:"name"`
	} `json:"affiliation"`
}

type crossrefDate struct {
	DateParts [][]int `json:"date-parts"`
}

type crossrefResponse struct {
	Status  string       `json:"status"`
	Message crossrefWork `json:"message"`
}

type crossrefListResponse struct {
	Status struct {
		Message string `json:"message"`
	} `json:"status"`
	Message struct {
		Items []crossrefWork `json:"items"`
		Total int            `json:"total-results"`
	} `json:"message"`
}

func (c *Client) WaitForRateLimit() {
	if c.rateLimit > 0 {
		elapsed := time.Since(c.lastRequest)
		if elapsed < c.rateLimit {
			time.Sleep(c.rateLimit - elapsed)
		}
	}
	c.lastRequest = time.Now()
}

func (c *Client) buildRequest(ctx context.Context, method, urlStr string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, urlStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	userAgent := "DeepWrite/1.0"
	if c.userAgent != "" {
		userAgent = c.userAgent
	}
	if c.mailto != "" {
		userAgent = fmt.Sprintf("%s (mailto:%s)", userAgent, c.mailto)
	}
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

func (c *Client) LookupByDOI(ctx context.Context, doi string) (*reference.Reference, error) {
	c.WaitForRateLimit()

	doi = strings.TrimSpace(doi)
	if doi == "" {
		return nil, fmt.Errorf("DOI is required")
	}

	urlStr := fmt.Sprintf("%s/works/%s", c.baseURL, url.PathEscape(doi))

	req, err := c.buildRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DOI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("DOI not found: %s", doi)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("crossref API error: %d - %s", resp.StatusCode, string(body))
	}

	var crResp crossrefResponse
	if err := json.NewDecoder(resp.Body).Decode(&crResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.workToReference(&crResp.Message), nil
}

func (c *Client) Search(ctx context.Context, query string, limit, offset int) ([]reference.Reference, int, error) {
	c.WaitForRateLimit()

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("rows", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))

	urlStr := fmt.Sprintf("%s/works?%s", c.baseURL, params.Encode())

	req, err := c.buildRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("crossref API error: %d - %s", resp.StatusCode, string(body))
	}

	var crResp crossrefListResponse
	if err := json.NewDecoder(resp.Body).Decode(&crResp); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	refs := make([]reference.Reference, len(crResp.Message.Items))
	for i, work := range crResp.Message.Items {
		refs[i] = *c.workToReference(&work)
	}

	return refs, crResp.Message.Total, nil
}

func (c *Client) workToReference(work *crossrefWork) *reference.Reference {
	ref := &reference.Reference{
		DOI:       work.DOI,
		Title:     c.getFirstOrEmpty(work.Title),
		Type:      c.mapType(work.Type),
		Volume:    work.Volume,
		Issue:     work.Issue,
		Pages:     work.Page,
		Publisher: work.Publisher,
		Language:  work.Language,
		URL:       work.URL,
		Abstract:  c.stripHTMLTags(work.Abstract),
	}

	if len(work.ContainerTitle) > 0 {
		ref.Source = work.ContainerTitle[0]
	}

	if len(work.ISBN) > 0 {
		ref.ISBN = work.ISBN[0]
	}

	ref.Authors, _ = json.Marshal(c.mapAuthors(work.Author))

	if work.PublishedPrint != nil && len(work.PublishedPrint.DateParts) > 0 && len(work.PublishedPrint.DateParts[0]) > 0 {
		year := work.PublishedPrint.DateParts[0][0]
		ref.Year = &year
	} else if work.PublishedOnline != nil && len(work.PublishedOnline.DateParts) > 0 && len(work.PublishedOnline.DateParts[0]) > 0 {
		year := work.PublishedOnline.DateParts[0][0]
		ref.Year = &year
	}

	if len(work.Subject) > 0 {
		keywords, _ := json.Marshal(work.Subject)
		ref.Keywords = keywords
	}

	return ref
}

func (c *Client) getFirstOrEmpty(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

func (c *Client) mapAuthors(authors []crossrefAuthor) []reference.Author {
	result := make([]reference.Author, len(authors))
	for i, a := range authors {
		author := reference.Author{
			Family:  a.Family,
			Given:   a.Given,
			Suffix:  a.Suffix,
			Literal: a.Name,
		}
		if a.ORCID != "" {
			author.ORCID = strings.TrimPrefix(a.ORCID, "https://orcid.org/")
		}
		result[i] = author
	}
	return result
}

func (c *Client) mapType(crType string) reference.ReferenceType {
	switch crType {
	case "journal-article", "article":
		return reference.ReferenceTypeArticle
	case "book":
		return reference.ReferenceTypeBook
	case "book-chapter":
		return reference.ReferenceTypeBookChapter
	case "proceedings-article", "conference-paper":
		return reference.ReferenceTypeConference
	case "dissertation":
		return reference.ReferenceTypeThesis
	case "report":
		return reference.ReferenceTypeReport
	case "posted-content", "preprint":
		return reference.ReferenceTypePreprint
	case "webpage":
		return reference.ReferenceTypeWeb
	default:
		return reference.ReferenceTypeUnknown
	}
}

func (c *Client) stripHTMLTags(s string) string {
	s = strings.ReplaceAll(s, "<jats:title>", "<strong>")
	s = strings.ReplaceAll(s, "</jats:title>", "</strong>")
	s = strings.ReplaceAll(s, "<jats:p>", "<p>")
	s = strings.ReplaceAll(s, "</jats:p>", "</p>")
	s = strings.ReplaceAll(s, "<jats:sec>", "")
	s = strings.ReplaceAll(s, "</jats:sec>", "")
	s = strings.ReplaceAll(s, "<jats:sup>", "<sup>")
	s = strings.ReplaceAll(s, "</jats:sup>", "</sup>")
	s = strings.ReplaceAll(s, "<jats:sub>", "<sub>")
	s = strings.ReplaceAll(s, "</jats:sub>", "</sub>")
	s = strings.ReplaceAll(s, "<jats:italic>", "<i>")
	s = strings.ReplaceAll(s, "</jats:italic>", "</i>")
	s = strings.ReplaceAll(s, "<jats:bold>", "<b>")
	s = strings.ReplaceAll(s, "</jats:bold>", "</b>")
	s = strings.ReplaceAll(s, "<jats:sc>", "<span style=\"font-variant:small-caps\">")
	s = strings.ReplaceAll(s, "</jats:sc>", "</span>")

	re := regexp.MustCompile(`<jats:\w+[^>]*>`)
	s = re.ReplaceAllString(s, "")
	re = regexp.MustCompile(`</jats:\w+>`)
	s = re.ReplaceAllString(s, "")

	return strings.TrimSpace(s)
}
