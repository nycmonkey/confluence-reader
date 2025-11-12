# Markdown Export Plan - LLM-Friendly Confluence Content

## Goal
Export Confluence content as clean, LLM-friendly Markdown with git-friendly versioning support.

## Success Criteria
1. ✅ Convert Confluence storage format (HTML) to clean Markdown
2. ✅ Preserve document structure (headings, lists, code blocks, tables)
3. ✅ Handle Confluence-specific macros gracefully
4. ✅ Git-friendly: Consistent formatting, stable diffs, one file per page
5. ✅ LLM-optimized: Clear structure, minimal noise, good context preservation
6. ✅ Track changes: Metadata to detect updates since last sync

---

## Phase 1: Research & Analysis

### Step 1.1: Analyze Sample Data Structure
**Action**: Clone a small sample of real Confluence content to understand formats

**Tasks**:
```bash
# Create test output directory
mkdir -p ./test-sample

# Clone a single small space for analysis
CONFLUENCE_DOMAIN="..." \
CONFLUENCE_EMAIL="..." \
CONFLUENCE_API_TOKEN="..." \
CONFLUENCE_OUTPUT_DIR="./test-sample" \
./confluence-reader
```

**Inspect**:
- [ ] Examine `content.html` files - what HTML/macros are present?
- [ ] Check `metadata.json` - what version/timestamp data is available?
- [ ] Identify common Confluence macros (code, info, warning, toc, etc.)
- [ ] Note attachment references and how they're embedded
- [ ] Document any nested page hierarchies

**Deliverable**: `CONFLUENCE_FORMATS.md` documenting observed patterns

---

### Step 1.2: Research HTML-to-Markdown Solutions
**Action**: Evaluate Go libraries for HTML→Markdown conversion

**Options to Research**:
1. **github.com/JohannesKaufmann/html-to-markdown** (pure Go, customizable)
   - Pros: Native Go, plugin system, widely used
   - Cons: May need custom rules for Confluence macros
   
2. **github.com/gomarkdown/markdown** (bidirectional)
   - Pros: Full parsing + rendering
   - Cons: Primarily for rendering, may not handle complex HTML
   
3. **Pandoc** (external binary)
   - Pros: Industry standard, handles everything
   - Cons: External dependency, harder to customize

4. **Custom implementation** using `golang.org/x/net/html`
   - Pros: Full control, Confluence-specific optimization
   - Cons: More development effort

**Tasks**:
- [ ] Test each option with sample Confluence HTML
- [ ] Evaluate macro handling (code blocks, panels, etc.)
- [ ] Check table conversion quality
- [ ] Assess extensibility for custom rules

**Deliverable**: Decision matrix with recommendation

---

## Phase 2: Initial Implementation

### Step 2.1: Add Markdown Conversion Package
**Action**: Create `pkg/markdown/` package with conversion logic

**Structure**:
```
pkg/markdown/
├── converter.go        # Main conversion logic
├── converter_test.go   # Unit tests
├── macros.go          # Confluence macro handlers
├── macros_test.go     # Macro tests
└── fixtures/          # Test HTML samples
```

**Interface Design**:
```go
type Converter interface {
    // Convert Confluence storage HTML to Markdown
    Convert(html string) (markdown string, err error)
    
    // ConvertWithMetadata includes frontmatter
    ConvertWithMetadata(html string, meta PageMetadata) (markdown string, err error)
}

type PageMetadata struct {
    Title      string
    PageID     string
    SpaceKey   string
    Version    int
    UpdatedAt  time.Time
    Author     string
    ParentID   string
}
```

**Deliverable**: Working converter package with tests

---

### Step 2.2: Implement Core HTML→Markdown
**Action**: Convert basic HTML elements to Markdown

