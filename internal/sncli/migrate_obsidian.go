package sncli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jonhadfield/gosn-v2/items"
)

// ObsidianExporter implements the Provider interface for Obsidian.
type ObsidianExporter struct {
	outputDir string
	noteMap   map[string]*items.Note // UUID to Note mapping
	titleMap  map[string]string      // Title to filename mapping
}

// NewObsidianExporter creates a new Obsidian exporter.
func NewObsidianExporter(outputDir string) *ObsidianExporter {
	return &ObsidianExporter{
		outputDir: outputDir,
		noteMap:   make(map[string]*items.Note),
		titleMap:  make(map[string]string),
	}
}

// Name returns the provider name.
func (o *ObsidianExporter) Name() string {
	return "obsidian"
}

// Validate checks if the exporter configuration is valid.
func (o *ObsidianExporter) Validate() error {
	if o.outputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	return nil
}

// Export exports notes to Obsidian format.
func (o *ObsidianExporter) Export(notes items.Items, config MigrationExportConfig) error {
	// Build note and title maps
	for _, item := range notes {
		if note, ok := item.(*items.Note); ok {
			o.noteMap[note.UUID] = note
			filename := o.sanitizeFilename(note.Content.GetTitle())
			o.titleMap[note.Content.GetTitle()] = filename
		}
	}

	// Export each note
	for _, item := range notes {
		note, ok := item.(*items.Note)
		if !ok {
			continue
		}

		// Convert note to markdown
		markdown := o.convertToMarkdown(note, notes, config)

		// Get filename
		filename := o.sanitizeFilename(note.Content.GetTitle())
		if filename == "" {
			filename = fmt.Sprintf("untitled-%s", note.UUID[:8])
		}
		filename = o.ensureUniqueFilename(filename)

		// Write file
		if !config.DryRun {
			filepath := filepath.Join(config.OutputDir, filename+".md")
			if err := os.WriteFile(filepath, []byte(markdown), 0644); err != nil {
				return fmt.Errorf("failed to write note %s: %w", filename, err)
			}
		}
	}

	return nil
}

// GenerateMOCs generates Maps of Content for the exported notes.
func (o *ObsidianExporter) GenerateMOCs(notes items.Items, config MOCConfig) ([]MOCFile, error) {
	builder := NewMOCBuilder(notes, config)
	return builder.Generate()
}

// convertToMarkdown converts a Standard Notes note to Obsidian markdown format.
func (o *ObsidianExporter) convertToMarkdown(note *items.Note, allNotes items.Items, config MigrationExportConfig) string {
	var sb strings.Builder

	// Add frontmatter
	if config.TagStyle == TagStyleFrontmatter || config.TagStyle == TagStyleBoth {
		sb.WriteString("---\n")
		sb.WriteString(fmt.Sprintf("title: \"%s\"\n", escapeYAMLString(note.Content.GetTitle())))

		// Get tags
		tags := extractNoteTags(note, allNotes)
		if len(tags) > 0 {
			sb.WriteString("tags: [")
			for i, tag := range tags {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(tag)
			}
			sb.WriteString("]\n")
		}

		// Add timestamps
		sb.WriteString(fmt.Sprintf("created: %s\n", note.CreatedAt))
		sb.WriteString(fmt.Sprintf("updated: %s\n", note.UpdatedAt))

		// Add UUID if configured
		if config.PreserveUUID {
			sb.WriteString(fmt.Sprintf("uuid: %s\n", note.UUID))
		}

		sb.WriteString("source: standard-notes\n")
		sb.WriteString("---\n\n")
	}

	// Add title as heading
	sb.WriteString(fmt.Sprintf("# %s\n\n", note.Content.GetTitle()))

	// Add content
	content := note.Content.GetText()
	content = o.convertLinks(content)
	sb.WriteString(content)

	// Add inline tags if configured
	if config.TagStyle == TagStyleInline || config.TagStyle == TagStyleBoth {
		tags := extractNoteTags(note, allNotes)
		if len(tags) > 0 {
			sb.WriteString("\n\n")
			for _, tag := range tags {
				sb.WriteString(fmt.Sprintf("#%s ", tag))
			}
		}
	}

	return sb.String()
}

// convertLinks converts any links in content to Obsidian wikilinks.
func (o *ObsidianExporter) convertLinks(content string) string {
	// This is a basic implementation - could be enhanced to detect
	// Standard Notes references and convert them to wikilinks
	return content
}

// sanitizeFilename creates a valid filename from a note title.
func (o *ObsidianExporter) sanitizeFilename(title string) string {
	// Remove or replace invalid filename characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	filename := invalidChars.ReplaceAllString(title, "-")

	// Replace multiple spaces with single space
	multiSpace := regexp.MustCompile(`\s+`)
	filename = multiSpace.ReplaceAllString(filename, " ")

	// Trim spaces
	filename = strings.TrimSpace(filename)

	// Limit length
	maxLen := 200
	if len(filename) > maxLen {
		filename = filename[:maxLen]
	}

	// Handle empty filename
	if filename == "" {
		filename = "untitled"
	}

	return filename
}

// ensureUniqueFilename ensures the filename is unique by appending numbers if needed.
func (o *ObsidianExporter) ensureUniqueFilename(filename string) string {
	// Check if already used
	if _, exists := o.titleMap[filename]; !exists {
		o.titleMap[filename] = filename
		return filename
	}

	// Append number to make unique
	counter := 1
	for {
		uniqueName := fmt.Sprintf("%s-%d", filename, counter)
		if _, exists := o.titleMap[uniqueName]; !exists {
			o.titleMap[uniqueName] = uniqueName
			return uniqueName
		}
		counter++
	}
}

// escapeYAMLString escapes special characters in YAML strings.
func escapeYAMLString(s string) string {
	// Escape quotes
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
