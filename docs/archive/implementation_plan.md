# Implementation Plan - Markdown Export

## Current Status

**Completed**: Phase 1 (Research & Analysis)  
**Next**: Phase 2 (Integration with Clone Pipeline)

---

## Phase 2: Integration (NEXT SESSION)

### Goal
Integrate the markdown converter into the existing clone pipeline so users can export Markdown alongside HTML.

### Step 2.1: Add Environment Variable Support

**File**: `main.go`

**Changes**:
```go
// After other env var checks, add:
exportMarkdown := os.Getenv("CONFLUENCE_EXPORT_MARKDOWN")
if exportMarkdown == "true" {
    fmt.Println("Markdown export enabled")
}
```

**Tests**:
- Run with `CONFLUENCE_EXPORT_MARKDOWN=true` and verify message
- Run without and verify no message

**Time**: 15 minutes

---

### Step 2.2: Add Frontmatter Support

**File**: `pkg/markdown/converter.go`

**New Type**:
```go
type PageMetadata struct {
    Title      string
    PageID     string
    SpaceKey   string
    Version    int
    UpdatedAt  time.Time
    Author     string
    ParentID   string
    URL        string
}
```

**New Method**:
```go
func (c *Converter) ConvertWithMetadata(html string, meta PageMetadata) (string, error) {
    // Generate YAML frontmatter
    frontmatter := generateFrontmatter(meta)
    
    // Convert HTML to markdown
    markdown, err := c.Convert(html)
    if err != nil {
        return "", err
    }
    
    // Combine frontmatter + markdown
    return frontmatter + "\n\n" + markdown, nil
}

func generateFrontmatter(meta PageMetadata) string {
    return fmt.Sprintf(`---
title: "%s"
confluence_id: "%s"
space_key: "%s"
version: %d
last_updated: "%s"
author: "%s"
parent_id: "%s"
url: "%s"
---`, 
        escapeYAML(meta.Title),
        meta.PageID,
        meta.SpaceKey,
        meta.Version,
        meta.UpdatedAt.Format(time.RFC3339),
        meta.Author,
        meta.ParentID,
        meta.URL,
    )
}

func escapeYAML(s string) string {
    // Escape quotes in YAML strings
    return strings.ReplaceAll(s, `"`, `\"`)
}
```

**Tests**:
```go
func TestConvertWithMetadata(t *testing.T) {
    html := "<h1>Test</h1><p>Content</p>"
    meta := PageMetadata{
        Title:     "Test Page",
        PageID:    "123456",
        SpaceKey:  "TEST",
        Version:   5,
        UpdatedAt: time.Now(),
        Author:    "user@example.com",
        ParentID:  "789",
        URL:       "https://company.atlassian.net/wiki/spaces/TEST/pages/123456",
    }
    
    conv := NewConverter()
    md, err := conv.ConvertWithMetadata(html, meta)
    
    if err != nil {
        t.Fatalf("Conversion failed: %v", err)
    }
    
    // Check frontmatter
    if !strings.HasPrefix(md, "---\n") {
        t.Error("Missing frontmatter")
    }
    
    if !strings.Contains(md, "title: \"Test Page\"") {
        t.Error("Missing title in frontmatter")
    }
    
    // Check content follows frontmatter
    if !strings.Contains(md, "# Test") {
        t.Error("Missing markdown content")
    }
}
```

**Time**: 1 hour

---

### Step 2.3: Modify Cloner to Support Markdown Export

**File**: `pkg/clone/clone.go`

**Add Fields**:
```go
type Cloner struct {
    client         *client.Client
    outputDir      string
    exportMarkdown bool               // NEW
    converter      *markdown.Converter // NEW
}
```

**Add Method**:
```go
// EnableMarkdownExport enables markdown export alongside HTML
func (cl *Cloner) EnableMarkdownExport() {
    cl.exportMarkdown = true
    cl.converter = markdown.NewConverter()
}
```

**Modify clonePage**:
```go
func (cl *Cloner) clonePage(page client.Page, pagesDir string) error {
    // ... existing code to save HTML ...
    
    // NEW: Export markdown if enabled
    if cl.exportMarkdown && fullPage.Body != nil && fullPage.Body.Storage != nil {
        md, err := cl.convertPageToMarkdown(fullPage)
        if err != nil {
            fmt.Printf("    Warning: Failed to convert to markdown: %v\n", err)
        } else {
            mdPath := filepath.Join(pageDir, "content.md")
            if err := os.WriteFile(mdPath, []byte(md), 0644); err != nil {
                fmt.Printf("    Warning: Failed to save markdown: %v\n", err)
            }
        }
    }
    
    return nil
}

