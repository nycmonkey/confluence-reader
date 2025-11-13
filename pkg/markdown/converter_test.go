package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBasicConversion(t *testing.T) {
	html := `<h1>Test Page</h1>
<p>This is a <strong>test</strong> with <em>formatting</em>.</p>
<ul>
<li>Item 1</li>
<li>Item 2</li>
</ul>`

	conv := NewConverter()
	markdown, err := conv.Convert(html)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if markdown == "" {
		t.Fatal("Expected non-empty markdown output")
	}

	t.Logf("Result:\n%s", markdown)
}

func TestRealConfluencePage(t *testing.T) {
	// Find a sample page
	samplePath := filepath.Join("..", "..", "test-sample", "ACACADEV", "pages")
	entries, err := os.ReadDir(samplePath)
	if err != nil {
		t.Skip("Sample data not available:", err)
	}

	if len(entries) == 0 {
		t.Skip("No sample pages found")
	}

	// Test first page
	testPage := filepath.Join(samplePath, entries[0].Name(), "content.html")
	htmlBytes, err := os.ReadFile(testPage)
	if err != nil {
		t.Fatalf("Failed to read sample page: %v", err)
	}

	conv := NewConverter()
	markdown, err := conv.Convert(string(htmlBytes))
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	t.Logf("Converted page: %s", entries[0].Name())
	t.Logf("HTML size: %d bytes", len(htmlBytes))
	t.Logf("Markdown size: %d bytes", len(markdown))
	t.Logf("First 500 chars:\n%s", markdown[:min(500, len(markdown))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestConfluenceCodeMacro(t *testing.T) {
	html := `<ac:structured-macro ac:name="code" ac:schema-version="1">
<ac:parameter ac:name="language">python</ac:parameter>
<ac:plain-text-body><![CDATA[def hello():
    print("Hello, World!")]]></ac:plain-text-body>
</ac:structured-macro>`

	conv := NewConverter()
	markdown, err := conv.Convert(html)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	expected := "```python"
	if !contains(markdown, expected) {
		t.Errorf("Expected markdown to contain '%s', got:\n%s", expected, markdown)
	}

	t.Logf("Result:\n%s", markdown)
}

func TestTOCRemoval(t *testing.T) {
	html := `<p><ac:structured-macro ac:name="toc" ac:schema-version="1" /></p>
<h2>Heading</h2>
<p>Content</p>`

	conv := NewConverter()
	markdown, err := conv.Convert(html)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Should not contain toc references
	if contains(markdown, "toc") {
		t.Errorf("Expected TOC to be removed, got:\n%s", markdown)
	}

	t.Logf("Result:\n%s", markdown)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestInternalLink(t *testing.T) {
	html := `<ac:link><ri:page ri:content-title="Another Page" /><ac:plain-text-link-body><![CDATA[Click here]]></ac:plain-text-link-body></ac:link>`

	conv := NewConverter()
	markdown, err := conv.Convert(html)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Should convert to Markdown link
	if !contains(markdown, "[") || !contains(markdown, "]") || !contains(markdown, ".md") {
		t.Errorf("Expected Markdown link format, got:\n%s", markdown)
	}

	t.Logf("Result:\n%s", markdown)
}

func TestConvertWithMetadata(t *testing.T) {
	html := `<h1>Test Page</h1>
<p>This is a test page with content.</p>`

	meta := PageMetadata{
		Title:     "Test Page",
		PageID:    "123456",
		SpaceKey:  "TEST",
		Version:   5,
		Author:    "user@example.com",
		ParentID:  "789",
		URL:       "https://company.atlassian.net/wiki/spaces/TEST/pages/123456",
	}

	conv := NewConverter()
	markdown, err := conv.ConvertWithMetadata(html, meta)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Check frontmatter present
	if !contains(markdown, "---") {
		t.Error("Expected frontmatter with --- delimiters")
	}

	if !contains(markdown, `title: "Test Page"`) {
		t.Error("Expected title in frontmatter")
	}

	if !contains(markdown, `confluence_id: "123456"`) {
		t.Error("Expected confluence_id in frontmatter")
	}

	if !contains(markdown, `space_key: "TEST"`) {
		t.Error("Expected space_key in frontmatter")
	}

	if !contains(markdown, `version: 5`) {
		t.Error("Expected version in frontmatter")
	}

	// Check content follows frontmatter
	if !contains(markdown, "# Test Page") {
		t.Error("Expected markdown content after frontmatter")
	}

	t.Logf("Result:\n%s", markdown)
}

func TestFrontmatterYAMLEscaping(t *testing.T) {
	html := `<h1>Test</h1>`

	// Title with quotes that need escaping
	meta := PageMetadata{
		Title:    `Test "Quoted" Page`,
		PageID:   "123",
		SpaceKey: "TEST",
		Version:  1,
	}

	conv := NewConverter()
	markdown, err := conv.ConvertWithMetadata(html, meta)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Should have escaped quotes
	if !contains(markdown, `Test \"Quoted\" Page`) {
		t.Errorf("Expected escaped quotes in YAML, got:\n%s", markdown)
	}

	t.Logf("Result:\n%s", markdown)
}

func TestEmoticonConversion(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "tick emoticon",
			html:     `<p>Test passed <ac:emoticon ac:name="tick" /></p>`,
			expected: "✅",
		},
		{
			name:     "cross emoticon",
			html:     `<p>Test failed <ac:emoticon ac:name="cross" /></p>`,
			expected: "❌",
		},
		{
			name:     "warning emoticon",
			html:     `<p><ac:emoticon ac:name="warning" /> Be careful!</p>`,
			expected: "⚠️",
		},
		{
			name:     "multiple emoticons",
			html:     `<p><ac:emoticon ac:name="tick" /> Success <ac:emoticon ac:name="cross" /> Failed</p>`,
			expected: "✅",
		},
		{
			name:     "unknown emoticon",
			html:     `<p><ac:emoticon ac:name="unknown-emoji" /></p>`,
			expected: ":unknown-emoji:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConverter()
			markdown, err := conv.Convert(tt.html)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if !contains(markdown, tt.expected) {
				t.Errorf("Expected markdown to contain '%s', got:\n%s", tt.expected, markdown)
			}

			t.Logf("Result:\n%s", markdown)
		})
	}
}
