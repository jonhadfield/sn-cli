package sncli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

// ExportFormat represents export format type
type ExportFormat string

const (
	FormatMarkdown ExportFormat = "markdown"
	FormatHTML     ExportFormat = "html"
	FormatJSON     ExportFormat = "json"
)

// ExportEnhancedConfig holds enhanced export configuration
type ExportEnhancedConfig struct {
	Session       *cache.Session
	OutputDir     string
	Format        ExportFormat
	ByTags        bool
	WithMetadata  bool
	StaticSite    string // hugo, jekyll, or empty
	IncludeTrashed bool
	Debug         bool
}

// ExportedNote represents a note for export with metadata
type ExportedNote struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	UUID      string   `json:"uuid"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	Trashed   bool     `json:"trashed,omitempty"`
}

// Run executes the enhanced export
func (e *ExportEnhancedConfig) Run() error {
	// Get all notes
	noteFilter := items.Filter{
		Type: common.SNItemTypeNote,
	}

	getNoteConfig := GetNoteConfig{
		Session: e.Session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  []items.Filter{noteFilter},
		},
		Debug: e.Debug,
	}

	rawNotes, err := getNoteConfig.Run()
	if err != nil {
		return fmt.Errorf("failed to get notes: %w", err)
	}

	// Filter out trashed notes if requested
	if !e.IncludeTrashed {
		var filtered items.Items
		for _, item := range rawNotes {
			note := item.(*items.Note)
			if note.Content.Trashed == nil || !*note.Content.Trashed {
				filtered = append(filtered, item)
			}
		}
		rawNotes = filtered
	}

	// Get all tags for tag resolution
	tagFilter := items.Filter{
		Type: common.SNItemTypeTag,
	}

	getTagConfig := GetTagConfig{
		Session: e.Session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  []items.Filter{tagFilter},
		},
		Debug: e.Debug,
	}

	rawTags, err := getTagConfig.Run()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	// Build tag UUID to name map
	tagMap := make(map[string]string)
	for _, item := range rawTags {
		tag := item.(*items.Tag)
		tagMap[tag.UUID] = tag.Content.GetTitle()
	}

	// Convert notes to exportable format
	var exportedNotes []ExportedNote
	for _, item := range rawNotes {
		note := item.(*items.Note)

		// Resolve tag names
		var tagNames []string
		refs := note.Content.References()
		for _, ref := range refs {
			if ref.ContentType == common.SNItemTypeTag {
				if tagName, ok := tagMap[ref.UUID]; ok {
					tagNames = append(tagNames, tagName)
				}
			}
		}

		trashed := false
		if note.Content.Trashed != nil {
			trashed = *note.Content.Trashed
		}

		exportedNotes = append(exportedNotes, ExportedNote{
			Title:     note.Content.GetTitle(),
			Content:   note.Content.GetText(),
			UUID:      note.UUID,
			Tags:      tagNames,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
			Trashed:   trashed,
		})
	}

	// Export based on organization strategy
	if e.ByTags {
		return e.exportByTags(exportedNotes)
	}

	return e.exportFlat(exportedNotes)
}

// exportFlat exports all notes to a single directory
func (e *ExportEnhancedConfig) exportFlat(notes []ExportedNote) error {
	if err := os.MkdirAll(e.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, note := range notes {
		if err := e.exportNote(note, e.OutputDir); err != nil {
			return err
		}
	}

	return nil
}

// exportByTags exports notes organized by tags into folders
func (e *ExportEnhancedConfig) exportByTags(notes []ExportedNote) error {
	// Group notes by tags
	tagNotes := make(map[string][]ExportedNote)
	untagged := []ExportedNote{}

	for _, note := range notes {
		if len(note.Tags) == 0 {
			untagged = append(untagged, note)
			continue
		}

		for _, tag := range note.Tags {
			tagNotes[tag] = append(tagNotes[tag], note)
		}
	}

	// Export untagged notes
	if len(untagged) > 0 {
		untaggedDir := filepath.Join(e.OutputDir, "untagged")
		if err := os.MkdirAll(untaggedDir, 0755); err != nil {
			return fmt.Errorf("failed to create untagged directory: %w", err)
		}

		for _, note := range untagged {
			if err := e.exportNote(note, untaggedDir); err != nil {
				return err
			}
		}
	}

	// Export notes by tag
	for tag, taggedNotes := range tagNotes {
		// Sanitize tag name for directory
		dirName := sanitizeFilename(tag)
		tagDir := filepath.Join(e.OutputDir, dirName)

		if err := os.MkdirAll(tagDir, 0755); err != nil {
			return fmt.Errorf("failed to create tag directory %s: %w", tag, err)
		}

		for _, note := range taggedNotes {
			if err := e.exportNote(note, tagDir); err != nil {
				return err
			}
		}
	}

	return nil
}

// exportNote exports a single note in the specified format
func (e *ExportEnhancedConfig) exportNote(note ExportedNote, dir string) error {
	switch e.Format {
	case FormatMarkdown:
		return e.exportMarkdown(note, dir)
	case FormatHTML:
		return e.exportHTML(note, dir)
	case FormatJSON:
		return e.exportJSON(note, dir)
	default:
		return fmt.Errorf("unsupported export format: %s", e.Format)
	}
}

// exportMarkdown exports note as Markdown with optional frontmatter
func (e *ExportEnhancedConfig) exportMarkdown(note ExportedNote, dir string) error {
	filename := sanitizeFilename(note.Title)
	if filename == "" {
		filename = note.UUID[:8]
	}
	filepath := filepath.Join(dir, filename+".md")

	var content strings.Builder

	// Add frontmatter if requested or static site format specified
	if e.WithMetadata || e.StaticSite != "" {
		content.WriteString("---\n")

		switch e.StaticSite {
		case "hugo":
			content.WriteString(fmt.Sprintf("title: \"%s\"\n", escapeYAML(note.Title)))
			content.WriteString(fmt.Sprintf("date: %s\n", note.CreatedAt))
			content.WriteString(fmt.Sprintf("lastmod: %s\n", note.UpdatedAt))
			if len(note.Tags) > 0 {
				content.WriteString("tags:\n")
				for _, tag := range note.Tags {
					content.WriteString(fmt.Sprintf("  - \"%s\"\n", escapeYAML(tag)))
				}
			}
			content.WriteString("draft: false\n")

		case "jekyll":
			content.WriteString(fmt.Sprintf("title: \"%s\"\n", escapeYAML(note.Title)))
			content.WriteString(fmt.Sprintf("date: %s\n", note.CreatedAt))
			if len(note.Tags) > 0 {
				content.WriteString("tags: [")
				for i, tag := range note.Tags {
					if i > 0 {
						content.WriteString(", ")
					}
					content.WriteString(fmt.Sprintf("\"%s\"", escapeYAML(tag)))
				}
				content.WriteString("]\n")
			}

		default:
			// Generic frontmatter
			content.WriteString(fmt.Sprintf("title: \"%s\"\n", escapeYAML(note.Title)))
			content.WriteString(fmt.Sprintf("uuid: %s\n", note.UUID))
			content.WriteString(fmt.Sprintf("created: %s\n", note.CreatedAt))
			content.WriteString(fmt.Sprintf("updated: %s\n", note.UpdatedAt))
			if len(note.Tags) > 0 {
				content.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(note.Tags, ", ")))
			}
		}

		content.WriteString("---\n\n")
	}

	// Add note content
	if note.Content != "" {
		content.WriteString(note.Content)
	} else {
		content.WriteString(fmt.Sprintf("# %s\n\n(Empty note)\n", note.Title))
	}

	return os.WriteFile(filepath, []byte(content.String()), 0644)
}

// exportHTML exports note as HTML
func (e *ExportEnhancedConfig) exportHTML(note ExportedNote, dir string) error {
	filename := sanitizeFilename(note.Title)
	if filename == "" {
		filename = note.UUID[:8]
	}
	filepath := filepath.Join(dir, filename+".html")

	var html strings.Builder

	html.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	html.WriteString(fmt.Sprintf("  <meta charset=\"utf-8\">\n"))
	html.WriteString(fmt.Sprintf("  <title>%s</title>\n", escapeHTML(note.Title)))
	html.WriteString("  <style>\n")
	html.WriteString("    body { font-family: sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; }\n")
	html.WriteString("    h1 { color: #333; }\n")
	html.WriteString("    .metadata { color: #666; font-size: 0.9em; margin-bottom: 20px; }\n")
	html.WriteString("    .tags { margin-top: 10px; }\n")
	html.WriteString("    .tag { background: #e0e0e0; padding: 2px 8px; border-radius: 3px; margin-right: 5px; }\n")
	html.WriteString("    pre { background: #f5f5f5; padding: 10px; border-radius: 5px; overflow-x: auto; }\n")
	html.WriteString("  </style>\n")
	html.WriteString("</head>\n<body>\n")

	html.WriteString(fmt.Sprintf("  <h1>%s</h1>\n", escapeHTML(note.Title)))

	if e.WithMetadata {
		html.WriteString("  <div class=\"metadata\">\n")
		html.WriteString(fmt.Sprintf("    <div>Created: %s</div>\n", formatDate(note.CreatedAt)))
		html.WriteString(fmt.Sprintf("    <div>Updated: %s</div>\n", formatDate(note.UpdatedAt)))

		if len(note.Tags) > 0 {
			html.WriteString("    <div class=\"tags\">Tags: ")
			for _, tag := range note.Tags {
				html.WriteString(fmt.Sprintf("<span class=\"tag\">%s</span>", escapeHTML(tag)))
			}
			html.WriteString("</div>\n")
		}
		html.WriteString("  </div>\n")
	}

	// Convert markdown to HTML (simple conversion)
	contentHTML := simpleMarkdownToHTML(note.Content)
	html.WriteString(fmt.Sprintf("  <div class=\"content\">%s</div>\n", contentHTML))

	html.WriteString("</body>\n</html>")

	return os.WriteFile(filepath, []byte(html.String()), 0644)
}

// exportJSON exports note as JSON
func (e *ExportEnhancedConfig) exportJSON(note ExportedNote, dir string) error {
	filename := sanitizeFilename(note.Title)
	if filename == "" {
		filename = note.UUID[:8]
	}
	filepath := filepath.Join(dir, filename+".json")

	data, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal note to JSON: %w", err)
	}

	return os.WriteFile(filepath, data, 0644)
}

// sanitizeFilename removes invalid characters from filenames
func sanitizeFilename(name string) string {
	// Replace invalid characters with underscore
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Trim spaces and dots
	result = strings.TrimSpace(result)
	result = strings.Trim(result, ".")

	// Limit length
	if len(result) > 200 {
		result = result[:200]
	}

	return result
}

// escapeYAML escapes quotes in YAML strings
func escapeYAML(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// formatDate formats timestamp for display
func formatDate(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return t.Format("2006-01-02 15:04:05")
}

// simpleMarkdownToHTML does basic markdown to HTML conversion
func simpleMarkdownToHTML(md string) string {
	if md == "" {
		return "<p>(Empty note)</p>"
	}

	html := escapeHTML(md)

	// Convert headers
	lines := strings.Split(html, "\n")
	var result strings.Builder
	inCodeBlock := false

	for _, line := range lines {
		// Code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				result.WriteString("</pre>\n")
				inCodeBlock = false
			} else {
				result.WriteString("<pre>")
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			result.WriteString(line + "\n")
			continue
		}

		// Headers
		if strings.HasPrefix(line, "### ") {
			result.WriteString(fmt.Sprintf("<h3>%s</h3>\n", strings.TrimPrefix(line, "### ")))
		} else if strings.HasPrefix(line, "## ") {
			result.WriteString(fmt.Sprintf("<h2>%s</h2>\n", strings.TrimPrefix(line, "## ")))
		} else if strings.HasPrefix(line, "# ") {
			result.WriteString(fmt.Sprintf("<h1>%s</h1>\n", strings.TrimPrefix(line, "# ")))
		} else if line == "" {
			result.WriteString("<br>\n")
		} else {
			result.WriteString(fmt.Sprintf("<p>%s</p>\n", line))
		}
	}

	return result.String()
}