// convertPageToMarkdown converts a page to markdown with frontmatter
func (cl *Cloner) convertPageToMarkdown(page client.Page) (string, error) {
    // Extract metadata
    meta := markdown.PageMetadata{
        Title:     page.Title,
        PageID:    page.ID,
        SpaceKey:  page.SpaceKey, // TODO: Extract from page
        Version:   page.Version.Number,
        UpdatedAt: page.Version.CreatedAt,
        Author:    page.Version.AuthoredBy.Email,
        ParentID:  page.ParentID,
        URL:       fmt.Sprintf("https://%s/wiki/spaces/%s/pages/%s", cl.domain, page.SpaceKey, page.ID),
    }
    
    // Convert with metadata
    return cl.converter.ConvertWithMetadata(page.Body.Storage.Value, meta)
}
```

**Note**: Need to add `SpaceKey` and `domain` fields to `Cloner` or pass them differently.

**Time**: 1.5 hours

---

### Step 2.4: Wire Up in Main

**File**: `main.go`

**Changes**:
```go
// After creating cloner
cloner := clone.NewCloner(client, outputDir)

// NEW: Enable markdown export if requested
if os.Getenv("CONFLUENCE_EXPORT_MARKDOWN") == "true" {
    cloner.EnableMarkdownExport()
}

// Run clone
if err := cloner.Clone(); err != nil {
    // ... error handling
}
```

**Time**: 15 minutes

---

### Step 2.5: Test End-to-End

**Manual Test**:
```bash
# Export a small space with markdown enabled
CONFLUENCE_DOMAIN="..." \
CONFLUENCE_EMAIL="..." \
CONFLUENCE_API_TOKEN="..." \
CONFLUENCE_OUTPUT_DIR="./test-markdown-export" \
CONFLUENCE_EXPORT_MARKDOWN=true \
./confluence-reader
```

**Verify**:
- [ ] Both `content.html` and `content.md` files created
- [ ] Markdown has YAML frontmatter at top
- [ ] Frontmatter contains correct metadata
- [ ] Markdown content is clean (no HTML artifacts)
- [ ] Code blocks preserved with language hints
- [ ] Internal links converted to relative MD links
- [ ] All existing tests still pass

**Time**: 1 hour

---

### Step 2.6: Update Documentation

**File**: `README.md`

**Add Section**:
```markdown
## Markdown Export

### Overview

confluence-reader can export pages as Markdown alongside HTML, making content more readable and LLM-friendly.

### Usage

Enable markdown export with environment variable:

```bash
CONFLUENCE_EXPORT_MARKDOWN=true ./confluence-reader
```

### Output Structure

With markdown export enabled, each page will have:

```
{SPACE_KEY}/pages/{PAGE_ID}_{Title}/
├── metadata.json      # Page metadata (existing)
├── content.html       # Original HTML storage format (existing)
└── content.md         # NEW: Markdown conversion
```

### Markdown Format

Each `.md` file includes:

1. **YAML Frontmatter** - Page metadata (title, ID, version, etc.)
2. **Markdown Content** - Clean, readable markdown

