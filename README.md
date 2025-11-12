# Confluence Reader

An interactive Go client for cloning Confluence content from Atlassian Cloud using the official Confluence Cloud REST API v2.

## Features

- **Interactive CLI**: Easy-to-use command-line interface
- **Complete Clone**: Downloads all spaces, pages, and attachments
- **Organized Structure**: Saves content in a logical directory hierarchy
- **Metadata Preservation**: Stores page metadata, versions, and attachment info
- **Content Format**: Saves page content in HTML storage format
- **Markdown Export**: Optional conversion to LLM-friendly Markdown with YAML frontmatter
- **Progress Tracking**: Real-time progress updates during the clone process
- **Concurrent Processing**: Fast page downloads with concurrent workers

## Prerequisites

- Go 1.21 or later
- Atlassian Cloud account with Confluence access
- API token (create one at https://id.atlassian.com/manage-profile/security/api-tokens)

## Installation

```bash
git clone <repository-url>
cd confluence-reader
go build -o confluence-reader
```

## Usage

Run the interactive client:

```bash
./confluence-reader
```

You will be prompted to enter:
1. **Confluence domain**: Your Atlassian domain (e.g., `yourcompany.atlassian.net`)
2. **Email**: Your Atlassian account email
3. **API token**: Your Atlassian API token
4. **Output directory**: Where to save the cloned content (default: `./confluence-data`)

### Environment Variables

Alternatively, you can use environment variables to avoid interactive prompts:

```bash
CONFLUENCE_DOMAIN="yourcompany.atlassian.net" \
CONFLUENCE_EMAIL="your-email@example.com" \
CONFLUENCE_API_TOKEN="your-api-token" \
CONFLUENCE_OUTPUT_DIR="./confluence-data" \
./confluence-reader
```

### Markdown Export (Optional)

Enable markdown export to convert Confluence pages to LLM-friendly Markdown format:

```bash
CONFLUENCE_EXPORT_MARKDOWN=true ./confluence-reader
```

Or with all environment variables:

```bash
CONFLUENCE_DOMAIN="yourcompany.atlassian.net" \
CONFLUENCE_EMAIL="your-email@example.com" \
CONFLUENCE_API_TOKEN="your-api-token" \
CONFLUENCE_OUTPUT_DIR="./confluence-data" \
CONFLUENCE_EXPORT_MARKDOWN=true \
./confluence-reader
```

When markdown export is enabled:
- Each page is saved as both `content.html` (original) and `content.md` (converted)
- Markdown files include YAML frontmatter with page metadata
- Clean, readable format suitable for feeding to LLMs like ChatGPT or Claude
- Git-friendly format for tracking documentation changes

## Output Structure

The tool creates the following directory structure:

```
confluence-data/
├── SPACE_KEY_1/
│   ├── space.json                    # Space metadata
│   └── pages/
│       ├── PAGE_ID_1_Page_Title/
│       │   ├── metadata.json         # Page metadata
│       │   ├── content.html          # Page content (storage format)
│       │   ├── content.md            # Markdown conversion (if enabled)
│       │   └── attachments/          # Page attachments (if any)
│       │       ├── file1.pdf
│       │       ├── file1.pdf.json    # Attachment metadata
│       │       └── ...
│       └── ...
├── SPACE_KEY_2/
│   └── ...
└── ...
```

### File Contents

- **space.json**: Contains space ID, key, name, type, status, and description
- **metadata.json**: Contains page ID, title, status, space ID, parent ID, and version info
- **content.html**: Page content in Confluence storage format (HTML)
- **content.md**: Markdown conversion with YAML frontmatter (if markdown export enabled)
- **attachments/**: Directory containing all page attachments with their metadata

### Markdown Format

When `CONFLUENCE_EXPORT_MARKDOWN=true`, each page includes a `content.md` file:

```markdown
---
title: "Getting Started Guide"
confluence_id: "123456"
space_key: "DOC"
version: 5
author: "user@example.com"
parent_id: "789"
url: "https://yourcompany.atlassian.net/wiki/spaces/DOC/pages/123456"
---

# Getting Started Guide

Welcome to the documentation...

## Prerequisites

- Item 1
- Item 2

## Code Example

```python
def hello():
    print("Hello, World!")
```

## Related Pages

- [Another Page](another-page.md)
```

**Markdown Features**:
- ✅ Headings, paragraphs, lists, tables
- ✅ Bold, italic, links, images
- ✅ Code blocks with syntax highlighting
- ✅ Confluence macros converted to Markdown equivalents
- ✅ Warning/Info panels → Blockquotes with emoji (⚠️, ℹ️, 📝)
- ✅ Internal page links → Relative markdown links
- ✅ TOC macros removed (redundant in Markdown)

**LLM Usage**:

Feed markdown files to ChatGPT, Claude, or other LLMs:

```bash
# Copy to clipboard (macOS)
cat content.md | pbcopy

# Or use with LLM APIs
curl https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"gpt-4\",
    \"messages\": [{
      \"role\": \"user\",
      \"content\": \"$(cat content.md)\n\nSummarize this documentation.\"
    }]
  }"
```

## API Details

This tool uses the Confluence Cloud REST API v2:
- Base endpoint: `https://{domain}/wiki/api/v2`
- Authentication: HTTP Basic Auth (email + API token)
- Pagination: Cursor-based with automatic handling
- Rate limiting: Respects API rate limits with proper error handling

## Security Notes

- API tokens are sensitive credentials - never commit them to version control
- The tool uses HTTPS for all API communications
- Basic authentication is used as recommended by Atlassian for API tokens

## Error Handling

The tool includes robust error handling:
- Failed space clones are logged but don't stop the overall process
- Failed page clones are logged and skipped
- Failed attachment downloads are logged and skipped
- Failed markdown conversions are logged but HTML is still saved
- Network errors and API errors are reported with detailed messages

## Use Cases

- **Documentation Backup**: Create local backups of your Confluence documentation
- **LLM Integration**: Convert documentation to Markdown for ChatGPT, Claude, etc.
- **Git Tracking**: Track documentation changes over time with git-friendly Markdown
- **Offline Access**: Read Confluence content without internet connection
- **Migration**: Prepare content for migration to other platforms
- **Search**: Use local tools like `grep` to search across all documentation

## Performance

- **Concurrent Processing**: Downloads 5 pages simultaneously for faster cloning
- **Markdown Conversion**: ~2ms per page (negligible overhead)
- **Typical Performance**: 100-page space clones in under 1 minute

## Troubleshooting

### Authentication Issues
- Verify your API token is valid and not expired
- Ensure you're using the correct Atlassian account email
- Check that your domain includes `.atlassian.net`

### Missing Content
- Check console output for warnings about failed pages
- Verify you have access to the spaces/pages in Confluence
- Some pages may be restricted or archived

### Markdown Export Issues
- Markdown conversion failures are logged but don't stop the clone
- HTML content is always saved as fallback
- Complex tables or macros may not convert perfectly (graceful degradation)

## License

MIT License
