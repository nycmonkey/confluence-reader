package markdown

import (
	"fmt"
	htmlpkg "html"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// Converter handles HTML to Markdown conversion with Confluence-specific support
type Converter struct {
	// No internal state needed - using functional API
}

// PageMetadata contains metadata for a Confluence page
type PageMetadata struct {
	Title     string
	PageID    string
	SpaceKey  string
	Version   int
	UpdatedAt time.Time
	Author    string
	ParentID  string
	URL       string
}

// NewConverter creates a new converter with Confluence-specific configuration
func NewConverter() *Converter {
	return &Converter{}
}

// Convert converts Confluence storage HTML to Markdown
func (c *Converter) Convert(html string) (string, error) {
	// Pre-process: Clean up Confluence-specific elements
	html = preProcess(html)

	// Convert to Markdown using default settings
	markdown, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("conversion failed: %w", err)
	}

	// Post-process: Clean up output
	markdown = postProcess(markdown)

	return markdown, nil
}

// ConvertWithMetadata converts HTML to Markdown with YAML frontmatter
func (c *Converter) ConvertWithMetadata(html string, meta PageMetadata) (string, error) {
	// Generate YAML frontmatter
	frontmatter := generateFrontmatter(meta)

	// Convert HTML to markdown
	markdown, err := c.Convert(html)
	if err != nil {
		return "", err
	}

	// Combine frontmatter + markdown
	return frontmatter + "\n\n" + markdown, nil
}

// generateFrontmatter creates YAML frontmatter from page metadata
func generateFrontmatter(meta PageMetadata) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	
	sb.WriteString(fmt.Sprintf("title: \"%s\"\n", escapeYAML(meta.Title)))
	sb.WriteString(fmt.Sprintf("confluence_id: \"%s\"\n", meta.PageID))
	sb.WriteString(fmt.Sprintf("space_key: \"%s\"\n", meta.SpaceKey))
	sb.WriteString(fmt.Sprintf("version: %d\n", meta.Version))
	
	if !meta.UpdatedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("last_updated: \"%s\"\n", meta.UpdatedAt.Format(time.RFC3339)))
	}
	
	if meta.Author != "" {
		sb.WriteString(fmt.Sprintf("author: \"%s\"\n", escapeYAML(meta.Author)))
	}
	
	if meta.ParentID != "" {
		sb.WriteString(fmt.Sprintf("parent_id: \"%s\"\n", meta.ParentID))
	}
	
	if meta.URL != "" {
		sb.WriteString(fmt.Sprintf("url: \"%s\"\n", meta.URL))
	}
	
	sb.WriteString("---")
	return sb.String()
}

// escapeYAML escapes quotes in YAML string values
func escapeYAML(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// preProcess cleans up Confluence-specific HTML before conversion
func preProcess(html string) string {
	// Remove TOC macros (redundant in Markdown)
	html = removeTOCMacros(html)

	// Convert Confluence emoticons to Unicode emoji
	html = convertEmoticons(html)

	// Convert Confluence code macros to standard pre/code
	html = convertCodeMacros(html)

	// Convert Confluence warning/info panels to blockquotes
	html = convertPanelMacros(html)

	// Convert Confluence internal links to standard links
	html = convertInternalLinks(html)

	// Remove child pages macro (TODO: needs page hierarchy context)
	html = removeChildrenMacro(html)

	return html
}

// removeTOCMacros removes Confluence TOC macros
func removeTOCMacros(html string) string {
	// Pattern: <ac:structured-macro ac:name="toc" ... />
	re := regexp.MustCompile(`<ac:structured-macro\s+ac:name="toc"[^>]*/>`)
	html = re.ReplaceAllString(html, "")

	// Also remove wrapping paragraphs if they're now empty
	html = strings.ReplaceAll(html, "<p></p>", "")

	return html
}

// convertEmoticons converts Confluence emoticon tags to Unicode emoji
func convertEmoticons(html string) string {
	// Map Confluence emoticon names to Unicode emoji
	emoticons := map[string]string{
		"smile":         "😊",
		"sad":           "😞",
		"cheeky":        "😜",
		"laugh":         "😆",
		"wink":          "😉",
		"thumbs-up":     "👍",
		"thumbs-down":   "👎",
		"tick":          "✅",
		"cross":         "❌",
		"warning":       "⚠️",
		"information":   "ℹ️",
		"tick-box":      "☑️",
		"question":      "❓",
		"light-on":      "💡",
		"light-off":     "🔦",
		"star":          "⭐",
		"heart":         "❤️",
		"plus":          "➕",
		"minus":         "➖",
		"flag":          "🚩",
	}

	// Pattern: <ac:emoticon ac:name="emoticon_name" />
	re := regexp.MustCompile(`<ac:emoticon\s+ac:name="([^"]+)"\s*/>`)

	html = re.ReplaceAllStringFunc(html, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) > 1 {
			name := matches[1]
			if emoji, ok := emoticons[name]; ok {
				return emoji
			}
			// Unknown emoticon - return text placeholder
			return fmt.Sprintf(":%s:", name)
		}
		return match
	})

	return html
}