Example:
```markdown
---
title: "Getting Started"
confluence_id: "123456"
space_key: "DOC"
version: 5
last_updated: "2024-01-15T10:30:00Z"
---

# Getting Started

Welcome to the documentation...
```

### Supported Features

- ✅ Headings, lists, tables, code blocks
- ✅ Bold, italic, links, images
- ✅ Confluence code macros → Fenced code blocks
- ✅ Warning/Info panels → Blockquotes with emoji
- ✅ Internal page links → Relative markdown links
- ✅ TOC macros removed (redundant)

### LLM Usage

Feed markdown files to ChatGPT, Claude, or other LLMs:

```bash
cat content.md | pbcopy  # macOS
# Paste into LLM chat
```

Or use with LLM APIs:
```python
with open('content.md') as f:
    context = f.read()
    
response = openai.ChatCompletion.create(
    model="gpt-4",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": f"Based on this documentation:\n\n{context}\n\nAnswer: ..."}
    ]
)
```
```

**Time**: 30 minutes

---

## Phase 2 Summary

### Total Effort: 4-6 hours

### Deliverables:
1. ✅ Frontmatter support (`ConvertWithMetadata()`)
2. ✅ Integration with clone pipeline
3. ✅ Environment variable support
4. ✅ End-to-end testing
5. ✅ User documentation

### Success Criteria:
- [ ] Can export with `CONFLUENCE_EXPORT_MARKDOWN=true`
- [ ] Both HTML and Markdown files created
- [ ] Markdown has proper frontmatter
- [ ] All existing tests pass
- [ ] No breaking changes

---

## Phase 3: Testing & Refinement (FUTURE)

### Goal
Test on full Confluence instance, handle edge cases

### Tasks:
1. Export full instance (all spaces)
2. Validate markdown quality
3. Test with LLM (ChatGPT, Claude)
4. Handle page hierarchy (children links)
5. Fix any conversion issues found
6. Performance testing

**Estimated**: 3-4 hours

---

## Phase 4: LLM Optimization (FUTURE)

### Goal
Optimize output for LLM consumption

### Tasks:
1. Add breadcrumbs from page hierarchy
2. Add "Related Pages" section
3. Improve code block context (add titles)
4. Add cross-references
5. Clean up formatting edge cases

**Estimated**: 3-4 hours

---

## Phase 5: Git Integration (FUTURE)

### Goal
Make output git-friendly with change tracking

### Tasks:
1. Implement slug-based filenames
2. Flat page structure (vs nested)
3. Add `.confluence-sync.json` for tracking
4. Implement incremental sync (only changed pages)
5. Optional auto-commit with stats

**Estimated**: 2-3 hours

---

## Phase 6: Validation & Polish (FUTURE)

### Goal
Production-ready quality

### Tasks:
1. Comprehensive validation script
2. Performance benchmarks
3. Edge case handling (tables, nested macros)
4. Error message improvements
5. Progress output enhancements

**Estimated**: 2-3 hours

---

## Phase 7: Final UAT (FUTURE)

### Goal
User acceptance and documentation

### Tasks:
1. End-to-end test with real user
2. Collect feedback
3. Fix any issues found
4. Final documentation polish
5. Release notes

**Estimated**: 1-2 hours

---

## Total Timeline

| Phase | Effort | Cumulative |
|-------|--------|------------|
| Phase 1 | 3-4h | 3-4h |
| Phase 2 | 4-6h | 7-10h |
| Phase 3 | 3-4h | 10-14h |
| Phase 4 | 3-4h | 13-18h |
| Phase 5 | 2-3h | 15-21h |
| Phase 6 | 2-3h | 17-24h |
| Phase 7 | 1-2h | 18-26h |

**Total**: ~18-26 hours (~25 hours average)

---

**Last Updated**: 2025-11-12  
**Next**: Phase 2 - Start with Step 2.1 (Environment Variable)
