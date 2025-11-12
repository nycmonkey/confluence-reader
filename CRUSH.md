# CRUSH.md - Confluence Reader Agent Guide

## Project Overview

**confluence-reader** is a production-ready Go CLI tool that clones Atlassian Confluence Cloud content locally using the official REST API v2. It downloads spaces, pages, and attachments in an organized directory structure with metadata preservation.

**Key characteristics:**
- Zero external dependencies (pure Go standard library)
- ~634 lines of Go code across 3 main files
- 64.6% test coverage on client package
- Interactive CLI with environment variable support
- Single binary deployment (cross-platform)

## Essential Commands

### Build & Run
```bash
# Build the application
make build                      # Creates ./confluence-reader binary
go build -o confluence-reader   # Alternative direct build

# Run the application
make run                        # Build and run
./confluence-reader            # Run directly (prompts for input)

# Run with environment variables (non-interactive)
CONFLUENCE_DOMAIN="company.atlassian.net" \
CONFLUENCE_EMAIL="user@example.com" \
CONFLUENCE_API_TOKEN="token" \
CONFLUENCE_OUTPUT_DIR="./backups/$(date +%Y%m%d_%H%M%S)" \
./confluence-reader
```

### Testing
```bash
# Run all tests
make test                       # Verbose output
go test ./...                   # Standard output

# Run with coverage
make coverage                   # Generates coverage.html
go test ./... -cover            # Coverage summary only
go test ./... -coverprofile=coverage.out
```

### Development
```bash
# Format code (always run before committing)
make fmt                        # go fmt ./...

# Run linter (requires golangci-lint)
make lint                       # Lints all packages

# Download/update dependencies
make deps                       # go mod download && go mod tidy

# Clean build artifacts
make clean                      # Removes binaries, coverage files, and output data
```

### Cross-compilation
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o confluence-reader-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o confluence-reader.exe

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o confluence-reader-macos-intel

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o confluence-reader-macos-arm
```

## Project Structure

```
confluence-reader/
├── main.go                          # Entry point - interactive CLI (98 lines)
├── go.mod                           # Module definition (Go 1.21+)
├── Makefile                         # Build automation
├── example-backup.sh                # Example automation script
│
├── pkg/
│   ├── client/                      # Confluence API client package
│   │   ├── client.go               # API implementation (306 lines)
│   │   └── client_test.go          # Tests with mock HTTP server
│   └── clone/                       # Clone orchestration package
│       ├── clone.go                # Clone logic (230 lines)
│       └── clone_test.go           # Utility tests
│
├── backups/                         # Default output location (not in repo)
├── confluence-data/                 # Alternative output location
│
└── Documentation/
    ├── README.md                    # User guide
    ├── USAGE.md                     # Usage examples
    ├── ARCHITECTURE.md              # Technical architecture (331 lines)
    ├── PROJECT_SUMMARY.md           # Project summary
    ├── QUICKREF.md                  # Quick reference
    └── .env.example                 # Environment variable template
```

## Architecture & Code Organization

### Three-Layer Architecture

1. **Presentation Layer** (`main.go`)
   - Interactive CLI prompts with input validation
   - Environment variable support (fallback to interactive)
   - Progress display and error reporting
   - Entry point that orchestrates client + cloner

2. **Client Layer** (`pkg/client/`)
   - HTTP communication with Confluence Cloud REST API v2
   - Authentication: HTTP Basic Auth (email + API token)
   - Automatic cursor-based pagination
   - 30-second timeout per request
   - Data models: `Space`, `Page`, `Attachment`, response wrappers

3. **Clone Layer** (`pkg/clone/`)
   - Business logic for cloning content
   - Directory structure management
   - Graceful error handling (log and continue)
   - Filename sanitization
   - JSON metadata persistence

### Key API Endpoints Used

| Endpoint | Purpose | Pagination |
|----------|---------|------------|
| `GET /wiki/api/v2/spaces` | List all accessible spaces | Cursor-based |
| `GET /wiki/api/v2/spaces/{id}/pages` | List pages in space | Cursor-based |
| `GET /wiki/api/v2/pages/{id}?body-format=storage` | Get page with HTML content | N/A |
| `GET /wiki/api/v2/pages/{id}/attachments` | List page attachments | Cursor-based |
| `GET {attachment.downloadLink.url}` | Download attachment binary | N/A |

**Base URL pattern:** `https://{domain}/wiki/api/v2`

