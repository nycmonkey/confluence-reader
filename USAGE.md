# Confluence Content Cloner - Example Usage

## Quick Start

1. Create an API token:
   - Go to https://id.atlassian.com/manage-profile/security/api-tokens
   - Click "Create API token"
   - Give it a name (e.g., "Confluence Cloner")
   - Copy the token (you won't be able to see it again)

2. Run the tool:
   ```bash
   ./confluence-reader
   ```

3. Enter your credentials when prompted:
   ```
   Confluence Content Cloner
   ========================

   Enter your Confluence domain (e.g., yourcompany.atlassian.net): mycompany.atlassian.net
   Enter your Atlassian account email: user@example.com
   Enter your API token: **************************
   Enter output directory (default: ./confluence-data): ./my-confluence-backup
   ```

## Example Output

```
Initializing Confluence client...
Starting clone process...

Fetching spaces...
Found 3 space(s)

[1/3] Processing space: Engineering (ENG)
  Fetching pages...
  Found 15 page(s)
  [1/15] Cloning page: Architecture Overview
  [2/15] Cloning page: API Documentation
    Found 2 attachment(s)
    [1/2] Downloading: api-diagram.png
    [2/2] Downloading: swagger.json
  [3/15] Cloning page: Deployment Guide
  ...

[2/3] Processing space: Product (PROD)
  Fetching pages...
  Found 8 page(s)
  ...

[3/3] Processing space: Marketing (MKT)
  Fetching pages...
  Found 5 page(s)
  ...

Clone completed successfully!
Content saved to: ./my-confluence-backup
```

## What Gets Cloned

For each space:
- ✅ Space metadata (name, key, description, type, status)
- ✅ All pages in the space
- ✅ Page content in HTML storage format
- ✅ Page metadata (title, version, parent relationships)
- ✅ All attachments for each page
- ✅ Attachment metadata

## Use Cases

1. **Backup**: Create a local backup of your Confluence content
2. **Migration**: Prepare content for migration to another system
3. **Offline Access**: Access Confluence content without internet connection
4. **Analysis**: Analyze content structure and relationships
5. **Archival**: Archive old spaces before decommissioning

## Tips

- The tool handles pagination automatically, so it works with large Confluence instances
- Progress is shown in real-time for long-running operations
- If a single page or attachment fails, the tool continues with the rest
- Content is saved in the native Confluence storage format (HTML) for maximum fidelity
