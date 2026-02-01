package sncli

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

// MOCBuilder builds Maps of Content from notes.
type MOCBuilder struct {
	notes     items.Items
	tags      map[string]items.Items // tag name to notes
	tagCounts map[string]int
	config    MOCConfig
}

// NewMOCBuilder creates a new MOC builder.
func NewMOCBuilder(notes items.Items, config MOCConfig) *MOCBuilder {
	mb := &MOCBuilder{
		notes:     notes,
		tags:      make(map[string]items.Items),
		tagCounts: make(map[string]int),
		config:    config,
	}

	mb.buildTagIndex()

	return mb
}

// buildTagIndex groups notes by their tags.
func (mb *MOCBuilder) buildTagIndex() {
	// Build tag map from all items
	tagMap := make(map[string]string) // UUID to title
	for _, item := range mb.notes {
		if item.GetContentType() == common.SNItemTypeTag {
			tag := item.(*items.Tag)
			tagMap[tag.UUID] = tag.Content.GetTitle()
		}
	}

	// Group notes by tag
	for _, item := range mb.notes {
		if item.GetContentType() != common.SNItemTypeNote {
			continue
		}

		note := item.(*items.Note)
		refs := note.Content.References()

		for _, ref := range refs {
			if ref.ContentType == common.SNItemTypeTag {
				if tagTitle, exists := tagMap[ref.UUID]; exists {
					mb.tags[tagTitle] = append(mb.tags[tagTitle], item)
					mb.tagCounts[tagTitle]++
				}
			}
		}
	}
}

// Generate generates MOC files based on the configured style.
func (mb *MOCBuilder) Generate() ([]MOCFile, error) {
	switch mb.config.Style {
	case MOCStyleFlat:
		return mb.generateFlatMOCs()
	case MOCStyleHierarchical:
		return mb.generateHierarchicalMOCs()
	case MOCStylePARA:
		return mb.generatePARAMOCs()
	case MOCStyleTopicBased:
		return mb.generateTopicMOCs()
	default:
		return mb.generateFlatMOCs()
	}
}

// generateFlatMOCs generates a flat MOC structure.
func (mb *MOCBuilder) generateFlatMOCs() ([]MOCFile, error) {
	mocs := []MOCFile{}

	// Run content analysis to discover themes
	analyzer := NewContentAnalyzer(mb.notes)
	themes := analyzer.AnalyzeContent()

	// Generate MOC for each top-level tag
	topLevelTags := mb.identifyTopLevelTags()
	for _, tag := range topLevelTags {
		moc := mb.createTagMOC(tag)
		mocs = append(mocs, moc)
	}

	// Generate MOC for each discovered content theme
	for _, theme := range themes {
		if theme.NoteCount >= 2 {
			moc := mb.createThemeMOC(theme)
			mocs = append(mocs, moc)
		}
	}

	// Generate Home MOC (after others so it can reference them)
	homeMOC := mb.createHomeMOC(topLevelTags, themes)
	mocs = append([]MOCFile{homeMOC}, mocs...)

	return mocs, nil
}

// generateHierarchicalMOCs generates a hierarchical MOC structure.
func (mb *MOCBuilder) generateHierarchicalMOCs() ([]MOCFile, error) {
	// For now, use flat MOCs - can be enhanced later
	return mb.generateFlatMOCs()
}

// generatePARAMOCs generates MOCs using the PARA method.
func (mb *MOCBuilder) generatePARAMOCs() ([]MOCFile, error) {
	// For now, use flat MOCs - can be enhanced later
	return mb.generateFlatMOCs()
}

// generateTopicMOCs generates topic-based MOCs.
func (mb *MOCBuilder) generateTopicMOCs() ([]MOCFile, error) {
	// For now, use flat MOCs - can be enhanced later
	return mb.generateFlatMOCs()
}

// createHomeMOC creates the main Home MOC file.
func (mb *MOCBuilder) createHomeMOC(topLevelTags []string, themes []ContentTheme) MOCFile {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString("title: Home\n")
	sb.WriteString("tags: [moc, index]\n")
	sb.WriteString("---\n\n")
	sb.WriteString("# 🏠 Home\n\n")
	sb.WriteString("Welcome to your knowledge base!\n\n")

	// Tag-based MOCs
	if len(topLevelTags) > 0 {
		sb.WriteString("## 📂 Main Categories\n\n")
		for _, tag := range topLevelTags {
			icon := mb.getIconForTag(tag)
			noteCount := mb.tagCounts[tag]
			sb.WriteString(fmt.Sprintf("- %s [[%s MOC]] (%d notes)\n", icon, tag, noteCount))
		}
		sb.WriteString("\n")
	}

	// Content theme MOCs
	if len(themes) > 0 {
		sb.WriteString("## 🎯 Content Themes\n\n")
		for i, theme := range themes {
			if i >= 10 {
				break // Limit to top 10 in home
			}
			icon := "📝"
			sb.WriteString(fmt.Sprintf("- %s [[%s MOC]] (%d notes)\n", icon, theme.Name, theme.NoteCount))
		}
		sb.WriteString("\n")
	}

	if mb.config.IncludeStats {
		sb.WriteString("\n## 📊 Quick Stats\n\n")
		noteCount := 0
		for _, item := range mb.notes {
			if item.GetContentType() == common.SNItemTypeNote {
				noteCount++
			}
		}
		sb.WriteString(fmt.Sprintf("- Total Notes: %d\n", noteCount))
		sb.WriteString(fmt.Sprintf("- Total Tags: %d\n", len(mb.tags)))
	}

	if mb.config.IncludeRecent {
		sb.WriteString("\n## 🔍 Recently Updated\n\n")
		recentNotes := mb.getRecentNotes(mb.config.RecentCount)
		for _, note := range recentNotes {
			sb.WriteString(fmt.Sprintf("- [[%s]]\n", note.Content.GetTitle()))
		}
	}

	return MOCFile{
		Filename: "Home.md",
		Title:    "Home",
		Content:  sb.String(),
		Tags:     []string{"moc", "index"},
		Order:    0,
	}
}

