# Architecture Documentation

## Overview

The Confluence Reader is a Go-based CLI application that clones content from Atlassian Confluence Cloud using the official REST API v2. It's designed with a clean, modular architecture for maintainability and testability.

## Project Structure

```
confluence-reader/
├── main.go                          # Entry point - interactive CLI
├── go.mod                           # Go module definition
├── README.md                        # User documentation
├── USAGE.md                         # Usage examples
├── ARCHITECTURE.md                  # This file
├── Makefile                         # Build automation
├── .gitignore                       # Git ignore rules
├── .env.example                     # Environment variable template
└── pkg/                             # Package directory
    ├── client/                      # API client package
    │   ├── client.go                # Confluence API client
    │   └── client_test.go           # Client tests
    └── clone/                       # Cloning logic package
        ├── clone.go                 # Clone orchestration
        └── clone_test.go            # Clone tests
```

## Architecture Layers

### 1. Presentation Layer (main.go)

**Responsibility**: User interaction and input collection

- Interactive prompts for configuration
- Environment variable support
- Input validation
- Progress display
- Error reporting

**Key Functions**:
- `main()`: Orchestrates the entire flow
  1. Collects user input (with env fallback)
  2. Validates required parameters
  3. Initializes client and cloner
  4. Executes clone operation
  5. Reports results

### 2. Client Layer (pkg/client)

**Responsibility**: HTTP communication with Confluence API

**Key Components**:

- `Client`: Main API client struct
  - Handles authentication (HTTP Basic Auth)
  - Manages HTTP connections
  - Implements retry logic (via HTTP timeout)
  - Abstracts API communication

**Data Models**:
- `Space`: Represents a Confluence space
- `Page`: Represents a Confluence page
- `Attachment`: Represents a page attachment
- `*Response`: Wrapper types for API responses with pagination

**Key Methods**:
- `NewClient(domain, email, apiToken)`: Constructor
- `doRequest(method, path, params)`: Low-level HTTP request handler
- `GetSpaces()`: Retrieves all spaces with pagination
- `GetSpacePages(spaceID)`: Retrieves all pages in a space
- `GetPage(pageID)`: Retrieves single page with full content
- `GetPageAttachments(pageID)`: Retrieves page attachments
- `DownloadAttachment(url)`: Downloads attachment binary data

**Design Decisions**:
- Cursor-based pagination handled automatically
- Basic Auth for simplicity and security
- 30-second timeout to prevent hanging
- JSON unmarshaling for type safety

### 3. Clone Layer (pkg/clone)

**Responsibility**: Business logic for cloning content

**Key Components**:

- `Cloner`: Orchestrates the cloning process
  - Manages directory structure
  - Coordinates API calls
  - Handles errors gracefully
  - Saves content to filesystem

**Key Methods**:
- `NewCloner(client, outputDir)`: Constructor
- `Clone()`: Main entry point for cloning
- `cloneSpace(space)`: Clones a single space
- `clonePage(page, dir)`: Clones a single page
- `downloadAttachment(attachment, dir)`: Downloads and saves attachment
- `sanitizeFilename(name)`: Sanitizes filenames for filesystem
- `saveJSON(path, data)`: Helper for saving JSON metadata

**Design Decisions**:
- Hierarchical directory structure mirrors Confluence organization
- Separate metadata files for searchability
- HTML storage format for content fidelity
- Graceful error handling (log and continue)
- Progress reporting for user feedback

## Data Flow

```
User Input → main.go
    ↓
Client Creation → client.NewClient()
    ↓
Cloner Creation → clone.NewCloner()
    ↓
Clone Execution → cloner.Clone()
    ↓
    ├─→ GetSpaces() → [Space, Space, ...]
    │       ↓
    │   For each Space:
    │       ├─→ Save space.json
    │       ├─→ GetSpacePages(spaceID) → [Page, Page, ...]
    │       │       ↓
    │       │   For each Page:
    │       │       ├─→ GetPage(pageID) → Page (with content)
    │       │       ├─→ Save metadata.json
    │       │       ├─→ Save content.html
    │       │       ├─→ GetPageAttachments(pageID) → [Attachment, ...]
    │       │       │       ↓
    │       │       │   For each Attachment:
    │       │       │       ├─→ DownloadAttachment(url) → binary data
    │       │       │       ├─→ Save attachment file
    │       │       │       └─→ Save attachment.json
    │       │       └─→ Continue
    │       └─→ Continue
    └─→ Complete
```

## Directory Structure Output

```
confluence-data/
├── SPACE_KEY/
│   ├── space.json                              # Space metadata
│   └── pages/
│       └── PAGE_ID_Page_Title/
│           ├── metadata.json                   # Page metadata
│           ├── content.html                    # Page HTML content
│           └── attachments/                    # Optional
│               ├── filename.ext                # Attachment file
│               └── filename.ext.json           # Attachment metadata
```

