# Quick Reference Guide

## Installation

```bash
# Clone and build
git clone <repo-url>
cd confluence-reader
make build

# Or just
go build -o confluence-reader
```

## Basic Usage

### Interactive Mode
```bash
./confluence-reader
```
Follow the prompts.

### Environment Variables Mode
```bash
export CONFLUENCE_DOMAIN="mycompany.atlassian.net"
export CONFLUENCE_EMAIL="user@example.com"
export CONFLUENCE_API_TOKEN="ATATT3xFfGF0..."
export CONFLUENCE_OUTPUT_DIR="./backup"
./confluence-reader
```

### One-liner
```bash
CONFLUENCE_DOMAIN="mycompany.atlassian.net" \
CONFLUENCE_EMAIL="user@example.com" \
CONFLUENCE_API_TOKEN="ATATT3xFfGF0..." \
./confluence-reader
```

## Getting API Token

1. Visit: https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Name it (e.g., "Confluence Reader")
4. Copy and save securely

## Makefile Commands

```bash
make build      # Build binary
make test       # Run tests
make coverage   # Generate coverage report
make clean      # Remove artifacts
make run        # Build and run
make fmt        # Format code
make deps       # Download dependencies
make help       # Show all commands
```

## Output Structure

```
confluence-data/
├── ENGINEERING/                     # Space directory (by key)
│   ├── space.json                   # Space metadata
│   └── pages/
│       └── 123456_Architecture/     # Page directory (id_title)
│           ├── metadata.json        # Page info
│           ├── content.html         # Page content
│           └── attachments/         # If page has attachments
│               ├── diagram.png
│               └── diagram.png.json
```

## File Formats

### space.json
```json
{
  "id": "123456",
  "key": "ENG",
  "name": "Engineering",
  "type": "global",
  "status": "current",
  "description": "Engineering documentation"
}
```

### metadata.json
```json
{
  "id": "789012",
  "title": "Architecture Overview",
  "status": "current",
  "spaceId": "123456",
  "parentId": "345678",
  "version": {
    "number": 5,
    "createdAt": "2024-01-15T10:30:00.000Z"
  }
}
```

### content.html
Raw Confluence storage format (HTML):
```html
<ac:structured-macro ac:name="info">
  <ac:rich-text-body>
    <p>This is Confluence content...</p>
  </ac:rich-text-body>
</ac:structured-macro>
```

## Troubleshooting

### Authentication Errors
- Verify domain format: `company.atlassian.net` (no https://)
- Check email is correct
- Regenerate API token if needed
- Ensure account has Confluence access

### Rate Limiting
- Tool respects API rate limits
- If you hit limits, wait and retry
- Consider smaller batches for large instances

### Partial Failures
- Tool logs warnings but continues
- Check output for "Warning:" messages
- Re-run to capture missed content

### Permission Errors
- Ensure output directory is writable
- Check disk space available
- Verify no file system restrictions

## API Rate Limits

Atlassian Cloud has rate limits:
- **Authenticated requests**: ~200 per minute per IP
- Tool handles this with timeouts
- Large instances may take significant time

## Tips

1. **Test First**: Try with small space before full clone
2. **Disk Space**: Check available space (can be GBs for large instances)
3. **Network**: Stable connection required for large clones
4. **Security**: Never commit API tokens to version control
5. **Backups**: Regular clones serve as good backups

## Common Use Cases

### Regular Backups
```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d)
CONFLUENCE_DOMAIN="company.atlassian.net" \
CONFLUENCE_EMAIL="backup@company.com" \
CONFLUENCE_API_TOKEN="$CONFLUENCE_TOKEN" \
CONFLUENCE_OUTPUT_DIR="./backups/$DATE" \
./confluence-reader
```

### Migration Preparation
1. Clone entire Confluence instance
2. Analyze structure in filesystem
3. Transform as needed for target system
4. Import to new platform

### Offline Documentation
1. Clone content
2. Serve with local web server:
```bash
cd confluence-data
python3 -m http.server 8000
# Visit http://localhost:8000
```

## Development

### Project Structure
```
├── main.go              # CLI entry point
├── pkg/client/          # API client
├── pkg/clone/           # Clone logic
├── README.md           # User docs
├── USAGE.md            # Examples
├── ARCHITECTURE.md     # Technical docs
└── Makefile            # Build automation
```

### Running Tests
```bash
go test ./...
go test ./... -v                    # Verbose
go test ./... -cover               # With coverage
go test ./pkg/client -run TestAPI  # Specific test
```

### Adding Features
1. Update relevant package (client or clone)
2. Add tests
3. Update documentation
4. Run `make test` and `make build`
5. Update CHANGELOG

## Support

- **Issues**: Check GitHub issues
- **API Docs**: https://developer.atlassian.com/cloud/confluence/rest/v2/
- **API Tokens**: https://id.atlassian.com/manage-profile/security/api-tokens

## License

MIT License - see LICENSE file