// createTagMOC creates a MOC for a specific tag.
func (mb *MOCBuilder) createTagMOC(tag string) MOCFile {
	var sb strings.Builder

	notes := mb.tags[tag]
	titleTag := toTitleCase(tag)

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("title: %s\n", titleTag))
	sb.WriteString(fmt.Sprintf("tags: [moc, %s]\n", tag))
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# %s %s\n\n", mb.getIconForTag(tag), titleTag))

	// List all notes
	sb.WriteString("## Notes\n\n")
	for _, item := range notes {
		if note, ok := item.(*items.Note); ok {
			sb.WriteString(fmt.Sprintf("- [[%s]]\n", note.Content.GetTitle()))
		}
	}

	sb.WriteString(fmt.Sprintf("\n---\n**Tagged Notes**: #%s (%d notes)\n", tag, len(notes)))

	return MOCFile{
		Filename: fmt.Sprintf("%s MOC.md", titleTag),
		Title:    fmt.Sprintf("%s MOC", titleTag),
		Content:  sb.String(),
		Tags:     []string{"moc", tag},
		Order:    1,
	}
}

// createThemeMOC creates a MOC for a discovered content theme.
func (mb *MOCBuilder) createThemeMOC(theme ContentTheme) MOCFile {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("title: %s\n", theme.Name))
	sb.WriteString(fmt.Sprintf("tags: [moc, theme, %s]\n", strings.ToLower(theme.Name)))
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# 🎯 %s\n\n", theme.Name))
	sb.WriteString(fmt.Sprintf("*Discovered theme based on content analysis (%d notes)*\n\n", theme.NoteCount))

	// Show key phrases if available
	if len(theme.Phrases) > 0 {
		sb.WriteString("## 🔑 Key Phrases\n\n")
		for i, phrase := range theme.Phrases {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("- `%s`\n", phrase))
		}
		sb.WriteString("\n")
	}

	// List related notes
	sb.WriteString("## 📄 Related Notes\n\n")
	for _, noteUUID := range theme.RelatedNotes {
		// Find the note by UUID
		for _, item := range mb.notes {
			if item.GetContentType() != common.SNItemTypeNote {
				continue
			}
			note, ok := item.(*items.Note)
			if ok && note.UUID == noteUUID {
				sb.WriteString(fmt.Sprintf("- [[%s]]\n", note.Content.GetTitle()))
				break
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\n---\n**Content Theme**: %d notes connected by shared concepts\n", theme.NoteCount))

	return MOCFile{
		Filename: fmt.Sprintf("%s MOC.md", theme.Name),
		Title:    fmt.Sprintf("%s MOC", theme.Name),
		Content:  sb.String(),
		Tags:     []string{"moc", "theme", strings.ToLower(theme.Name)},
		Order:    2,
	}
}

// identifyTopLevelTags identifies the most important tags to create MOCs for.
func (mb *MOCBuilder) identifyTopLevelTags() []string {
	type tagScore struct {
		tag   string
		score float64
		count int
	}

	totalNotes := 0
	for _, item := range mb.notes {
		if item.GetContentType() == common.SNItemTypeNote {
			totalNotes++
		}
	}

	if totalNotes == 0 {
		return []string{}
	}

	var scored []tagScore
	for tag, count := range mb.tagCounts {
		// Skip tags with too few notes
		if count < mb.config.MinNotesPerMOC {
			continue
		}

		// Calculate frequency score
		frequency := float64(count) / float64(totalNotes)
		score := frequency

		// Boost score for known top-level categories
		topLevelCategories := []string{"work", "personal", "learning", "projects", "ideas", "reference"}
		for _, cat := range topLevelCategories {
			if strings.EqualFold(tag, cat) {
				score += 0.5
			}
		}

		scored = append(scored, tagScore{tag, score, count})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].count > scored[j].count
		}
		return scored[i].score > scored[j].score
	})

	// Take top tags (max 10)
	maxTopLevel := 10
	if len(scored) < maxTopLevel {
		maxTopLevel = len(scored)
	}

	result := make([]string, maxTopLevel)
	for i := 0; i < maxTopLevel; i++ {
		result[i] = scored[i].tag
	}

	return result
}

// getRecentNotes gets the most recently updated notes.
func (mb *MOCBuilder) getRecentNotes(count int) []*items.Note {
	var notes []*items.Note

	for _, item := range mb.notes {
		if note, ok := item.(*items.Note); ok {
			notes = append(notes, note)
		}
	}

	// Sort by UpdatedAt descending
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].UpdatedAt > notes[j].UpdatedAt
	})

	// Take top N
	if len(notes) > count {
		notes = notes[:count]
	}

	return notes
}

// getIconForTag returns an appropriate emoji icon for a tag.
func (mb *MOCBuilder) getIconForTag(tag string) string {
	icons := map[string]string{
		"work":     "💼",
		"personal": "🏠",
		"learning": "📚",
		"projects": "🚀",
		"ideas":    "💡",
		"reference": "📖",
		"meetings": "🤝",
		"planning": "📋",
		"security": "🔐",
		"code":     "💻",
		"design":   "🎨",
		"research": "🔬",
		"health":   "🏥",
		"finance":  "💰",
		"travel":   "✈️",
	}

	if icon, exists := icons[strings.ToLower(tag)]; exists {
		return icon
	}

	return "📄"
}

// toTitleCase converts a string to title case (first letter uppercase).
func toTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
