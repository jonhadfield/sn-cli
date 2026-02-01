package sncli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

// MigrateConfig holds configuration for migration operations.
type MigrateConfig struct {
	Session      *cache.Session
	Provider     string
	OutputDir    string
	GenerateMOCs bool
	MOCStyle     MOCStyle
	MOCDepth     int
	TagFilter    []string
	DryRun       bool
	Debug        bool
}

// MigrationResult contains the results of a migration operation.
type MigrationResult struct {
	NotesExported int
	MOCsCreated   int
	TagsProcessed int
	Duration      time.Duration
	OutputPath    string
	Warnings      []string
	Errors        []string
}

// MOCStyle represents different MOC organization styles.
type MOCStyle string

const (
	MOCStyleFlat         MOCStyle = "flat"
	MOCStyleHierarchical MOCStyle = "hierarchical"
	MOCStylePARA         MOCStyle = "para"
	MOCStyleTopicBased   MOCStyle = "topic"
	MOCStyleAuto         MOCStyle = "auto"
)

// Provider interface defines operations for different export targets.
type Provider interface {
	Name() string
	Validate() error
	Export(notes items.Items, config MigrationExportConfig) error
	GenerateMOCs(notes items.Items, mocConfig MOCConfig) ([]MOCFile, error)
}

// MigrationExportConfig holds provider-specific export configuration.
type MigrationExportConfig struct {
	OutputDir    string
	PreserveUUID bool
	LinkStyle    LinkStyle
	TagStyle     TagStyle
	DryRun       bool
	Debug        bool
}

// MOCConfig holds MOC generation configuration.
type MOCConfig struct {
	Style          MOCStyle
	MaxDepth       int
	MinNotesPerMOC int
	IncludeStats   bool
	IncludeRecent  bool
	RecentCount    int
}

// LinkStyle defines how links are formatted.
type LinkStyle string

const (
	LinkStyleWikilink LinkStyle = "wikilink"
	LinkStyleMarkdown LinkStyle = "markdown"
	LinkStyleRelative LinkStyle = "relative"
)

// TagStyle defines where tags appear in exported notes.
type TagStyle string

const (
	TagStyleFrontmatter TagStyle = "frontmatter"
	TagStyleInline      TagStyle = "inline"
	TagStyleBoth        TagStyle = "both"
)

// MOCFile represents a generated Map of Content file.
type MOCFile struct {
	Filename string
	Title    string
	Content  string
	Tags     []string
	Order    int
}

// Validate checks if the migration configuration is valid.
func (m *MigrateConfig) Validate() error {
	if m.Session == nil {
		return fmt.Errorf("session is required")
	}

	if m.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if m.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	// Validate MOC style
	validStyles := map[MOCStyle]bool{
		MOCStyleFlat:         true,
		MOCStyleHierarchical: true,
		MOCStylePARA:         true,
		MOCStyleTopicBased:   true,
		MOCStyleAuto:         true,
	}

	if m.GenerateMOCs && !validStyles[m.MOCStyle] {
		return fmt.Errorf("invalid MOC style: %s", m.MOCStyle)
	}

	if m.MOCDepth < 1 || m.MOCDepth > 10 {
		return fmt.Errorf("MOC depth must be between 1 and 10")
	}

	// Check if output directory exists
	if !m.DryRun {
		if _, err := os.Stat(m.OutputDir); err == nil {
			return fmt.Errorf("output directory already exists: %s", m.OutputDir)
		}
	}

	return nil
}