// convertCodeMacros converts Confluence code macros to HTML pre/code blocks
func convertCodeMacros(html string) string {
	// Pattern: <ac:structured-macro ac:name="code">...</ac:structured-macro>
	re := regexp.MustCompile(`(?s)<ac:structured-macro\s+ac:name="code"[^>]*>(.*?)</ac:structured-macro>`)

	html = re.ReplaceAllStringFunc(html, func(match string) string {
		// Extract language
		langRe := regexp.MustCompile(`<ac:parameter\s+ac:name="language">([^<]+)</ac:parameter>`)
		langMatch := langRe.FindStringSubmatch(match)
		lang := ""
		if len(langMatch) > 1 {
			lang = langMatch[1]
		}

		// Extract code content
		codeRe := regexp.MustCompile(`(?s)<ac:plain-text-body><!\[CDATA\[(.*?)\]\]></ac:plain-text-body>`)
		codeMatch := codeRe.FindStringSubmatch(match)
		code := ""
		if len(codeMatch) > 1 {
			code = codeMatch[1]
		}

		// Escape HTML entities in code (prevents html-to-markdown from parsing as HTML)
		codeEscaped := htmlpkg.EscapeString(code)

		// Return as HTML pre/code with language class
		if lang != "" {
			return fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, lang, codeEscaped)
		}
		return fmt.Sprintf(`<pre><code>%s</code></pre>`, codeEscaped)
	})

	return html
}

// convertPanelMacros converts warning/info/note panels to blockquotes
func convertPanelMacros(html string) string {
	panels := map[string]string{
		"warning": "⚠️ Warning",
		"info":    "ℹ️ Info",
		"note":    "📝 Note",
	}

	for panelType, prefix := range panels {
		re := regexp.MustCompile(fmt.Sprintf(`(?s)<ac:structured-macro\s+ac:name="%s"[^>]*>(.*?)</ac:structured-macro>`, panelType))

		html = re.ReplaceAllStringFunc(html, func(match string) string {
			// Extract content from rich-text-body
			contentRe := regexp.MustCompile(`(?s)<ac:rich-text-body>(.*?)</ac:rich-text-body>`)
			contentMatch := contentRe.FindStringSubmatch(match)
			content := ""
			if len(contentMatch) > 1 {
				content = contentMatch[1]
			}

			// Return as blockquote with prefix
			return fmt.Sprintf(`<blockquote><p><strong>%s:</strong> %s</p></blockquote>`, prefix, content)
		})
	}

	return html
}

// convertInternalLinks converts Confluence internal page links to standard anchors
func convertInternalLinks(html string) string {
	// Pattern: <ac:link><ri:page ri:content-title="Page Title" /><ac:plain-text-link-body>Link Text</ac:plain-text-link-body></ac:link>
	re := regexp.MustCompile(`(?s)<ac:link>(.*?)</ac:link>`)

	html = re.ReplaceAllStringFunc(html, func(match string) string {
		// Extract page title
		titleRe := regexp.MustCompile(`<ri:page\s+ri:content-title="([^"]+)"`)
		titleMatch := titleRe.FindStringSubmatch(match)
		pageTitle := ""
		if len(titleMatch) > 1 {
			pageTitle = titleMatch[1]
		}

		// Extract link text
		textRe := regexp.MustCompile(`(?s)<ac:plain-text-link-body>(?:<!\[CDATA\[)?(.*?)(?:\]\]>)?</ac:plain-text-link-body>`)
		textMatch := textRe.FindStringSubmatch(match)
		linkText := pageTitle // Default to page title
		if len(textMatch) > 1 && textMatch[1] != "" {
			linkText = textMatch[1]
		}

		if pageTitle == "" {
			// No page reference, keep original content
			return match
		}

		// Convert page title to slug
		slug := titleToSlug(pageTitle)

		// Return as standard HTML anchor (will be converted to Markdown [text](url))
		return fmt.Sprintf(`<a href="%s.md">%s</a>`, slug, linkText)
	})

	return html
}

// removeChildrenMacro removes children page listing macro
func removeChildrenMacro(html string) string {
	re := regexp.MustCompile(`<ac:structured-macro\s+ac:name="children"[^>]*/>`)
	html = re.ReplaceAllString(html, `<!-- Child pages: (requires hierarchy context) -->`)
	return html
}

// postProcess cleans up the generated Markdown
func postProcess(markdown string) string {
	// Normalize excessive blank lines (max 2 consecutive)
	re := regexp.MustCompile(`\n{3,}`)
	markdown = re.ReplaceAllString(markdown, "\n\n")

	// Trim trailing whitespace on lines
	lines := strings.Split(markdown, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	markdown = strings.Join(lines, "\n")

	// Ensure file ends with single newline
	markdown = strings.TrimRight(markdown, "\n") + "\n"

	return markdown
}

// titleToSlug converts a page title to a URL-safe slug
func titleToSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special chars with dashes
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim dashes from ends
	slug = strings.Trim(slug, "-")

	return slug
}
