#!/bin/bash
#
# Example script showing automated backup of Confluence content
# 
# Usage:
#   1. Create API token at https://id.atlassian.com/manage-profile/security/api-tokens
#   2. Set environment variables below
#   3. Run: ./example-backup.sh
#
# This script will:
#   - Create timestamped backup directory
#   - Clone all Confluence content
#   - Create a summary report
#

set -e  # Exit on error

# Configuration - REPLACE THESE VALUES
DOMAIN="yourcompany.atlassian.net"
EMAIL="your-email@example.com"
API_TOKEN="YOUR_API_TOKEN_HERE"

# Or use environment variables
DOMAIN="${CONFLUENCE_DOMAIN:-$DOMAIN}"
EMAIL="${CONFLUENCE_EMAIL:-$EMAIL}"
API_TOKEN="${CONFLUENCE_API_TOKEN:-$API_TOKEN}"

# Validate configuration
if [ "$DOMAIN" = "yourcompany.atlassian.net" ] || [ -z "$DOMAIN" ]; then
    echo "Error: Please set CONFLUENCE_DOMAIN"
    echo "Edit this script or set environment variable: export CONFLUENCE_DOMAIN='company.atlassian.net'"
    exit 1
fi

if [ "$EMAIL" = "your-email@example.com" ] || [ -z "$EMAIL" ]; then
    echo "Error: Please set CONFLUENCE_EMAIL"
    exit 1
fi

if [ "$API_TOKEN" = "YOUR_API_TOKEN_HERE" ] || [ -z "$API_TOKEN" ]; then
    echo "Error: Please set CONFLUENCE_API_TOKEN"
    echo "Get token at: https://id.atlassian.com/manage-profile/security/api-tokens"
    exit 1
fi

# Build if needed
if [ ! -f "./confluence-reader" ]; then
    echo "Building confluence-reader..."
    go build -o confluence-reader
fi

# Create backup directory with timestamp
BACKUP_DIR="./backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

echo "================================================"
echo "Confluence Backup Script"
echo "================================================"
echo "Domain: $DOMAIN"
echo "Email: $EMAIL"
echo "Output: $BACKUP_DIR"
echo "Started: $(date)"
echo "================================================"
echo

# Run the clone
CONFLUENCE_DOMAIN="$DOMAIN" \
CONFLUENCE_EMAIL="$EMAIL" \
CONFLUENCE_API_TOKEN="$API_TOKEN" \
CONFLUENCE_OUTPUT_DIR="$BACKUP_DIR" \
./confluence-reader

# Create summary report
REPORT_FILE="$BACKUP_DIR/backup-report.txt"
{
    echo "Confluence Backup Report"
    echo "========================"
    echo ""
    echo "Backup Date: $(date)"
    echo "Domain: $DOMAIN"
    echo "Output Directory: $BACKUP_DIR"
    echo ""
    echo "Statistics:"
    echo "-----------"
    
    # Count spaces
    SPACE_COUNT=$(find "$BACKUP_DIR" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')
    echo "Total Spaces: $SPACE_COUNT"
    
    # Count pages
    PAGE_COUNT=$(find "$BACKUP_DIR" -type f -name "metadata.json" | wc -l | tr -d ' ')
    echo "Total Pages: $PAGE_COUNT"
    
    # Count attachments
    ATTACHMENT_COUNT=$(find "$BACKUP_DIR" -type d -name "attachments" -exec sh -c 'find "$1" -type f ! -name "*.json" | wc -l' _ {} \; | awk '{s+=$1} END {print s}')
    echo "Total Attachments: ${ATTACHMENT_COUNT:-0}"
    
    # Calculate total size
    TOTAL_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
    echo "Total Size: $TOTAL_SIZE"
    
    echo ""
    echo "Backup completed successfully!"
    
} > "$REPORT_FILE"

echo
echo "================================================"
echo "Backup completed successfully!"
echo "================================================"
cat "$REPORT_FILE"
echo
echo "Backup location: $BACKUP_DIR"
echo "Report saved to: $REPORT_FILE"
