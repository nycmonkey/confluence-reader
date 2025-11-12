package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("example.atlassian.net", "user@example.com", "test-token")

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	expected := "https://example.atlassian.net/wiki/api/v2"
	if client.baseURL != expected {
		t.Errorf("Expected baseURL %s, got %s", expected, client.baseURL)
	}

	if client.email != "user@example.com" {
		t.Errorf("Expected email user@example.com, got %s", client.email)
	}

	if client.apiToken != "test-token" {
		t.Errorf("Expected apiToken test-token, got %s", client.apiToken)
	}
}

func TestGetSpaces(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication
		user, pass, ok := r.BasicAuth()
		if !ok || user != "user@example.com" || pass != "test-token" {
			t.Error("Expected basic auth credentials")
		}

		// Verify path
		if r.URL.Path != "/spaces" {
			t.Errorf("Expected path /spaces, got %s", r.URL.Path)
		}

		// Return mock response
		response := SpaceResponse{
			Results: []Space{
				{
					ID:     "123",
					Key:    "TEST",
					Name:   "Test Space",
					Type:   "global",
					Status: "current",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := &Client{
		baseURL:    server.URL,
		email:      "user@example.com",
		apiToken:   "test-token",
		httpClient: http.DefaultClient,
	}

	spaces, err := client.GetSpaces()
	if err != nil {
		t.Fatalf("GetSpaces failed: %v", err)
	}

	if len(spaces) != 1 {
		t.Fatalf("Expected 1 space, got %d", len(spaces))
	}

	if spaces[0].Key != "TEST" {
		t.Errorf("Expected space key TEST, got %s", spaces[0].Key)
	}
}

func TestGetSpacePages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path
		if r.URL.Path != "/spaces/123/pages" {
			t.Errorf("Expected path /spaces/123/pages, got %s", r.URL.Path)
		}

		// Return mock response
		response := PageResponse{
			Results: []Page{
				{
					ID:      "456",
					Title:   "Test Page",
					Status:  "current",
					SpaceID: "123",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		email:      "user@example.com",
		apiToken:   "test-token",
		httpClient: http.DefaultClient,
	}

	pages, err := client.GetSpacePages("123")
	if err != nil {
		t.Fatalf("GetSpacePages failed: %v", err)
	}

	if len(pages) != 1 {
		t.Fatalf("Expected 1 page, got %d", len(pages))
	}

	if pages[0].Title != "Test Page" {
		t.Errorf("Expected page title 'Test Page', got %s", pages[0].Title)
	}
}

func TestGetPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path
		if r.URL.Path != "/pages/456" {
			t.Errorf("Expected path /pages/456, got %s", r.URL.Path)
		}

		// Verify body-format parameter
		if r.URL.Query().Get("body-format") != "storage" {
			t.Error("Expected body-format=storage parameter")
		}

		// Return mock response
		page := Page{
			ID:      "456",
			Title:   "Test Page",
			Status:  "current",
			SpaceID: "123",
			Body: &struct {
				Storage *struct {
					Value          string `json:"value"`
					Representation string `json:"representation"`
				} `json:"storage"`
			}{
				Storage: &struct {
					Value          string `json:"value"`
					Representation string `json:"representation"`
				}{
					Value:          "<p>Test content</p>",
					Representation: "storage",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		email:      "user@example.com",
		apiToken:   "test-token",
		httpClient: http.DefaultClient,
	}

	page, err := client.GetPage("456")
	if err != nil {
		t.Fatalf("GetPage failed: %v", err)
	}

	if page.Title != "Test Page" {
		t.Errorf("Expected page title 'Test Page', got %s", page.Title)
	}

	if page.Body == nil || page.Body.Storage == nil {
		t.Fatal("Expected page body storage to be present")
	}

	if page.Body.Storage.Value != "<p>Test content</p>" {
		t.Errorf("Expected body content '<p>Test content</p>', got %s", page.Body.Storage.Value)
	}
}

func TestGetPageAttachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path
		if r.URL.Path != "/pages/456/attachments" {
			t.Errorf("Expected path /pages/456/attachments, got %s", r.URL.Path)
		}

		// Return mock response
		response := AttachmentResponse{
			Results: []Attachment{
				{
					ID:        "789",
					Title:     "test.pdf",
					Type:      "attachment",
					MediaType: "application/pdf",
					FileSize:  12345,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		email:      "user@example.com",
		apiToken:   "test-token",
		httpClient: http.DefaultClient,
	}

	attachments, err := client.GetPageAttachments("456")
	if err != nil {
		t.Fatalf("GetPageAttachments failed: %v", err)
	}

	if len(attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(attachments))
	}

	if attachments[0].Title != "test.pdf" {
		t.Errorf("Expected attachment title 'test.pdf', got %s", attachments[0].Title)
	}
}

func TestPagination(t *testing.T) {
	callCount := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		var response SpaceResponse

		if callCount == 1 {
			// First page
			response = SpaceResponse{
				Results: []Space{
					{ID: "1", Key: "SPACE1", Name: "Space 1"},
				},
				Links: &struct {
					Next string `json:"next"`
				}{
					Next: serverURL + "/spaces?cursor=abc123",
				},
			}
		} else {
			// Second page (no more results)
			response = SpaceResponse{
				Results: []Space{
					{ID: "2", Key: "SPACE2", Name: "Space 2"},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	serverURL = server.URL

	client := &Client{
		baseURL:    server.URL,
		email:      "user@example.com",
		apiToken:   "test-token",
		httpClient: http.DefaultClient,
	}

	spaces, err := client.GetSpaces()
	if err != nil {
		t.Fatalf("GetSpaces failed: %v", err)
	}

	if len(spaces) != 2 {
		t.Fatalf("Expected 2 spaces from pagination, got %d", len(spaces))
	}

	if callCount != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", callCount)
	}
}
