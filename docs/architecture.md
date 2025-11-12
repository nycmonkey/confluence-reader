# Technical Architecture - Markdown Export

## Overview

The Markdown export feature extends the existing confluence-reader tool with a three-stage conversion pipeline that transforms Confluence's HTML storage format into clean, LLM-friendly Markdown.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Confluence API v2                         │
└────────────────────┬────────────────────────────────────────┘
                     │ HTTPS
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                 pkg/client/client.go                         │
│  • GetSpaces()                                               │
│  • GetPages(spaceID)                                         │
│  • GetPageContent(pageID) → HTML storage format              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                 pkg/clone/clone.go                           │
│  • Clone()         ┌──────────────────────┐                 │
│  • cloneSpace()    │  NEW: if markdown    │                 │
│  • clonePage() ────┤  enabled, convert    │                 │
│                    │  HTML → Markdown     │                 │
│                    └──────────┬───────────┘                 │
└───────────────────────────────┼─────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│            pkg/markdown/converter.go (NEW)                   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Stage 1: Preprocess (Confluence → Standard HTML)     │  │
│  │  • Remove TOC macros                                 │  │
│  │  • Convert code macros → <pre><code>                 │  │
│  │  • Convert panels → <blockquote>                     │  │
│  │  • Convert internal links → <a href="slug.md">       │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       ▼                                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Stage 2: Convert (HTML → Markdown)                   │  │
│  │  Uses: github.com/JohannesKaufmann/                  │  │
│  │        html-to-markdown/v2                           │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       ▼                                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Stage 3: Postprocess (Clean Markdown)                │  │
│  │  • Normalize blank lines (max 2 consecutive)         │  │
│  │  • Trim trailing whitespace                          │  │
│  │  • Ensure single trailing newline                    │  │
│  └────────────────────┬─────────────────────────────────┘  │
└────────────────────────┼────────────────────────────────────┘
                         ▼
                  ┌─────────────┐
                  │ content.md  │ (on disk)
                  └─────────────┘
```

## Component Details

### pkg/markdown/converter.go

**Purpose**: Convert Confluence HTML to Markdown

**Key Functions**:
```go
type Converter struct {}

func NewConverter() *Converter

// Basic conversion
func (c *Converter) Convert(html string) (string, error)

// With metadata (frontmatter)
func (c *Converter) ConvertWithMetadata(html string, meta PageMetadata) (string, error)
```

**Preprocessing Functions**:
- `preProcess(html)` - Main preprocessing coordinator
- `removeTOCMacros(html)` - Strip TOC macros
- `convertCodeMacros(html)` - Code blocks with language hints
- `convertPanelMacros(html)` - Warning/Info/Note panels
- `convertInternalLinks(html)` - Page links to relative MD links
- `removeChildrenMacro(html)` - Child page listing

**Postprocessing Functions**:
- `postProcess(markdown)` - Clean up output
- `titleToSlug(title)` - Convert page title to URL-safe slug

### Integration Points

**pkg/clone/clone.go** (modifications needed):
```go
type Cloner struct {
    client         *client.Client
    outputDir      string
    exportMarkdown bool              // NEW
    converter      *markdown.Converter // NEW
}

func (cl *Cloner) clonePage(page client.Page, pagesDir string) error {
    // ... existing HTML save logic ...
    
    // NEW: Export markdown if enabled
    if cl.exportMarkdown && fullPage.Body != nil {
        md, err := cl.converter.Convert(fullPage.Body.Storage.Value)
        if err != nil {
            log.Printf("Warning: Failed to convert to markdown: %v", err)
        } else {
            mdPath := filepath.Join(pageDir, "content.md")
            os.WriteFile(mdPath, []byte(md), 0644)
        }
    }
}
```

**main.go** (modifications needed):
```go
// Add prompt or env var check
exportMarkdown := os.Getenv("CONFLUENCE_EXPORT_MARKDOWN") == "true"
if exportMarkdown {
    fmt.Println("Markdown export enabled")
}

