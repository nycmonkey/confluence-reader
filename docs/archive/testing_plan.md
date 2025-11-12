# Testing Plan - Markdown Export Feature

## Test Strategy

### Testing Levels
1. **Unit Tests** - Individual functions (pkg/markdown)
2. **Integration Tests** - Converter + Clone pipeline
3. **End-to-End Tests** - Full export with real Confluence
4. **Acceptance Tests** - User validation, LLM testing

---

## Phase 1: Unit Tests (COMPLETE ✅)

### pkg/markdown/converter_test.go

#### Test 1: Basic HTML Conversion ✅
```go
func TestBasicConversion(t *testing.T)
```
**Coverage**: Headings, paragraphs, bold, italic, lists  
**Status**: ✅ Passing

#### Test 2: Real Confluence Page ✅
```go
func TestRealConfluencePage(t *testing.T)
```
**Coverage**: 479 real pages from Confluence  
**Status**: ✅ Passing  
**Notes**: Validates production readiness

#### Test 3: Confluence Code Macro ✅
```go
func TestConfluenceCodeMacro(t *testing.T)
```
**Coverage**: Code blocks with CDATA, language hints  
**Status**: ✅ Passing  
**Notes**: Tests HTML escaping fix

#### Test 4: TOC Removal ✅
```go
func TestTOCRemoval(t *testing.T)
```
**Coverage**: TOC macro preprocessing  
**Status**: ✅ Passing

#### Test 5: Internal Link Conversion ✅
```go
func TestInternalLink(t *testing.T)
```
**Coverage**: Confluence page links → MD links  
**Status**: ✅ Passing  
**Notes**: Tests slug generation

---

## Phase 2: Integration Tests (NEXT SESSION)

### Test 6: Frontmatter Generation (TODO)
```go
func TestConvertWithMetadata(t *testing.T) {
    html := "<h1>Test</h1><p>Content</p>"
    meta := PageMetadata{
        Title:     "Test Page",
        PageID:    "123456",
        SpaceKey:  "TEST",
        Version:   5,
        UpdatedAt: time.Now(),
    }
    
    conv := NewConverter()
    md, err := conv.ConvertWithMetadata(html, meta)
    
    // Verify frontmatter present
    assert.Contains(t, md, "---\n")
    assert.Contains(t, md, "title: \"Test Page\"")
    assert.Contains(t, md, "confluence_id: \"123456\"")
    
    // Verify content follows
    assert.Contains(t, md, "# Test")
}
```

**Coverage**: Frontmatter generation, YAML escaping  
**Priority**: High  
**Time**: 30 minutes

---

### Test 7: Clone Integration (TODO)
```go
func TestCloneWithMarkdown(t *testing.T) {
    // Create temporary output directory
    tmpDir := t.TempDir()
    
    // Create mock client with test page
    mockClient := &MockClient{
        pages: []Page{
            {
                ID:    "123",
                Title: "Test Page",
                Body:  &Body{Storage: &Storage{Value: "<h1>Test</h1>"}},
            },
        },
    }
    
    // Create cloner with markdown enabled
    cloner := NewCloner(mockClient, tmpDir)
    cloner.EnableMarkdownExport()
    
    // Clone
    err := cloner.Clone()
    assert.NoError(t, err)
    
    // Verify both HTML and MD files created
    htmlPath := filepath.Join(tmpDir, "TEST", "pages", "123_Test_Page", "content.html")
    mdPath := filepath.Join(tmpDir, "TEST", "pages", "123_Test_Page", "content.md")
    
    assert.FileExists(t, htmlPath)
    assert.FileExists(t, mdPath)
    
    // Verify markdown has frontmatter
    mdContent, _ := os.ReadFile(mdPath)
    assert.Contains(t, string(mdContent), "---\n")
    assert.Contains(t, string(mdContent), "title:")
}
```

**Coverage**: End-to-end integration  
**Priority**: Critical  
**Time**: 1 hour  
**Dependencies**: Need mock client or real test space

---

### Test 8: Error Handling (TODO)
```go
func TestConversionError(t *testing.T) {
    // Test with invalid HTML
    html := "<<<broken>>>"
    conv := NewConverter()
    
    // Should not panic, should return error or empty string
    md, err := conv.Convert(html)
    
    if err != nil {
        assert.NotEmpty(t, err.Error())
    } else {
        // Graceful degradation - return something
        assert.NotNil(t, md)
    }
}

func TestMarkdownSaveError(t *testing.T) {
    // Test with read-only directory
    tmpDir := t.TempDir()
    os.Chmod(tmpDir, 0444) // Read-only
    
    cloner := NewCloner(mockClient, tmpDir)
    cloner.EnableMarkdownExport()
    
    // Should log warning but not fail entire clone
    err := cloner.Clone()
    // Should NOT return error (graceful degradation)
    assert.NoError(t, err)
}
```

**Coverage**: Error handling, graceful degradation  
**Priority**: Medium  
**Time**: 30 minutes

---

## Phase 3: End-to-End Tests (FUTURE)

### Test 9: Small Space Export
**Type**: Manual / Automated  
**Goal**: Export small real space (10-20 pages)

**Steps**:
```bash
# 1. Export with markdown enabled
CONFLUENCE_EXPORT_MARKDOWN=true \
CONFLUENCE_OUTPUT_DIR="./test-e2e" \
./confluence-reader

# 2. Verify structure
find ./test-e2e -name "*.md" | wc -l  # Should match page count
find ./test-e2e -name "*.html" | wc -l  # Should match page count

# 3. Validate markdown syntax
npx markdownlint ./test-e2e/**/*.md

# 4. Check frontmatter
for f in $(find ./test-e2e -name "*.md" | head -5); do
    echo "Validating $f..."
    head -20 "$f" | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)"
done
```

