package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	baseAPIPath = "/wiki/api/v2"
	userAgent   = "confluence-reader/1.0"
)

// Client is a Confluence API client
type Client struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new Confluence API client
func NewClient(domain, email, apiToken string) *Client {
	return &Client{
		baseURL:  fmt.Sprintf("https://%s%s", domain, baseAPIPath),
		email:    email,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, queryParams url.Values) ([]byte, error) {
	reqURL := c.baseURL + path
	if len(queryParams) > 0 {
		reqURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Space represents a Confluence space
type Space struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Description *struct {
		Plain *struct {
			Value string `json:"value"`
		} `json:"plain"`
	} `json:"description"`
}

// SpaceResponse is the response for listing spaces
type SpaceResponse struct {
	Results []Space `json:"results"`
	Links   *struct {
		Next string `json:"next"`
	} `json:"_links"`
}

// GetSpaces retrieves all spaces
func (c *Client) GetSpaces() ([]Space, error) {
	var allSpaces []Space
	cursor := ""

	for {
		params := url.Values{}
		params.Set("limit", "100")
		if cursor != "" {
			params.Set("cursor", cursor)
		}

		body, err := c.doRequest("GET", "/spaces", params)
		if err != nil {
			return nil, fmt.Errorf("failed to get spaces: %w", err)
		}

		var response SpaceResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse spaces response: %w", err)
		}

		allSpaces = append(allSpaces, response.Results...)

		if response.Links == nil || response.Links.Next == "" {
			break
		}

		// Extract cursor from next URL
		nextURL, err := url.Parse(response.Links.Next)
		if err != nil {
			break
		}
		cursor = nextURL.Query().Get("cursor")
		if cursor == "" {
			break
		}
	}

	return allSpaces, nil
}

// Page represents a Confluence page
type Page struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Title    string `json:"title"`
	SpaceID  string `json:"spaceId"`
	ParentID string `json:"parentId"`
	Version  *struct {
		Number int    `json:"number"`
		When   string `json:"createdAt"`
	} `json:"version"`
	Body *struct {
		Storage *struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
}

// PageResponse is the response for listing pages
type PageResponse struct {
	Results []Page `json:"results"`
	Links   *struct {
		Next string `json:"next"`
	} `json:"_links"`
}

// GetSpacePages retrieves all pages in a space
func (c *Client) GetSpacePages(spaceID string) ([]Page, error) {
	var allPages []Page
	cursor := ""

	for {
		params := url.Values{}
		params.Set("limit", "100")
		if cursor != "" {
			params.Set("cursor", cursor)
		}

		path := fmt.Sprintf("/spaces/%s/pages", spaceID)
		body, err := c.doRequest("GET", path, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get pages for space %s: %w", spaceID, err)
		}

		var response PageResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse pages response: %w", err)
		}

		allPages = append(allPages, response.Results...)

		if response.Links == nil || response.Links.Next == "" {
			break
		}

		nextURL, err := url.Parse(response.Links.Next)
		if err != nil {
			break
		}
		cursor = nextURL.Query().Get("cursor")
		if cursor == "" {
			break
		}
	}

	return allPages, nil
}

// GetPage retrieves a single page with full content
func (c *Client) GetPage(pageID string) (*Page, error) {
	params := url.Values{}
	params.Set("body-format", "storage")

	path := fmt.Sprintf("/pages/%s", pageID)
	body, err := c.doRequest("GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get page %s: %w", pageID, err)
	}

	var page Page
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}

	return &page, nil
}

// Attachment represents a page attachment
type Attachment struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	MediaType string `json:"mediaType"`
	FileSize  int64  `json:"fileSize"`
	Download  *struct {
		URL string `json:"url"`
	} `json:"-"` // Handled by custom unmarshaler
	DownloadURL string `json:"-"` // Extracted URL
}

// UnmarshalJSON implements custom JSON unmarshaling for Attachment
// The Confluence API inconsistently returns downloadLink as either a string or an object
func (a *Attachment) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid recursion
	type Alias Attachment
	aux := &struct {
		DownloadLink json.RawMessage `json:"downloadLink"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try to parse downloadLink as a string first
	var urlString string
	if err := json.Unmarshal(aux.DownloadLink, &urlString); err == nil {
		a.DownloadURL = urlString
		return nil
	}

	// If that fails, try to parse as an object
	var urlObj struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(aux.DownloadLink, &urlObj); err == nil {
		a.DownloadURL = urlObj.URL
		return nil
	}

	// If both fail, leave DownloadURL empty (will be handled gracefully)
	return nil
}

// AttachmentResponse is the response for listing attachments
type AttachmentResponse struct {
	Results []Attachment `json:"results"`
	Links   *struct {
		Next string `json:"next"`
	} `json:"_links"`
}

// GetPageAttachments retrieves all attachments for a page
func (c *Client) GetPageAttachments(pageID string) ([]Attachment, error) {
	var allAttachments []Attachment
	cursor := ""

	for {
		params := url.Values{}
		params.Set("limit", "100")
		if cursor != "" {
			params.Set("cursor", cursor)
		}

		path := fmt.Sprintf("/pages/%s/attachments", pageID)
		body, err := c.doRequest("GET", path, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get attachments for page %s: %w", pageID, err)
		}

		var response AttachmentResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse attachments response: %w", err)
		}

		allAttachments = append(allAttachments, response.Results...)

		if response.Links == nil || response.Links.Next == "" {
			break
		}

		nextURL, err := url.Parse(response.Links.Next)
		if err != nil {
			break
		}
		cursor = nextURL.Query().Get("cursor")
		if cursor == "" {
			break
		}
	}

	return allAttachments, nil
}

// DownloadAttachment downloads an attachment to a writer
func (c *Client) DownloadAttachment(downloadURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