cloner := clone.NewCloner(client, outputDir)
if exportMarkdown {
    cloner.EnableMarkdownExport() // NEW method
}
```

## Directory Structure

### Current
```
{output-dir}/
├── {SPACE_KEY}/
│   ├── space.json
│   └── pages/
│       └── {PAGE_ID}_{Page_Title}/
│           ├── metadata.json
│           ├── content.html
│           └── attachments/
│               └── {filename}.{ext}
```

### With Markdown Export
```
{output-dir}/
├── {SPACE_KEY}/
│   ├── space.json
│   └── pages/
│       └── {PAGE_ID}_{Page_Title}/
│           ├── metadata.json
│           ├── content.html        # Original HTML
│           ├── content.md          # NEW: Markdown conversion
│           └── attachments/
│               └── {filename}.{ext}
```

### Future (Git-Friendly)
```
{output-dir}/
├── .confluence-sync.json          # Sync state
├── README.md                      # Export overview
└── spaces/
    └── {SPACE_KEY}/
        ├── README.md              # Space overview
        └── pages/
            └── {page-slug}.md     # Clean slug filenames
```

## Data Flow

1. **Clone Process** (pkg/clone/clone.go)
   - Fetches page with HTML storage format from API
   - Saves `content.html` (existing behavior)

2. **Markdown Conversion** (NEW)
   - Checks if markdown export enabled
   - Calls `converter.Convert(html)`
   - Saves `content.md` alongside HTML

3. **Metadata Generation** (TODO: Phase 2)
   - Extracts page metadata (title, ID, version, etc.)
   - Generates YAML frontmatter
   - Prepends to Markdown content

## Dependencies

### External Libraries
```
github.com/JohannesKaufmann/html-to-markdown/v2 v2.4.0
  ├── github.com/JohannesKaufmann/dom v0.2.0
  └── golang.org/x/net v0.43.0
```

**Rationale**: 
- Pure Go (no external binaries)
- Well-maintained, widely used
- Excellent HTML→Markdown quality
- Fast (~2ms per page)

### Internal Dependencies
- `pkg/client` - Fetch page content
- `pkg/clone` - Orchestrate cloning process
- `pkg/markdown` - Convert HTML to Markdown (NEW)

## Performance Characteristics

### Current Measurements (Phase 1)
- **Conversion time**: ~2ms per page
- **Memory usage**: Negligible (<1MB for 479 pages)
- **Output size**: ~47% reduction (HTML → Markdown)

### Projected Performance
- **10,000 pages**: ~20 seconds for conversion
- **Bottleneck**: Network I/O (Confluence API), not conversion
- **Concurrency**: Already uses 5 concurrent page downloads

## Error Handling

### Three-Tier Strategy (Inherited from Clone)
1. **Critical errors** - Exit immediately (auth, network)
2. **Space-level errors** - Log warning, continue to next space
3. **Page-level errors** - Log warning, continue to next page

### Markdown-Specific Errors
- **Conversion failure** - Log warning, keep HTML, continue
- **Empty output** - Log warning, investigate preprocessor
- **Malformed HTML** - Library handles gracefully (best effort)

## Testing Strategy

### Unit Tests (pkg/markdown)
- [x] Basic HTML conversion
- [x] Confluence code macros
- [x] TOC removal
- [x] Internal links
- [x] Real Confluence pages (479 samples)

### Integration Tests (Phase 2)
- [ ] End-to-end export with markdown enabled
- [ ] Verify both HTML and MD files created
- [ ] Check frontmatter generation
- [ ] Validate internal links work

### Acceptance Tests (Phase 3)
- [ ] Export full Confluence instance
- [ ] Feed to LLM, verify understanding
- [ ] Check git diff quality
- [ ] Performance test (large instances)

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Complex HTML not handled | Medium | Extensive testing, graceful degradation |
| Performance on large instances | Low | Already fast, concurrent processing |
| Malformed Confluence HTML | Low | Library handles errors well |
| Internal link resolution | Medium | Build page index, slug mapping |
| Breaking existing functionality | High | Keep HTML export, add MD as optional |

## Future Enhancements

### Phase 4+
- Incremental sync (only changed pages)
- Git auto-commit with stats
- Attachment reference validation
- Custom Markdown plugins
- Export format options (GFM, CommonMark, etc.)

---

**Status**: Architecture validated in Phase 1  
**Next**: Implement integration (Phase 2)