**Success Criteria**:
- [ ] All pages have both HTML and MD
- [ ] Markdown is valid (linter passes)
- [ ] Frontmatter is valid YAML
- [ ] No HTML artifacts in markdown
- [ ] Code blocks preserved correctly
- [ ] Internal links work

**Time**: 1 hour

---

### Test 10: Large Space Export
**Type**: Performance test  
**Goal**: Export large space (500+ pages)

**Metrics to Track**:
- Total conversion time
- Memory usage
- Disk space used
- Conversion failures (should be 0 or logged)

**Success Criteria**:
- [ ] Completes without crashing
- [ ] < 1 minute for typical instance
- [ ] Memory usage stays reasonable (< 500MB)
- [ ] All pages converted successfully

**Time**: 1 hour

---

## Phase 4: Acceptance Tests (FUTURE)

### Test 11: LLM Consumption Test
**Type**: Manual validation  
**Goal**: Verify LLM can understand markdown

**Steps**:
1. Export documentation space
2. Select 3-5 representative pages
3. Feed to ChatGPT/Claude
4. Ask questions about content
5. Verify LLM understands correctly

**Questions to Test**:
- "Summarize this page"
- "What are the main steps in this guide?"
- "What code examples are shown?"
- "How does this relate to [other page]?"

**Success Criteria**:
- [ ] LLM provides accurate summaries
- [ ] LLM extracts correct information
- [ ] LLM follows code examples
- [ ] No confusion from formatting

**Time**: 30 minutes

---

### Test 12: Git Diff Quality Test
**Type**: Manual validation  
**Goal**: Verify diffs are meaningful

**Steps**:
```bash
# 1. Initial export
git init confluence-export
cd confluence-export
CONFLUENCE_EXPORT_MARKDOWN=true ../confluence-reader
git add .
git commit -m "Initial export"

# 2. Make small change in Confluence (edit one page)

# 3. Export again
CONFLUENCE_EXPORT_MARKDOWN=true ../confluence-reader

# 4. Check git diff
git diff

# 5. Verify diff shows only actual content change
```

**Success Criteria**:
- [ ] Only changed pages show in diff
- [ ] Diff shows actual content change (not formatting noise)
- [ ] Frontmatter version number updated
- [ ] No spurious whitespace changes

**Time**: 30 minutes

---

### Test 13: User Acceptance Test
**Type**: Beta user testing  
**Goal**: Real user validates feature

**Test Plan**:
1. Provide binary to user
2. User exports their Confluence
3. User reviews markdown quality
4. User tries LLM use case
5. Collect feedback

**Feedback Questions**:
- Is the markdown readable?
- Are any elements missing or broken?
- Does it work with your LLM workflow?
- Would you use this regularly?
- Any unexpected errors?

**Time**: User's time + 1 hour for feedback review

---

## Test Coverage Goals

| Component | Current | Target |
|-----------|---------|--------|
| pkg/markdown | 100% | 100% |
| pkg/clone (new code) | 0% | 80% |
| Integration | 0% | 70% |
| E2E | 0% | Manual tests |

---

## Regression Test Suite

After each phase, run full regression:

```bash
# 1. Run all unit tests
go test ./...

# 2. Run integration tests
go test ./pkg/clone -v -run Integration

# 3. Quick E2E test (small space)
./test-e2e-small.sh

# 4. Verify no existing functionality broken
# - Export without markdown flag
# - Verify HTML still created correctly
# - Verify attachments still downloaded
```

**Time**: 15 minutes per regression run

---

## Test Data Management

### Sample Data
- [x] 479 pages from real Confluence (Phase 1)
- [ ] Small test space (10-20 pages) for quick tests
- [ ] Large test space (500+ pages) for performance tests
- [ ] Edge case pages (complex tables, nested macros, etc.)

### Mock Data
- [ ] Create mock Confluence client for unit tests
- [ ] Sample pages with each macro type
- [ ] Edge case HTML (malformed, deeply nested, etc.)

---

## Testing Tools

### Already Using
- Go testing framework (`testing` package)
- `go test -v` for verbose output
- `go test -cover` for coverage

### To Add (Optional)
- `markdownlint` - Validate markdown syntax
- `yamllint` - Validate YAML frontmatter
- `golangci-lint` - Additional linting
- `testify` - Assert library (optional)

---

## Test Automation (Future)

### CI/CD Pipeline
```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.23
      - run: go test ./... -v -cover
      - run: go test ./... -race
```

**Time to set up**: 1 hour

---

## Next Session Test Checklist

Before starting Phase 2:
- [ ] All Phase 1 tests still passing
- [ ] No regressions in existing tests

During Phase 2:
- [ ] Write Test 6 (frontmatter) FIRST
- [ ] Implement frontmatter, verify test passes
- [ ] Write Test 7 (integration) BEFORE integration
- [ ] Implement integration, verify test passes
- [ ] Run full regression suite

After Phase 2:
- [ ] All tests passing (unit + integration)
- [ ] Code coverage report shows good coverage
- [ ] Manual E2E test successful

---

**Last Updated**: 2025-11-12  
**Test Status**: 5/5 Phase 1 tests passing  
**Next**: Write Phase 2 integration tests
