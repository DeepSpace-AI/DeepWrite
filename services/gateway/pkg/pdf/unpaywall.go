package pdf

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

type UnpaywallClient struct {
	httpClient *http.Client
	email      string
	baseURL    string
}

func NewUnpaywallClient(email string) *UnpaywallClient {
	return &UnpaywallClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		email:   email,
		baseURL: "https://api.unpaywall.org/v2",
	}
}

type UnpaywallResponse struct {
	DOI               string       `json:"doi"`
	Title             string       `json:"title"`
	Year              int          `json:"year"`
	Genre             string       `json:"genre"`
	IsOA              bool         `json:"is_oa"`
	IsParadox         bool         `json:"is_paradox"`
	OAStatus          string       `json:"oa_status"`
	HasRepositoryCopy bool         `json:"has_repository_copy"`
	BestOA            *OALocation  `json:"best_oa_location"`
	FirstOA           *OALocation  `json:"first_oa_location"`
	OALocations       []OALocation `json:"oa_locations"`
	OALocationsEmb    []OALocation `json:"oa_locations_embargoed"`
	Updated           string       `json:"updated"`
}

type OALocation struct {
	EndpointID            string `json:"endpoint_id"`
	Evidence              string `json:"evidence"`
	HostType              string `json:"host_type"`
	IsBest                bool   `json:"is_best"`
	License               string `json:"license"`
	LicenseID             string `json:"license_id"`
	Version               string `json:"version"`
	URL                   string `json:"url"`
	URLForPDF             string `json:"url_for_pdf"`
	URLForLandingPage     string `json:"url_for_landing_page"`
	PMHID                 string `json:"pmh_id"`
	RepositoryInstitution string `json:"repository_institution"`
	Updated               string `json:"updated"`
}

func (c *UnpaywallClient) LookupDOI(ctx context.Context, doi string) (*UnpaywallResponse, error) {
	doi = strings.TrimSpace(doi)
	if doi == "" {
		return nil, fmt.Errorf("DOI is required")
	}

	urlStr := fmt.Sprintf("%s/%s?email=%s", c.baseURL, url.PathEscape(doi), url.QueryEscape(c.email))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup DOI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("DOI not found in Unpaywall: %s", doi)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Unpaywall API error: %d - %s", resp.StatusCode, string(body))
	}

	var result UnpaywallResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *UnpaywallClient) GetPDFURL(ctx context.Context, doi string) (string, error) {
	result, err := c.LookupDOI(ctx, doi)
	if err != nil {
		return "", err
	}

	if result.BestOA != nil && result.BestOA.URLForPDF != "" {
		return result.BestOA.URLForPDF, nil
	}

	if result.BestOA != nil && result.BestOA.URL != "" {
		return result.BestOA.URL, nil
	}

	for _, loc := range result.OALocations {
		if loc.URLForPDF != "" {
			return loc.URLForPDF, nil
		}
	}

	for _, loc := range result.OALocations {
		if loc.URL != "" {
			return loc.URL, nil
		}
	}

	return "", fmt.Errorf("no open access PDF found for DOI: %s", doi)
}

func (c *UnpaywallClient) DownloadPDF(ctx context.Context, pdfURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pdfURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "DeepWrite/1.0 (mailto:"+c.email+")")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download PDF: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/pdf") && !strings.Contains(contentType, "application/octet-stream") {
		return nil, fmt.Errorf("URL does not return a PDF (Content-Type: %s)", contentType)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}

	return data, nil
}

func (c *UnpaywallClient) DownloadPDFByDOI(ctx context.Context, doi string) ([]byte, string, error) {
	pdfURL, err := c.GetPDFURL(ctx, doi)
	if err != nil {
		return nil, "", err
	}

	data, err := c.DownloadPDF(ctx, pdfURL)
	if err != nil {
		return nil, pdfURL, err
	}

	return data, pdfURL, nil
}
