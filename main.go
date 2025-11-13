package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nycmonkey/confluence-reader/pkg/client"
	"github.com/nycmonkey/confluence-reader/pkg/clone"
)

func main() {
	fmt.Println("Confluence Content Cloner")
	fmt.Println("========================")
	fmt.Println()

	// Check for environment variables first
	domain := os.Getenv("CONFLUENCE_DOMAIN")
	email := os.Getenv("CONFLUENCE_EMAIL")
	apiToken := os.Getenv("CONFLUENCE_API_TOKEN")
	outputDir := os.Getenv("CONFLUENCE_OUTPUT_DIR")
	exportMarkdown := os.Getenv("CONFLUENCE_EXPORT_MARKDOWN")
	sampleSpacesStr := os.Getenv("CONFLUENCE_SAMPLE_SPACES")
	samplePagesStr := os.Getenv("CONFLUENCE_SAMPLE_PAGES")

	// Parse sampling values
	sampleSpaces, err := strconv.Atoi(sampleSpacesStr)
	if err != nil {
		sampleSpaces = 0
	}
	samplePages, err := strconv.Atoi(samplePagesStr)
	if err != nil {
		samplePages = 0
	}

	scanner := bufio.NewScanner(os.Stdin)

	// Get Confluence domain
	if domain == "" {
		fmt.Print("Enter your Confluence domain (e.g., yourcompany.atlassian.net): ")
		scanner.Scan()
		domain = strings.TrimSpace(scanner.Text())
	} else {
		fmt.Printf("Using domain from environment: %s\n", domain)
	}
	if domain == "" {
		fmt.Println("Error: Domain is required")
		os.Exit(1)
	}

	// Get email
	if email == "" {
		fmt.Print("Enter your Atlassian account email: ")
		scanner.Scan()
		email = strings.TrimSpace(scanner.Text())
	} else {
		fmt.Printf("Using email from environment: %s\n", email)
	}
	if email == "" {
		fmt.Println("Error: Email is required")
		os.Exit(1)
	}

	// Get API token
	if apiToken == "" {
		fmt.Print("Enter your API token: ")
		scanner.Scan()
		apiToken = strings.TrimSpace(scanner.Text())
	} else {
		fmt.Println("Using API token from environment")
	}
	if apiToken == "" {
		fmt.Println("Error: API token is required")
		os.Exit(1)
	}

	// Get output directory
	if outputDir == "" {
		fmt.Print("Enter output directory (default: ./confluence-data): ")
		scanner.Scan()
		outputDir = strings.TrimSpace(scanner.Text())
		if outputDir == "" {
			outputDir = "./confluence-data"
		}
	} else {
		fmt.Printf("Using output directory from environment: %s\n", outputDir)
	}

	// Check markdown export flag
	if exportMarkdown == "true" {
		fmt.Println("Markdown export enabled")
	}

	fmt.Println()
	fmt.Println("Initializing Confluence client...")

	// Create client
	c := client.NewClient(domain, email, apiToken)

	// Create cloner
	cloner := clone.NewCloner(c, outputDir, sampleSpaces, samplePages)

	// Enable markdown export if requested
	if exportMarkdown == "true" {
		cloner.EnableMarkdownExport(domain)
	}

	// Start cloning
	fmt.Println("Starting clone process...")
	fmt.Println()

	if err := cloner.Clone(); err != nil {
		fmt.Printf("Error during clone: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Clone completed successfully!")
	fmt.Printf("Content saved to: %s\n", outputDir)
}