// Run executes the migration.
func (m *MigrateConfig) Run() (*MigrationResult, error) {
	startTime := time.Now()

	result := &MigrationResult{
		OutputPath: m.OutputDir,
		Warnings:   []string{},
		Errors:     []string{},
	}

	// Validate configuration
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Sync to get latest notes
	// Temporarily enable debug mode to disable internal spinner (CLI shows its own)
	originalDebug := m.Session.Debug
	m.Session.Debug = true
	so, err := Sync(cache.SyncInput{
		Session: m.Session,
		Close:   false,
	}, true)
	m.Session.Debug = originalDebug
	if err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}
	defer func() {
		_ = so.DB.Close()
	}()

	// Get all notes from cache
	var allPersistedItems cache.Items
	err = so.DB.All(&allPersistedItems)
	if err != nil {
		return nil, fmt.Errorf("getting items from db: %w", err)
	}

	// Convert to items
	allItems, err := allPersistedItems.ToItems(m.Session)
	if err != nil {
		return nil, fmt.Errorf("converting items: %w", err)
	}

	// Filter for notes only
	filters := items.ItemFilters{
		MatchAny: false,
		Filters: []items.Filter{
			{Type: common.SNItemTypeNote},
		},
	}

	// Add trash filter
	filters.Filters = append(filters.Filters, items.Filter{
		Type:       common.SNItemTypeNote,
		Key:        "Trash",
		Comparison: "!=",
		Value:      "true",
	})

	allItems.Filter(filters)

	// Apply tag filter if specified
	if len(m.TagFilter) > 0 {
		allItems = filterByTags(allItems, m.TagFilter)
	}

	if len(allItems) == 0 {
		return nil, fmt.Errorf("no notes found to export")
	}

	// Get provider
	provider, err := getProvider(m.Provider, m.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Validate provider
	if err := provider.Validate(); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Create output directory
	if !m.DryRun {
		if err := os.MkdirAll(m.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Export notes
	exportConfig := MigrationExportConfig{
		OutputDir:    m.OutputDir,
		PreserveUUID: true,
		LinkStyle:    LinkStyleWikilink,
		TagStyle:     TagStyleFrontmatter,
		DryRun:       m.DryRun,
		Debug:        m.Debug,
	}

	if err := provider.Export(allItems, exportConfig); err != nil {
		return nil, fmt.Errorf("export failed: %w", err)
	}

	result.NotesExported = len(allItems)

	// Count unique tags
	tagSet := make(map[string]bool)
	for _, item := range allItems {
		if note, ok := item.(*items.Note); ok {
			tags := extractNoteTags(note, allItems)
			for _, tag := range tags {
				tagSet[tag] = true
			}
		}
	}
	result.TagsProcessed = len(tagSet)

	// Generate MOCs if requested
	if m.GenerateMOCs {
		mocConfig := MOCConfig{
			Style:          m.MOCStyle,
			MaxDepth:       m.MOCDepth,
			MinNotesPerMOC: 3,
			IncludeStats:   true,
			IncludeRecent:  true,
			RecentCount:    5,
		}

		mocs, err := provider.GenerateMOCs(allItems, mocConfig)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("MOC generation failed: %v", err))
		} else {
			// Write MOC files
			for _, moc := range mocs {
				if !m.DryRun {
					mocPath := filepath.Join(m.OutputDir, moc.Filename)
					if err := os.WriteFile(mocPath, []byte(moc.Content), 0644); err != nil {
						result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to write MOC %s: %v", moc.Filename, err))
					}
				}
			}
			result.MOCsCreated = len(mocs)
		}
	}

	result.Duration = time.Since(startTime)

	return result, nil
}

// getProvider returns the appropriate provider based on name.
func getProvider(name string, outputDir string) (Provider, error) {
	switch name {
	case "obsidian", "obs":
		return NewObsidianExporter(outputDir), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}

// filterByTags filters items to only include notes with specified tags.
func filterByTags(allItems items.Items, tagFilter []string) items.Items {
	if len(tagFilter) == 0 {
		return allItems
	}

	var filtered items.Items

	for _, item := range allItems {
		note, ok := item.(*items.Note)
		if !ok {
			continue
		}

		tags := extractNoteTags(note, allItems)
		for _, tag := range tags {
			for _, filterTag := range tagFilter {
				if tag == filterTag {
					filtered = append(filtered, item)
					break
				}
			}
		}
	}

	return filtered
}

// extractNoteTags gets all tag names for a note.
func extractNoteTags(note *items.Note, allItems items.Items) []string {
	var tags []string

	// Build tag map from all items
	tagMap := make(map[string]string)
	for _, item := range allItems {
		if item.GetContentType() == common.SNItemTypeTag {
			tag := item.(*items.Tag)
			tagMap[tag.UUID] = tag.Content.GetTitle()
		}
	}

	// Get tags from note references
	refs := note.Content.References()
	for _, ref := range refs {
		if ref.ContentType == common.SNItemTypeTag {
			if tagTitle, exists := tagMap[ref.UUID]; exists {
				tags = append(tags, tagTitle)
			}
		}
	}

	return tags
}