**Priority Conversions**:
1. **Structure**: `<h1>-<h6>` → `# - ######`
2. **Formatting**: `<strong>`, `<em>`, `<code>`, `<pre>`
3. **Lists**: `<ul>`, `<ol>`, `<li>` → `- ` and `1. `
4. **Links**: `<a href>` → `[text](url)`
5. **Images**: `<img>` → `![alt](url)` with attachment path fixup
6. **Tables**: `<table>` → Markdown tables
7. **Code blocks**: `<pre><code>` → ````lang` blocks

**Test Cases**:
- [ ] Simple text with inline formatting
- [ ] Nested lists (3+ levels)
- [ ] Tables with alignment
- [ ] Code blocks with language hints
- [ ] Mixed content paragraphs

**Deliverable**: Core converter passing all basic tests

---

### Step 2.3: Handle Confluence Macros
**Action**: Convert Confluence-specific macros to Markdown equivalents

**Common Macros**:
```html
<!-- Info panel -->
<ac:structured-macro ac:name="info">
  <ac:rich-text-body><p>Important info</p></ac:rich-text-body>
</ac:structured-macro>
→ > ℹ️ **Info**: Important info

<!-- Code block -->
<ac:structured-macro ac:name="code">
  <ac:parameter ac:name="language">python</ac:parameter>
  <ac:plain-text-body><![CDATA[def hello():...]]></ac:plain-text-body>
</ac:structured-macro>
→ ```python
  def hello():...
  ```

<!-- Warning panel -->
<ac:structured-macro ac:name="warning">
  → > ⚠️ **Warning**: ...

<!-- Table of Contents -->
<ac:structured-macro ac:name="toc">
  → <!-- TOC will be in headings --> or remove

<!-- Attachments -->
<ac:image>
  <ri:attachment ri:filename="diagram.png" />
</ac:image>
→ ![diagram.png](./attachments/diagram.png)
```

**Tasks**:
- [ ] Identify all macros in sample data
- [ ] Implement handler for each macro type
- [ ] Test with real Confluence HTML samples
- [ ] Document unsupported macros (strip gracefully)

**Deliverable**: Macro handlers with comprehensive tests

---

## Phase 3: Integration & Testing

### Step 3.1: Integrate into Clone Pipeline
**Action**: Add markdown export option to cloner

**Changes**:
```go
// pkg/clone/clone.go
type Cloner struct {
    client         *client.Client
    outputDir      string
    exportMarkdown bool  // NEW: Enable markdown export
    converter      *markdown.Converter  // NEW
}