## Code Patterns & Conventions

### Naming Conventions
- **Package names:** lowercase single word (`client`, `clone`)
- **Struct names:** PascalCase (`Client`, `Cloner`, `SpaceResponse`)
- **Methods:** camelCase with receiver prefix (`(c *Client) GetSpaces()`)
- **Private functions:** camelCase (`sanitizeFilename`, `saveJSON`)
- **Public APIs:** PascalCase (exported from packages)

### Error Handling Pattern
```go
// Three-level fault tolerance:
// 1. Critical: Exit immediately (auth failures, network errors)
if err := cloner.Clone(); err != nil {
    fmt.Printf("Error during clone: %v\n", err)
    os.Exit(1)
}

// 2. Space-level: Log warning and continue with next space
if err := cl.cloneSpace(space); err != nil {
    fmt.Printf("  Warning: Failed to clone space %s: %v\n", space.Key, err)
    continue
}

// 3. Page/Attachment-level: Log warning and continue (now concurrent)
// Pages are cloned concurrently with goroutines
go func(p client.Page) {
    defer wg.Done()
    if err := cl.clonePage(p, pagesDir); err != nil {
        mu.Lock()  // Protect console output
        fmt.Printf("    Warning: Failed to clone page %s: %v\n", p.Title, err)
        mu.Unlock()
    }
}(page)

// Wrap errors with context using %w for error chains
return fmt.Errorf("failed to create request: %w", err)
```

### Pagination Pattern
All list endpoints use cursor-based pagination:
```go
cursor := ""
for {
    params := url.Values{}
    params.Set("limit", "100")  // Max batch size
    if cursor != "" {
        params.Set("cursor", cursor)
    }
    
    // Make request, parse response
    
    // Check for next page
    if response.Links == nil || response.Links.Next == "" {
        break  // No more results
    }
    cursor = extractCursor(response.Links.Next)
}
```

### Concurrent Processing Pattern
Pages within a space are cloned concurrently:
```go
const maxConcurrent = 5
semaphore := make(chan struct{}, maxConcurrent)
var wg sync.WaitGroup
var mu sync.Mutex  // Protect console output

for _, page := range pages {
    wg.Add(1)
    go func(p client.Page) {
        defer wg.Done()
        
        // Acquire semaphore slot
        semaphore <- struct{}{}
        defer func() { <-semaphore }()
        
        // Clone page (thread-safe output)
        mu.Lock()
        fmt.Printf("Cloning: %s\n", p.Title)
        mu.Unlock()
        
        cl.clonePage(p, pagesDir)
    }(page)
}

wg.Wait()  // Wait for all pages to complete
```

### Directory Structure Output
```
{output-dir}/
├── {SPACE_KEY}/
│   ├── space.json                              # Space metadata
│   └── pages/
│       └── {PAGE_ID}_{Page_Title}/
│           ├── metadata.json                   # Page metadata
│           ├── content.html                    # HTML storage format
│           └── attachments/                    # Optional
│               ├── {filename}.{ext}            # Attachment file
│               └── {filename}.{ext}.json       # Attachment metadata
```

**Filename sanitization:**
- Removes invalid chars: `/ \ : * ? " < > |` → replaced with `_`
- Max length: 200 characters (truncated if longer)
- Trimmed whitespace

### Testing Patterns
- **Mock HTTP servers** using `httptest.NewServer` for API testing
- **Table-driven tests** for sanitization and utility functions
- **Verify authentication** in mock handlers (check BasicAuth)
- **Test pagination logic** by chaining multiple responses
- **Focus on critical paths** - client package has 64.6% coverage

Example test structure:
```go
func TestGetSpaces(t *testing.T) {
    // Create mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify authentication
        user, pass, ok := r.BasicAuth()
        if !ok || user != "user@example.com" || pass != "test-token" {
            t.Error("Expected basic auth credentials")
        }
        
        // Return mock response
        json.NewEncoder(w).Encode(SpaceResponse{...})
    }))
    defer server.Close()
    
    // Test client with mock server
    client := NewClient(...)
    // ... assertions
}
```

## Recent Changes

