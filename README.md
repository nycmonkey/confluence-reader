# Confluence Reader

An interactive Go client for cloning Confluence content from Atlassian Cloud using the official Confluence Cloud REST API v2.

## Features

- **Interactive CLI**: Easy-to-use command-line interface
- **Complete Clone**: Downloads all spaces, pages, and attachments
- **Organized Structure**: Saves content in a logical directory hierarchy
- **Metadata Preservation**: Stores page metadata, versions, and attachment info
- **Content Format**: Saves page content in HTML storage format
- **Progress Tracking**: Real-time progress updates during the clone process

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
- **attachments/**: Directory containing all page attachments with their metadata

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
- Network errors and API errors are reported with detailed messages

## License

MIT License