func (cl *Cloner) clonePage(page client.Page, pagesDir string) error {
    // ... existing save logic ...
    
    // NEW: Export markdown if enabled
    if cl.exportMarkdown && fullPage.Body != nil && fullPage.Body.Storage != nil {
        md, err := cl.converter.ConvertWithMetadata(
            fullPage.Body.Storage.Value,
            extractMetadata(fullPage),
        )
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
```

**CLI Changes** (`main.go`):
```go
// Add prompt for markdown export
fmt.Print("Export as Markdown? (y/n, default: y): ")
scanner.Scan()
exportMarkdown := strings.ToLower(strings.TrimSpace(scanner.Text()))
if exportMarkdown == "" || exportMarkdown == "y" {
    cloner.EnableMarkdownExport()
}
```

**Environment Variable**: `CONFLUENCE_EXPORT_MARKDOWN=true`

**Deliverable**: Integrated markdown export working end-to-end

---

### Step 3.2: Add Git-Friendly Frontmatter
**Action**: Prepend YAML frontmatter for metadata tracking

**Frontmatter Format**:
```markdown
---
title: "Page Title"
confluence_id: "123456"
space_key: "ENG"
parent_id: "789012"
version: 5
last_updated: "2024-01-15T10:30:00Z"
author: "user@example.com"
confluence_url: "https://company.atlassian.net/wiki/spaces/ENG/pages/123456"
---

# Page Title

Page content here...
```

**Benefits**:
- Git can track version changes via frontmatter
- LLMs can see page context and relationships
- Easy to detect which pages have updates
- Can generate index/TOC from frontmatter

**Tasks**:
- [ ] Implement frontmatter generation
- [ ] Test YAML parsing (verify valid YAML)
- [ ] Add option to include/exclude frontmatter
- [ ] Test with various special characters in titles

**Deliverable**: Pages with valid YAML frontmatter

---

### Step 3.3: Test with Real Data Sample
**Action**: Run full export on real Confluence space

**Test Script**:
```bash
#!/bin/bash
# test-markdown-export.sh

# Clean previous test
rm -rf ./markdown-test

# Run export with markdown enabled
CONFLUENCE_DOMAIN="company.atlassian.net" \
CONFLUENCE_EMAIL="user@example.com" \
CONFLUENCE_API_TOKEN="$TOKEN" \
CONFLUENCE_OUTPUT_DIR="./markdown-test" \
CONFLUENCE_EXPORT_MARKDOWN=true \
./confluence-reader

# Verify output
echo "=== Markdown Files Generated ==="
find ./markdown-test -name "*.md" | wc -l

echo "=== Sample Content ==="
find ./markdown-test -name "*.md" | head -3 | xargs -I {} sh -c 'echo "File: {}" && head -20 {}'

echo "=== Check for HTML remnants ==="
grep -r "<ac:" ./markdown-test/**/*.md || echo "✓ No Confluence macros found"
grep -r "<div" ./markdown-test/**/*.md || echo "✓ No div tags found"

echo "=== Frontmatter validation ==="
for file in $(find ./markdown-test -name "*.md" | head -5); do
    echo "Checking $file..."
    # Extract frontmatter and validate YAML
    sed -n '/^---$/,/^---$/p' "$file" | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)" && echo "✓ Valid YAML"
done
```

**Manual Review Checklist**:
- [ ] Headings properly converted
- [ ] Lists nested correctly
- [ ] Code blocks with language hints
- [ ] Tables readable
- [ ] Links functional (internal and external)
- [ ] Images point to correct paths
- [ ] No HTML artifacts (`<div>`, `<span>`, etc.)
- [ ] Frontmatter valid YAML
- [ ] Special characters escaped properly

**Deliverable**: Quality report with issues found

---

## Phase 4: Refinement & Optimization

### Step 4.1: Handle Edge Cases
**Action**: Fix issues found in testing

**Common Issues to Address**:
1. **Nested macros**: Macro inside panel inside table
2. **Malformed HTML**: Unclosed tags, invalid nesting
3. **Special characters**: Quotes, apostrophes in titles
4. **Very long pages**: Memory efficiency for large docs
5. **Empty sections**: Graceful handling of blank content
6. **Attachments**: Update image links to relative paths
7. **Internal links**: Convert Confluence page links to relative MD links

**Test Cases**:
```go
// pkg/markdown/converter_test.go
var edgeCases = []struct{
    name     string
    input    string
    expected string
}{
    {
        name: "nested_panels",
        input: `<ac:structured-macro ac:name="info">
                  <ac:rich-text-body>
                    <ac:structured-macro ac:name="code">...</>
                  </ac:rich-text-body>
                </ac:structured-macro>`,
        expected: "> ℹ️ **Info**:\n> ```\n> ...\n> ```",
    },
    // ... more cases
}
```

**Deliverable**: Robust converter handling all edge cases

---

### Step 4.2: Optimize for LLM Consumption
**Action**: Enhance output for LLM-friendliness

**Optimizations**:

1. **Add Context Headers**:
```markdown
---
title: "API Authentication"
space: "Engineering"
breadcrumb: "Home > Engineering > API Docs > Authentication"
related_pages:
  - "API Rate Limits"
  - "OAuth Setup"
---

<!-- CONTEXT: This page describes authentication for the REST API -->
<!-- SPACE: Engineering Documentation -->
<!-- LAST UPDATED: 2024-01-15 -->

# API Authentication
...
```

2. **Preserve Page Hierarchy**:
```
./ENGINEERING/
├── README.md                    # Space overview
├── _index.md                    # TOC with all pages
└── pages/
    ├── architecture/
    │   ├── overview.md
    │   └── decisions/
    │       └── adr-001.md
```

3. **Clean Formatting**:
- Remove excessive blank lines (max 2 consecutive)
- Normalize list indentation (2 spaces)
- Consistent heading levels (no h1→h4 jumps)
- Strip HTML comments unless semantic

4. **Add Cross-References**:
- Convert Confluence page links to relative MD links
- Add "See also" sections from page relationships
- Generate breadcrumbs from parent hierarchy

**Deliverable**: LLM-optimized markdown files

---

### Step 4.3: Implement Change Detection
**Action**: Track page versions to enable incremental updates

**Strategy**:
```go
// pkg/sync/tracker.go
type PageState struct {
    PageID      string    `json:"page_id"`
    Version     int       `json:"version"`
    UpdatedAt   time.Time `json:"updated_at"`
    ContentHash string    `json:"content_hash"` // SHA256 of content
    LastSynced  time.Time `json:"last_synced"`
}

type SyncTracker struct {
    states map[string]PageState
}

// Save to .confluence-sync.json in output dir
func (st *SyncTracker) Save(path string) error
func (st *SyncTracker) Load(path string) error
func (st *SyncTracker) NeedsUpdate(pageID string, version int) bool
```

**Workflow**:
1. Load previous sync state from `.confluence-sync.json`
2. For each page, check if version > last synced version
3. Only fetch/convert pages that changed
4. Update sync state after successful conversion
5. Save new sync state

**Benefits**:
- Faster subsequent syncs (skip unchanged pages)
- Git history shows actual content changes
- Can report "5 pages updated, 2 new, 1 deleted"

**Deliverable**: Incremental sync capability

---

## Phase 5: Git Integration

### Step 5.1: Design Git-Friendly Structure
**Action**: Optimize directory layout for version control

**Proposed Structure**:
```
confluence-export/
├── .confluence-sync.json       # Sync state (gitignored)
├── .gitignore                  # Ignore sync state, temp files
├── README.md                   # Export overview with stats
├── spaces/
│   ├── ENGINEERING/
│   │   ├── README.md           # Space overview
│   │   ├── _index.md           # Full TOC
│   │   └── pages/
│   │       ├── architecture-overview.md
│   │       ├── api-documentation.md
│   │       └── deployment-guide.md
│   └── PRODUCT/
│       └── ...
└── attachments/                # Optional: flatten attachments
    └── ENGINEERING/
        └── diagram-abc123.png
```

**Key Decisions**:
- **Flat page structure**: Easier navigation, simpler links
- **Slug-based filenames**: `architecture-overview.md` not `123456_Architecture_Overview.md`
- **Attachments separate**: Keep binary diffs isolated
- **README per space**: Context for each space
- **Global index**: Top-level overview of all content

**Tasks**:
- [ ] Implement slug generation (title → kebab-case)
- [ ] Handle slug collisions (append ID if needed)
- [ ] Generate space READMEs with page list
- [ ] Create global README with space inventory

**Deliverable**: Clean, navigable git repository structure

---

### Step 5.2: Add Git Automation
**Action**: Optionally auto-commit changes after sync

**Feature**: `--git-commit` flag or `CONFLUENCE_GIT_COMMIT=true`

**Implementation**:
```go
// pkg/git/auto.go
func AutoCommit(repoPath string) error {
    // Check if directory is git repo
    if !isGitRepo(repoPath) {
        return fmt.Errorf("not a git repository")
    }
    
    // Stage all changes
    cmd := exec.Command("git", "-C", repoPath, "add", ".")
    if err := cmd.Run(); err != nil {
        return err
    }
    
    // Generate commit message with stats
    stats := getChangeStats(repoPath)
    message := fmt.Sprintf(`Confluence sync: %s

- %d pages updated
- %d pages added
- %d pages removed
- %d attachments changed

Synced at: %s`, 
        time.Now().Format("2006-01-02 15:04"),
        stats.Updated, stats.Added, stats.Removed, stats.AttachmentsChanged,
        time.Now().Format(time.RFC3339),
    )
    
    // Commit
    cmd = exec.Command("git", "-C", repoPath, "commit", "-m", message)
    return cmd.Run()
}
```

**Workflow**:
```bash
# Initialize git repo for exports
cd confluence-export
git init
git add .
git commit -m "Initial Confluence export"

# Subsequent syncs with auto-commit
CONFLUENCE_GIT_COMMIT=true ./confluence-reader
# → Automatically commits changes with stats
```

**Deliverable**: Auto-commit feature for change tracking

---

## Phase 6: Validation & Polish

### Step 6.1: Comprehensive Testing
**Action**: Test entire pipeline with multiple spaces

**Test Scenarios**:
1. **Small space** (5-10 pages): Verify correctness
2. **Large space** (100+ pages): Test performance
3. **Complex content**: Macros, tables, code, images
4. **Incremental sync**: Modify pages, verify delta detection
5. **Git workflow**: Clone, sync, commit, diff

**Validation Script**:
```bash
#!/bin/bash
# validate-export.sh

echo "=== Running full validation ==="

# 1. Check markdown syntax
echo "Checking markdown syntax..."
find ./spaces -name "*.md" -exec npx markdownlint-cli {} \; || echo "Note: markdownlint not installed"

# 2. Verify frontmatter
echo "Validating YAML frontmatter..."
python3 << 'EOF'
import os, yaml
from pathlib import Path

errors = []
for md_file in Path('./spaces').rglob('*.md'):
    with open(md_file) as f:
        content = f.read()
        if content.startswith('---\n'):
            try:
                end = content.find('\n---\n', 4)
                yaml.safe_load(content[4:end])
            except Exception as e:
                errors.append(f"{md_file}: {e}")

if errors:
    print("❌ YAML errors found:")
    for e in errors: print(f"  {e}")
else:
    print("✅ All frontmatter valid")
EOF

# 3. Check for HTML remnants
echo "Checking for unconverted HTML..."
if grep -r "<ac:" ./spaces/**/*.md; then
    echo "❌ Confluence macros still present"
else
    echo "✅ No Confluence macros found"
fi

# 4. Verify links
echo "Checking internal links..."
find ./spaces -name "*.md" -exec grep -l "](.*\.md)" {} \; | head -5 | xargs -I {} sh -c 'echo "Links in {}:" && grep -o "](.*\.md)" {}'

# 5. Check file structure
echo "=== Directory structure ==="
tree -L 3 ./spaces | head -20

echo "=== Validation complete ==="
```

**Deliverable**: Validation report confirming quality

---

### Step 6.2: Performance Optimization
**Action**: Ensure efficient processing for large instances

**Optimizations**:
1. **Concurrent conversion**: Convert pages to markdown in parallel
2. **Streaming writes**: Don't buffer entire MD in memory
3. **Caching**: Cache converted macros/common patterns
4. **Incremental sync**: Skip unchanged pages (already implemented)

**Benchmark**:
```go
// pkg/markdown/converter_test.go
func BenchmarkConvert(b *testing.B) {
    html := loadFixture("large_page.html")
    conv := NewConverter()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := conv.Convert(html)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**Performance Targets**:
- Convert 1000 pages in < 2 minutes
- Memory usage < 500MB for 10,000 page instance
- Incremental sync < 30 seconds for 100 changed pages

**Deliverable**: Performance benchmarks and optimizations

---

### Step 6.3: Documentation & Examples
**Action**: Document the markdown export feature

**New Documentation**:

1. **MARKDOWN_EXPORT.md**: Complete guide
   - How to enable markdown export
   - Output structure explanation
   - Git workflow recommendations
   - LLM usage tips

2. **Update README.md**: Add markdown export section

3. **Example outputs**: Include sample exports in repo
   - `examples/sample-space/` with before/after

4. **Configuration guide**: All options documented
   - `CONFLUENCE_EXPORT_MARKDOWN`
   - `CONFLUENCE_MARKDOWN_FORMAT` (with-frontmatter, plain)
   - `CONFLUENCE_GIT_COMMIT`
   - `CONFLUENCE_INCREMENTAL`

**Deliverable**: Complete documentation for users

---

## Phase 7: Final Validation

### Step 7.1: End-to-End Test with Real Data
**Action**: Full workflow test with actual Confluence instance

**Test Plan**:
```bash
# 1. Fresh export with markdown
./confluence-reader --export-markdown

# 2. Verify quality
./scripts/validate-export.sh

# 3. Initialize git
cd confluence-export
git init
git add .
git commit -m "Initial export"

# 4. Make changes in Confluence (manually edit a page)

# 5. Incremental sync
cd ..
./confluence-reader --export-markdown --incremental

# 6. Check git diff
cd confluence-export
git diff

# 7. Verify only changed pages updated
git status

# 8. Commit changes
git add .
git commit -m "Incremental sync"

# 9. Test LLM consumption
# Copy markdown to LLM context and ask questions
```

**Success Checklist**:
- [ ] All pages converted to markdown
- [ ] No HTML artifacts in output
- [ ] Frontmatter valid and useful
- [ ] Git diffs are clean and meaningful
- [ ] Incremental sync detects changes correctly
- [ ] Links work (can navigate between pages)
- [ ] Images display correctly
- [ ] Code blocks have language hints
- [ ] Tables are readable
- [ ] LLM can understand and answer questions

**Deliverable**: Successful end-to-end demonstration

---

### Step 7.2: User Acceptance Testing
**Action**: Have real users test the feature

**Test Users**:
1. You (primary user)
2. Team member with different Confluence instance
3. Someone with large/complex Confluence content

**Feedback Collection**:
- [ ] Is markdown output readable?
- [ ] Are any important elements lost in conversion?
- [ ] Is git workflow intuitive?
- [ ] Does incremental sync work as expected?
- [ ] Any unexpected errors or warnings?
- [ ] Performance acceptable for your Confluence size?

**Deliverable**: User feedback incorporated into final version

---

## Rollout Plan

### Release Strategy

**v2.0.0 - Markdown Export**
- Core HTML→Markdown conversion
- Frontmatter with metadata
- Git-friendly structure
- Incremental sync
- Documentation

**Release Checklist**:
- [ ] All tests passing
- [ ] Documentation complete
- [ ] Example outputs in repo
- [ ] CHANGELOG.md updated
- [ ] README.md updated
- [ ] Performance benchmarks documented
- [ ] Known limitations documented

---

## Success Metrics

**How we know it's working**:
1. ✅ Can export entire Confluence instance to markdown
2. ✅ Markdown is clean (no HTML artifacts)
3. ✅ Git diffs show meaningful changes
4. ✅ LLMs can consume and understand the content
5. ✅ Incremental syncs are fast (< 1 minute for typical changes)
6. ✅ Links and images work correctly
7. ✅ Users report satisfaction with output quality

---

## Timeline Estimate

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 1: Research | 2-3 hours | 3h |
| Phase 2: Initial Implementation | 4-6 hours | 9h |
| Phase 3: Integration & Testing | 3-4 hours | 13h |
| Phase 4: Refinement | 3-4 hours | 17h |
| Phase 5: Git Integration | 2-3 hours | 20h |
| Phase 6: Validation & Polish | 2-3 hours | 23h |
| Phase 7: Final Validation | 1-2 hours | 25h |

**Total**: ~25 hours (3-4 days of focused work)

---

## Next Steps

To begin, I need:
1. ✅ Confirmation to proceed with Phase 1
2. ⏳ Access to test Confluence instance (domain, email, API token)
3. ⏳ Sample space to use for initial testing
4. ⏳ Confirmation on preferred HTML→Markdown library

**Ready to start Phase 1?** I can begin by cloning a sample of your Confluence data to analyze the actual HTML/macro formats we need to handle.