### Concurrent Page Cloning (Latest)
- **Added:** Concurrent page processing with semaphore pattern
- **Concurrency limit:** 5 pages at once (hardcoded in `clone.go`)
- **Thread-safety:** Mutex protects console output
- **Pattern:** Uses `sync.WaitGroup` + buffered channel semaphore
- **Performance:** 3-5x faster on spaces with many pages

### Attachment Download Fix (Latest)
- **Issue:** Confluence API inconsistently returns `downloadLink` as string OR object
- **Error:** `json: cannot unmarshal string into Go struct field`
- **Fix:** Custom `UnmarshalJSON` method tries both formats
- **Location:** `pkg/client/client.go` - `Attachment` struct
- **Pattern:** Uses `json.RawMessage` to defer parsing, tries string first, then object
- **Backwards compatible:** Handles both API response formats
- **Struct change:** Removed `Download *struct` field, added `DownloadURL string` field

## Important Gotchas

### 1. Go Version Requirement
- **Current go.mod:** Requires Go 1.21
- **Minimum version:** Go 1.21+ (per README)
- **Status:** Fixed - was incorrectly set to 1.25.4 (non-existent version)

### 2. Environment Variables
- All credentials **can** be provided via env vars (see `.env.example`)
- If env var is set, CLI will **not** prompt (shows "Using X from environment")
- Required env vars:
  - `CONFLUENCE_DOMAIN` (e.g., `company.atlassian.net`)
  - `CONFLUENCE_EMAIL` 
  - `CONFLUENCE_API_TOKEN` (create at https://id.atlassian.com/manage-profile/security/api-tokens)
  - `CONFLUENCE_OUTPUT_DIR` (optional, defaults to `./confluence-data`)

### 3. Authentication
- Uses **HTTP Basic Auth** with email + API token
- API token is NOT the password - must be generated separately
- Never log or persist credentials (security requirement)
- All communication over HTTPS only

### 4. Output Directory Behavior
- **Not cleaned automatically** - successive runs append/overwrite
- **Space directories use sanitized space keys** as folder names
- **Page directories use format:** `{PAGE_ID}_{sanitized_title}`
- For timestamped backups, pass custom output dir:
  ```bash
  CONFLUENCE_OUTPUT_DIR="./backups/$(date +%Y%m%d_%H%M%S)"
  ```

### 5. Error Recovery Strategy
- **Partial success is acceptable** - tool continues on failures
- Space clone failure → logs warning, continues to next space
- Page clone failure → logs warning, continues to next page
- Attachment failure → logs warning, continues to next attachment
- Only **critical failures** (auth, network) cause immediate exit
- Re-run to capture missed content (idempotent for existing content)

### 6. No .gitignore Currently
- Project **lacks .gitignore** file
- Should ignore:
  - Binary: `confluence-reader`, `*.exe`
  - Output: `confluence-data/`, `backups/`
  - Coverage: `coverage.out`, `coverage.html`
  - Env files: `.env`

### 7. Content Format
- Pages saved in **HTML storage format** (not rendered HTML)
- Storage format is Confluence's internal representation
- Contains Confluence-specific macros and markup
- For different formats, post-processing would be needed

### 8. Rate Limiting
- No built-in rate limit handling or exponential backoff
- Uses fixed 30-second timeout per request
- Batch size fixed at 100 items (API max)
- **Concurrent page processing** with semaphore (max 5 concurrent)
- If you hit 429 errors, tool will fail - no retry logic

### 9. Zero External Dependencies
- **Only uses Go standard library** - no `go.mod` dependencies
- Makes it very portable but limits functionality
- Any new features should prefer stdlib where possible
- If adding dependencies, justify carefully

## Security Considerations

### Credential Handling
- **Never log credentials** - no debug output of tokens/passwords
- **Never persist credentials** - not saved to disk
- **Env vars or interactive only** - no config files with credentials
- **HTTPS only** - all API calls encrypted
- **API tokens preferred** over passwords (per Atlassian best practices)

### File Permissions
- Directories created with `0755` (rwxr-xr-x)
- Files written with `0644` (rw-r--r--)
- No execution permissions on data files

### Input Validation
- Domain, email, API token all required (validation in main.go)
- Filenames sanitized to prevent path traversal
- No user input in HTTP headers (only pre-defined headers)

## Adding New Features

### Adding a New API Endpoint
1. **Define data model** in `pkg/client/client.go`
   ```go
   type NewResource struct {
       ID    string `json:"id"`
       Title string `json:"title"`
   }
   
   type NewResourceResponse struct {
       Results []NewResource `json:"results"`
       Links   *struct {
           Next string `json:"next"`
       } `json:"_links"`
   }
   ```

2. **Implement client method** with pagination if needed
   ```go
   func (c *Client) GetNewResources() ([]NewResource, error) {
       // Follow pagination pattern from GetSpaces()
   }
   ```

3. **Add test** in `pkg/client/client_test.go`
   ```go
   func TestGetNewResources(t *testing.T) {
       // Mock HTTP server pattern
   }
   ```

4. **Integrate in clone logic** if needed (`pkg/clone/clone.go`)
   ```go
   func (cl *Cloner) cloneNewResource(resource NewResource) error {
       // Save to filesystem
   }
   ```

### Modifying Output Structure
1. Update `clone/clone.go` save logic (directory creation, file writing)
2. Update `ARCHITECTURE.md` documentation
3. Update `README.md` if user-facing change
4. Consider backward compatibility with existing backups

### Adding CLI Flags
Currently uses prompts + env vars. To add flags:
1. Import `flag` package in `main.go`
2. Define flags before user prompts
3. Check flag values before prompting
4. Update `.env.example` and documentation
5. Consider maintaining backward compatibility

## Testing Guidelines

### Running Tests
```bash
# All tests
go test ./...

# Specific package
go test ./pkg/client -v
go test ./pkg/clone -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
```

### Writing New Tests
- **Use table-driven tests** for multiple input scenarios
- **Mock external dependencies** (HTTP servers, filesystem)
- **Test error cases** not just happy paths
- **Verify authentication** in API tests
- **Keep tests fast** - no actual network calls
- **Aim for >60% coverage** on new code

### Current Coverage
- `pkg/client`: 64.6% (good - critical API logic tested)
- `pkg/clone`: 7.4% (low - mostly integration/filesystem operations)
- Focus improvements on business logic and error paths

## Common Development Tasks

### Update API Version
1. Update `baseAPIPath` constant in `pkg/client/client.go`
2. Review Atlassian API changelog for breaking changes
3. Update data models if schema changed
4. Update/fix tests for new behavior
5. Update README and ARCHITECTURE with API version notes

### Add Filtering (e.g., by space key)
1. Add filter parameter to `main.go` (flag or prompt)
2. Pass filter to `Cloner`
3. Modify `Clone()` to filter spaces before cloning
4. Update documentation

### Add Progress Bar
1. Consider adding a dependency (e.g., `github.com/schollz/progressbar`)
   - Justify breaking zero-dependency rule
2. Or implement simple text-based progress (dots, percentages)
3. Track total items and completed count
4. Update in `clone/clone.go` progress messages

### Implement Incremental Sync
1. Create metadata file tracking last clone timestamp
2. Use API `last-modified` queries if available
3. Compare timestamps before downloading
4. Add "force full sync" flag for override
5. Requires more complex state management

## Documentation

### For Users
- **README.md** - Complete user guide (installation, features, usage)
- **USAGE.md** - Practical examples and use cases  
- **QUICKREF.md** - Quick reference for common tasks
- **.env.example** - Configuration template with comments

### For Developers  
- **ARCHITECTURE.md** - Deep technical documentation (2000+ words)
- **PROJECT_SUMMARY.md** - High-level overview and completion checklist
- **CRUSH.md** - This file (agent working guide)
- **Inline comments** - Code-level documentation

### When to Update Documentation
- **README**: User-facing features, requirements, output format changes
- **ARCHITECTURE**: Technical changes, new endpoints, data flow changes
- **CRUSH.md**: New patterns, gotchas, commands, or development workflows
- **USAGE.md**: New use cases or examples

## Performance Characteristics

### Current Behavior
- **Space processing** - sequential, one space at a time
- **Page processing** - concurrent with semaphore (max 5 pages at once)
- **Network bound** - speed depends on Confluence API response time
- **Batch size: 100** items per API call (Confluence API limit)
- **Memory: O(1)** per item - streaming approach, no buffering
- **Disk: O(n)** - proportional to total content size
- **Concurrency limit** - 5 concurrent page clones to avoid overwhelming API

### Optimization Opportunities
1. **Concurrent attachment downloads** - currently sequential within a page
2. **Connection pooling** - already uses `http.Client` pooling (good)
3. **Incremental sync** - only fetch changed content
4. **Compression** - gzip output directories for storage
5. **Rate limit backoff** - respect 429 responses with exponential backoff
6. **Tunable concurrency** - make max concurrent pages configurable (currently hardcoded to 5)

## Deployment

### Single Binary
- **No installation required** - copy binary and run
- **No runtime dependencies** - statically linked Go binary
- **Cross-platform** - compile for any Go-supported OS/arch
- **Size: ~7.9MB** (uncompressed)

### Automation
See `example-backup.sh` for cron job pattern:
```bash
#!/bin/bash
set -e

# Set credentials
export CONFLUENCE_DOMAIN="company.atlassian.net"
export CONFLUENCE_EMAIL="user@example.com"
export CONFLUENCE_API_TOKEN="token"
export CONFLUENCE_OUTPUT_DIR="./backups/$(date +%Y%m%d_%H%M%S)"

# Run clone
./confluence-reader

# Optional: compress output
tar -czf "${CONFLUENCE_OUTPUT_DIR}.tar.gz" "${CONFLUENCE_OUTPUT_DIR}"
rm -rf "${CONFLUENCE_OUTPUT_DIR}"
```

Add to crontab for scheduled backups:
```bash
# Daily at 2 AM
0 2 * * * /path/to/backup-script.sh >> /var/log/confluence-backup.log 2>&1
```

## Troubleshooting

### Build Issues
- **"go.mod requires go >= X"** - Update go.mod to match installed Go version
- **Import errors** - Run `go mod tidy` to sync dependencies
- **Build fails** - Check `go version` matches go.mod requirement

### Runtime Issues
- **401 Unauthorized** - Check API token is valid and not expired
- **403 Forbidden** - Check user has access to spaces/pages
- **404 Not Found** - Domain might be wrong, check full domain including `.atlassian.net`
- **Timeout errors** - Network issues or Confluence API slow, retry later
- **Rate limit (429)** - Too many requests, wait and retry (no automatic handling)

### Output Issues
- **Missing pages** - Check warnings in output, page may have failed
- **Empty content.html** - Page might be empty or restricted
- **Missing attachments** - Attachment download failed, check warnings
- **Weird filenames** - Long titles get truncated to 200 chars

### Testing Issues
- **Tests fail** - Run `go test ./... -v` for detailed output
- **Coverage issues** - Run `go test ./... -coverprofile=coverage.out` then `go tool cover -func=coverage.out`

## Future Enhancements (Ideas)

These are **not implemented** but documented as possibilities:

1. ~~**Concurrent processing** - Parallel downloads for speed~~ ✅ **IMPLEMENTED** (pages only, 5 concurrent)
2. **Incremental sync** - Track changes and only fetch updates
3. **Content filtering** - Include/exclude by space, date, author
4. **Format conversion** - Export to Markdown, PDF, etc.
5. **Search indexing** - Build local search index (e.g., with Bleve)
6. **Progress bars** - Visual indicators for long operations
7. **Resume support** - Checkpoint and resume interrupted clones
8. **Rate limit handling** - Exponential backoff for 429 responses
9. **Webhooks** - Real-time sync via Confluence webhooks
10. **Compression** - Automatic gzip of output
11. **Configuration file** - Alternative to env vars (be careful with credentials)
12. **Dry run mode** - Show what would be cloned without downloading

## Quick Reference Card

```bash
# Build
make build

# Test
make test
make coverage

# Run interactively
./confluence-reader

# Run automated (env vars)
CONFLUENCE_DOMAIN="..." CONFLUENCE_EMAIL="..." CONFLUENCE_API_TOKEN="..." ./confluence-reader

# Format & lint
make fmt
make lint

# Clean
make clean

# Cross-compile
GOOS=linux GOARCH=amd64 go build
GOOS=windows GOARCH=amd64 go build
```

---

**Last Updated:** 2025-01-11  
**Project Version:** 1.0  
**Go Version:** 1.21+  
**Recent Changes:**
- Added concurrent page cloning (5 workers)
- Fixed attachment downloadLink parsing issue
- Fixed go.mod version (was 1.25.4, now 1.21)