## API Integration

### Authentication
- **Method**: HTTP Basic Auth
- **Credentials**: Email + API Token
- **Security**: HTTPS only, credentials never logged

### Endpoints Used

1. **GET /wiki/api/v2/spaces**
   - Purpose: List all accessible spaces
   - Pagination: Cursor-based
   - Parameters: `limit=100`, `cursor`

2. **GET /wiki/api/v2/spaces/{id}/pages**
   - Purpose: List pages in a space
   - Pagination: Cursor-based
   - Parameters: `limit=100`, `cursor`

3. **GET /wiki/api/v2/pages/{id}**
   - Purpose: Get full page content
   - Parameters: `body-format=storage`
   - Returns: Complete page with HTML storage format

4. **GET /wiki/api/v2/pages/{id}/attachments**
   - Purpose: List page attachments
   - Pagination: Cursor-based
   - Parameters: `limit=100`, `cursor`

5. **GET {attachment.downloadLink.url}**
   - Purpose: Download attachment binary
   - Returns: Raw file data

### Pagination Strategy

All list endpoints use cursor-based pagination:
- Request with `limit=100` for batch size
- Check `_links.next` in response
- Extract `cursor` parameter from next URL
- Continue until no more results

## Error Handling Strategy

### Levels of Fault Tolerance

1. **Space Level**: If space clone fails → Log warning, continue with other spaces
2. **Page Level**: If page clone fails → Log warning, continue with other pages
3. **Attachment Level**: If attachment fails → Log warning, continue with other attachments
4. **Critical Failures**: Network errors, auth failures → Exit with error

### Rationale
- Maximize data retrieval even with partial failures
- Provide visibility into what succeeded/failed
- Allow re-runs to capture missed content

## Testing Strategy

### Unit Tests

**Client Package** (`client_test.go`):
- Mock HTTP server for API responses
- Test authentication headers
- Test pagination logic
- Test error handling
- Test data unmarshaling

**Clone Package** (`clone_test.go`):
- Test filename sanitization
- Test long filename truncation
- Test invalid character replacement

### Test Coverage
- Focus on critical paths
- Mock external dependencies (HTTP)
- Validate data transformations
- Ensure error conditions are handled

## Security Considerations

1. **Credentials**: Never logged or written to disk
2. **HTTPS**: All API communication encrypted
3. **API Tokens**: Preferred over passwords
4. **.gitignore**: Prevents accidental credential commits
5. **Environment Variables**: Secure alternative to interactive input

## Performance Characteristics

### Throughput
- Batch size: 100 items per API call (API limit)
- Concurrent: Sequential processing (respects rate limits)
- Network bound: Speed depends on API response time

### Optimizations
- Batch API calls with max page size
- Reuse HTTP client connection pooling
- Stream attachment downloads (no memory buffering)

### Scalability
- Memory: O(1) per item (streaming approach)
- Disk: O(n) where n = total content size
- Time: O(n) where n = number of API calls

## Future Enhancements

Possible improvements:
1. **Incremental Sync**: Track cloned content, only fetch changes
2. **Concurrent Downloads**: Parallel goroutines for attachments
3. **Filtering**: Options to include/exclude spaces or content types
4. **Export Formats**: Convert to Markdown, PDF, etc.
5. **Progress Bar**: Visual progress indicator
6. **Resume Capability**: Checkpoint and resume interrupted clones
7. **Rate Limit Handling**: Exponential backoff for 429 responses
8. **Webhooks**: Real-time sync via Confluence webhooks

## Dependencies

### Standard Library
- `encoding/json`: JSON parsing
- `net/http`: HTTP client
- `net/url`: URL parsing
- `io`: Stream operations
- `os`: File system operations
- `fmt`: Formatting
- `strings`: String utilities
- `time`: Timeouts

### External
- None (pure Go, zero dependencies)

## Build and Deployment

### Build
```bash
go build -o confluence-reader
```

### Cross-compilation
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o confluence-reader-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o confluence-reader.exe

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o confluence-reader-macos-intel

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o confluence-reader-macos-arm
```

### Deployment
- Single binary, no dependencies
- No installation required
- Portable across platforms
- No configuration files needed (uses env vars or prompts)

## Maintenance Guidelines

### Adding New Endpoints
1. Define data model in `client/client.go`
2. Implement API method in `Client`
3. Add test in `client_test.go`
4. Integrate in `clone/clone.go` if needed

### Modifying Output Structure
1. Update `clone/clone.go` save logic
2. Update `ARCHITECTURE.md` documentation
3. Update `README.md` if user-facing

### Updating API Version
1. Update `baseAPIPath` constant in `client.go`
2. Review API changes in Atlassian docs
3. Update data models if schema changed
4. Update tests for new behavior
5. Update README with new API details
