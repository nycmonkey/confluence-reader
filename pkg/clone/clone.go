package clone

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nycmonkey/confluence-reader/pkg/client"
)

// Cloner handles the cloning of Confluence content
type Cloner struct {
	client    *client.Client
	outputDir string
}

// NewCloner creates a new Cloner instance
func NewCloner(c *client.Client, outputDir string) *Cloner {
	return &Cloner{
		client:    c,
		outputDir: outputDir,
	}
}

// Clone performs the full clone operation
func (cl *Cloner) Clone() error {
	// Create output directory
	if err := os.MkdirAll(cl.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get all spaces
	fmt.Println("Fetching spaces...")
	spaces, err := cl.client.GetSpaces()
	if err != nil {
		return fmt.Errorf("failed to get spaces: %w", err)
	}

	fmt.Printf("Found %d space(s)\n", len(spaces))
	fmt.Println()

	// Clone each space
	for i, space := range spaces {
		fmt.Printf("[%d/%d] Processing space: %s (%s)\n", i+1, len(spaces), space.Name, space.Key)
		if err := cl.cloneSpace(space); err != nil {
			fmt.Printf("  Warning: Failed to clone space %s: %v\n", space.Key, err)
			continue
		}
	}

	return nil
}

// cloneSpace clones a single space
func (cl *Cloner) cloneSpace(space client.Space) error {
	// Create space directory
	spaceDir := filepath.Join(cl.outputDir, sanitizeFilename(space.Key))
	if err := os.MkdirAll(spaceDir, 0755); err != nil {
		return fmt.Errorf("failed to create space directory: %w", err)
	}

	// Save space metadata
	spaceMetadata := map[string]interface{}{
		"id":     space.ID,
		"key":    space.Key,
		"name":   space.Name,
		"type":   space.Type,
		"status": space.Status,
	}
	if space.Description != nil && space.Description.Plain != nil {
		spaceMetadata["description"] = space.Description.Plain.Value
	}

	metadataPath := filepath.Join(spaceDir, "space.json")
	if err := saveJSON(metadataPath, spaceMetadata); err != nil {
		return fmt.Errorf("failed to save space metadata: %w", err)
	}

	// Get all pages in space
	fmt.Printf("  Fetching pages...\n")
	pages, err := cl.client.GetSpacePages(space.ID)
	if err != nil {
		return fmt.Errorf("failed to get pages: %w", err)
	}

	fmt.Printf("  Found %d page(s)\n", len(pages))

	// Create pages directory
	pagesDir := filepath.Join(spaceDir, "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create pages directory: %w", err)
	}

	// Clone each page
	for j, page := range pages {
		fmt.Printf("  [%d/%d] Cloning page: %s\n", j+1, len(pages), page.Title)
		if err := cl.clonePage(page, pagesDir); err != nil {
			fmt.Printf("    Warning: Failed to clone page %s: %v\n", page.Title, err)
			continue
		}
	}

	return nil
}

// clonePage clones a single page
func (cl *Cloner) clonePage(page client.Page, pagesDir string) error {
	// Get full page content
	fullPage, err := cl.client.GetPage(page.ID)
	if err != nil {
		return fmt.Errorf("failed to get full page content: %w", err)
	}

	// Create page directory
	pageDirName := fmt.Sprintf("%s_%s", page.ID, sanitizeFilename(page.Title))
	pageDir := filepath.Join(pagesDir, pageDirName)
	if err := os.MkdirAll(pageDir, 0755); err != nil {
		return fmt.Errorf("failed to create page directory: %w", err)
	}

	// Save page metadata
	pageMetadata := map[string]interface{}{
		"id":       fullPage.ID,
		"title":    fullPage.Title,
		"status":   fullPage.Status,
		"spaceId":  fullPage.SpaceID,
		"parentId": fullPage.ParentID,
	}
	if fullPage.Version != nil {
		pageMetadata["version"] = map[string]interface{}{
			"number":    fullPage.Version.Number,
			"createdAt": fullPage.Version.When,
		}
	}

	metadataPath := filepath.Join(pageDir, "metadata.json")
	if err := saveJSON(metadataPath, pageMetadata); err != nil {
		return fmt.Errorf("failed to save page metadata: %w", err)
	}

	// Save page content (storage format)
	if fullPage.Body != nil && fullPage.Body.Storage != nil {
		contentPath := filepath.Join(pageDir, "content.html")
		if err := os.WriteFile(contentPath, []byte(fullPage.Body.Storage.Value), 0644); err != nil {
			return fmt.Errorf("failed to save page content: %w", err)
		}
	}

	// Get and save attachments
	attachments, err := cl.client.GetPageAttachments(page.ID)
	if err != nil {
		fmt.Printf("    Warning: Failed to get attachments: %v\n", err)
	} else if len(attachments) > 0 {
		fmt.Printf("    Found %d attachment(s)\n", len(attachments))
		attachmentsDir := filepath.Join(pageDir, "attachments")
		if err := os.MkdirAll(attachmentsDir, 0755); err != nil {
			return fmt.Errorf("failed to create attachments directory: %w", err)
		}

		for k, attachment := range attachments {
			fmt.Printf("    [%d/%d] Downloading: %s\n", k+1, len(attachments), attachment.Title)
			if err := cl.downloadAttachment(attachment, attachmentsDir); err != nil {
				fmt.Printf("      Warning: Failed to download attachment %s: %v\n", attachment.Title, err)
				continue
			}
		}
	}

	return nil
}

// downloadAttachment downloads and saves an attachment
func (cl *Cloner) downloadAttachment(attachment client.Attachment, attachmentsDir string) error {
	if attachment.Download == nil || attachment.Download.URL == "" {
		return fmt.Errorf("no download URL available")
	}

	data, err := cl.client.DownloadAttachment(attachment.Download.URL)
	if err != nil {
		return err
	}

	filename := sanitizeFilename(attachment.Title)
	filePath := filepath.Join(attachmentsDir, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save attachment: %w", err)
	}

	// Save attachment metadata
	metadataPath := filepath.Join(attachmentsDir, filename+".json")
	metadata := map[string]interface{}{
		"id":        attachment.ID,
		"title":     attachment.Title,
		"type":      attachment.Type,
		"mediaType": attachment.MediaType,
		"fileSize":  attachment.FileSize,
	}
	if err := saveJSON(metadataPath, metadata); err != nil {
		fmt.Printf("      Warning: Failed to save attachment metadata: %v\n", err)
	}

	return nil
}

// sanitizeFilename removes invalid characters from filenames
func sanitizeFilename(name string) string {
	// Replace invalid filename characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Limit length
	if len(result) > 200 {
		result = result[:200]
	}
	return strings.TrimSpace(result)
}

// saveJSON saves data as JSON to a file
func saveJSON(filepath string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, jsonData, 0644)
}
