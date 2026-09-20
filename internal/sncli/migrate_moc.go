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
	// children maps a tag title to the titles of its child tags, and roots
	// lists the tags with no parent. Standard Notes nests tags by giving the
	// child a TagToParentTag reference to its parent.
	children map[string][]string
	roots    []string
}

// NewMOCBuilder creates a new MOC builder.
func NewMOCBuilder(notes items.Items, config MOCConfig) *MOCBuilder {
	mb := &MOCBuilder{
		notes:     notes,
		tags:      make(map[string]items.Items),
		tagCounts: make(map[string]int),
		config:    config,
		children:  make(map[string][]string),
	}

	mb.buildTagIndex()

	return mb
}

// buildTagIndex groups notes by their tags and records the tag hierarchy.
func (mb *MOCBuilder) buildTagIndex() {
	// Build tag map from all items
	tagMap := make(map[string]string) // UUID to title
	for _, item := range mb.notes {
		if item.GetContentType() == common.SNItemTypeTag {
			tag := item.(*items.Tag)
			tagMap[tag.UUID] = tag.Content.GetTitle()
		}
	}

	mb.buildTagTree(tagMap)

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

// buildTagTree records each tag's children and the tags that have no parent.
func (mb *MOCBuilder) buildTagTree(tagMap map[string]string) {
	hasParent := make(map[string]bool)

	for _, item := range mb.notes {
		if item.GetContentType() != common.SNItemTypeTag {
			continue
		}

		tag, ok := item.(*items.Tag)
		if !ok {
			continue
		}

		title := tag.Content.GetTitle()

		for _, ref := range tag.Content.References() {
			if ref.ContentType != common.SNItemTypeTag {
				continue
			}

			parent, exists := tagMap[ref.UUID]
			if !exists || parent == title {
				continue
			}

			mb.children[parent] = append(mb.children[parent], title)
			hasParent[title] = true
		}
	}

	for _, title := range tagMap {
		if !hasParent[title] {
			mb.roots = append(mb.roots, title)
		}
	}

	sort.Strings(mb.roots)

	for parent := range mb.children {
		sort.Strings(mb.children[parent])
	}
}

// hasTagTree reports whether any tag is nested inside another.
func (mb *MOCBuilder) hasTagTree() bool {
	return len(mb.children) > 0
}

// descendantNoteCount counts the notes under a tag and everything below it,
// counting a note once however many of those tags it carries.
func (mb *MOCBuilder) descendantNoteCount(tag string) int {
	seen := make(map[string]bool)
	mb.collectDescendantNotes(tag, seen, make(map[string]bool))

	return len(seen)
}

func (mb *MOCBuilder) collectDescendantNotes(tag string, seen, visited map[string]bool) {
	if visited[tag] {
		return
	}

	visited[tag] = true

	for _, item := range mb.tags[tag] {
		seen[item.GetUUID()] = true
	}

	for _, child := range mb.children[tag] {
		mb.collectDescendantNotes(child, seen, visited)
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
	case MOCStyleAuto:
		return mb.generateAutoMOCs()
	default:
		return mb.generateFlatMOCs()
	}
}

// generateAutoMOCs picks the layout that suits the account: the tag tree if
// there is one, themes if the notes are barely tagged, otherwise flat.
func (mb *MOCBuilder) generateAutoMOCs() ([]MOCFile, error) {
	if mb.hasTagTree() {
		return mb.generateHierarchicalMOCs()
	}

	noteCount := mb.countNotes()
	if noteCount >= 20 && len(mb.tags)*5 < noteCount {
		return mb.generateTopicMOCs()
	}

	return mb.generateFlatMOCs()
}

func (mb *MOCBuilder) countNotes() int {
	var n int

	for _, item := range mb.notes {
		if item.GetContentType() == common.SNItemTypeNote {
			n++
		}
	}

	return n
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

// generateHierarchicalMOCs mirrors the Standard Notes tag tree: each tag MOC
// links to the MOCs of its child tags as well as its own notes. MaxDepth caps
// how far down MOCs are created; tags below the cut-off still have their notes
// listed on the deepest MOC, so nothing is dropped from the vault.
func (mb *MOCBuilder) generateHierarchicalMOCs() ([]MOCFile, error) {
	if !mb.hasTagTree() {
		// Without nesting there is nothing to be hierarchical about.
		return mb.generateFlatMOCs()
	}

	depth := mb.config.MaxDepth
	if depth < 1 {
		depth = 1
	}

	var mocs []MOCFile

	visited := make(map[string]bool)

	for _, root := range mb.roots {
		if mb.tagCounts[root] == 0 && len(mb.children[root]) == 0 {
			continue
		}

		mocs = append(mocs, mb.createHierarchicalMOCs(root, 1, depth, visited)...)
	}

	home := mb.createHierarchicalHomeMOC(depth)

	return append([]MOCFile{home}, mocs...), nil
}

// createHierarchicalMOCs creates the MOC for a tag and, while there is depth
// left, for its children.
func (mb *MOCBuilder) createHierarchicalMOCs(tag string, level, maxDepth int, visited map[string]bool) []MOCFile {
	if visited[tag] {
		return nil
	}

	visited[tag] = true

	atCutOff := level >= maxDepth

	var mocs []MOCFile

	moc := mb.createTagTreeMOC(tag, level, atCutOff)
	mocs = append(mocs, moc)

	if atCutOff {
		return mocs
	}

	for _, child := range mb.children[tag] {
		mocs = append(mocs, mb.createHierarchicalMOCs(child, level+1, maxDepth, visited)...)
	}

	return mocs
}

// createTagTreeMOC renders one tag's MOC. At the depth cut-off it absorbs the
// notes of everything below it rather than linking to MOCs that do not exist.
func (mb *MOCBuilder) createTagTreeMOC(tag string, level int, atCutOff bool) MOCFile {
	var sb strings.Builder

	titleTag := toTitleCase(tag)

	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "title: %s\n", titleTag)
	fmt.Fprintf(&sb, "tags: [moc, %s]\n", tag)
	sb.WriteString("---\n\n")
	fmt.Fprintf(&sb, "# %s %s\n\n", mb.getIconForTag(tag), titleTag)

	children := mb.children[tag]

	if len(children) > 0 && !atCutOff {
		sb.WriteString("## Sub-categories\n\n")

		for _, child := range children {
			fmt.Fprintf(&sb, "- %s [[%s MOC]] (%d notes)\n",
				mb.getIconForTag(child), toTitleCase(child), mb.descendantNoteCount(child))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("## Notes\n\n")
	mb.writeNoteLinks(&sb, mb.tags[tag])

	if atCutOff && len(children) > 0 {
		sb.WriteString("\n## Notes from sub-categories\n\n")
		fmt.Fprintf(&sb, "*Below the configured MOC depth (%d), so these are listed here rather than in their own MOCs.*\n\n", mb.config.MaxDepth)

		for _, child := range children {
			mb.writeDescendantNoteLinks(&sb, child, make(map[string]bool))
		}
	}

	fmt.Fprintf(&sb, "\n---\n**Tagged Notes**: #%s (%d notes)\n", tag, len(mb.tags[tag]))

	return MOCFile{
		Filename: fmt.Sprintf("%s MOC.md", titleTag),
		Title:    fmt.Sprintf("%s MOC", titleTag),
		Content:  sb.String(),
		Tags:     []string{"moc", tag},
		Order:    level,
	}
}

func (mb *MOCBuilder) writeNoteLinks(sb *strings.Builder, notes items.Items) {
	for _, item := range notes {
		if note, ok := item.(*items.Note); ok {
			fmt.Fprintf(sb, "- [[%s]]\n", note.Content.GetTitle())
		}
	}
}

func (mb *MOCBuilder) writeDescendantNoteLinks(sb *strings.Builder, tag string, visited map[string]bool) {
	if visited[tag] {
		return
	}

	visited[tag] = true

	mb.writeNoteLinks(sb, mb.tags[tag])

	for _, child := range mb.children[tag] {
		mb.writeDescendantNoteLinks(sb, child, visited)
	}
}

// createHierarchicalHomeMOC lists the root tags and the shape of the tree.
func (mb *MOCBuilder) createHierarchicalHomeMOC(depth int) MOCFile {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString("title: Home\n")
	sb.WriteString("tags: [moc, index]\n")
	sb.WriteString("---\n\n")
	sb.WriteString("# 🏠 Home\n\n")
	sb.WriteString("Welcome to your knowledge base!\n\n")
	sb.WriteString("## 📂 Main Categories\n\n")

	for _, root := range mb.roots {
		if mb.tagCounts[root] == 0 && len(mb.children[root]) == 0 {
			continue
		}

		fmt.Fprintf(&sb, "- %s [[%s MOC]] (%d notes)\n",
			mb.getIconForTag(root), toTitleCase(root), mb.descendantNoteCount(root))
	}

	sb.WriteString("\n")

	if mb.config.IncludeStats {
		sb.WriteString("## 📊 Quick Stats\n\n")
		fmt.Fprintf(&sb, "- Total Notes: %d\n", mb.countNotes())
		fmt.Fprintf(&sb, "- Total Tags: %d\n", len(mb.tags))
		fmt.Fprintf(&sb, "- MOC Depth: %d\n\n", depth)
	}

	if mb.config.IncludeRecent {
		sb.WriteString("## 🔍 Recently Updated\n\n")

		for _, note := range mb.getRecentNotes(mb.config.RecentCount) {
			fmt.Fprintf(&sb, "- [[%s]]\n", note.Content.GetTitle())
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

// paraCategories are the PARA buckets, in the order they appear on Home.
var paraCategories = []struct {
	name     string
	icon     string
	keywords []string
}{
	{"Projects", "🚀", []string{"project", "proj", "sprint", "launch", "milestone"}},
	{"Areas", "🎯", []string{"area", "ongoing", "admin", "health", "finance", "home", "work", "personal"}},
	{"Archive", "📦", []string{"archive", "archived", "old", "done", "complete", "completed", "inactive"}},
	{"Resources", "📚", []string{"reference", "resource", "reading", "learning", "research", "notes"}},
}

// generatePARAMOCs sorts tags into Projects, Areas, Resources and Archive.
// The mapping is a guess from the tag's name, so each MOC says what it matched
// on and Resources is the fallback rather than a claim about the notes.
func (mb *MOCBuilder) generatePARAMOCs() ([]MOCFile, error) {
	buckets := make(map[string][]string)

	tagNames := make([]string, 0, len(mb.tags))
	for tag := range mb.tags {
		tagNames = append(tagNames, tag)
	}

	sort.Strings(tagNames)

	for _, tag := range tagNames {
		bucket := classifyPARATag(tag)
		buckets[bucket] = append(buckets[bucket], tag)
	}

	var mocs []MOCFile

	for _, category := range paraCategories {
		tags := buckets[category.name]
		if len(tags) == 0 {
			continue
		}

		mocs = append(mocs, mb.createPARAMOC(category.name, category.icon, tags))
	}

	home := mb.createPARAHomeMOC(buckets)

	return append([]MOCFile{home}, mocs...), nil
}

// classifyPARATag picks a bucket from the tag's name, defaulting to Resources.
func classifyPARATag(tag string) string {
	lower := strings.ToLower(tag)

	for _, category := range paraCategories {
		for _, keyword := range category.keywords {
			if strings.Contains(lower, keyword) {
				return category.name
			}
		}
	}

	return "Resources"
}

func (mb *MOCBuilder) createPARAMOC(category, icon string, tags []string) MOCFile {
	var sb strings.Builder

	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "title: %s\n", category)
	fmt.Fprintf(&sb, "tags: [moc, para, %s]\n", strings.ToLower(category))
	sb.WriteString("---\n\n")
	fmt.Fprintf(&sb, "# %s %s\n\n", icon, category)
	fmt.Fprintf(&sb, "*Tags sorted here by name. Move anything that belongs elsewhere.*\n\n")

	for _, tag := range tags {
		fmt.Fprintf(&sb, "## %s %s\n\n", mb.getIconForTag(tag), toTitleCase(tag))
		mb.writeNoteLinks(&sb, mb.tags[tag])
		sb.WriteString("\n")
	}

	return MOCFile{
		Filename: fmt.Sprintf("%s MOC.md", category),
		Title:    fmt.Sprintf("%s MOC", category),
		Content:  sb.String(),
		Tags:     []string{"moc", "para", strings.ToLower(category)},
		Order:    1,
	}
}

func (mb *MOCBuilder) createPARAHomeMOC(buckets map[string][]string) MOCFile {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString("title: Home\n")
	sb.WriteString("tags: [moc, index, para]\n")
	sb.WriteString("---\n\n")
	sb.WriteString("# 🏠 Home\n\n")
	sb.WriteString("Welcome to your knowledge base, organised with the PARA method.\n\n")

	for _, category := range paraCategories {
		tags := buckets[category.name]
		if len(tags) == 0 {
			continue
		}

		noteCount := 0
		for _, tag := range tags {
			noteCount += mb.tagCounts[tag]
		}

		fmt.Fprintf(&sb, "- %s [[%s MOC]] (%d tags, %d notes)\n",
			category.icon, category.name, len(tags), noteCount)
	}

	sb.WriteString("\n")

	if mb.config.IncludeStats {
		sb.WriteString("## 📊 Quick Stats\n\n")
		fmt.Fprintf(&sb, "- Total Notes: %d\n", mb.countNotes())
		fmt.Fprintf(&sb, "- Total Tags: %d\n\n", len(mb.tags))
	}

	if mb.config.IncludeRecent {
		sb.WriteString("## 🔍 Recently Updated\n\n")

		for _, note := range mb.getRecentNotes(mb.config.RecentCount) {
			fmt.Fprintf(&sb, "- [[%s]]\n", note.Content.GetTitle())
		}
	}

	return MOCFile{
		Filename: "Home.md",
		Title:    "Home",
		Content:  sb.String(),
		Tags:     []string{"moc", "index", "para"},
		Order:    0,
	}
}

// generateTopicMOCs builds MOCs from the themes content analysis finds, rather
// than from tags, for vaults whose notes are barely tagged.
func (mb *MOCBuilder) generateTopicMOCs() ([]MOCFile, error) {
	analyzer := NewContentAnalyzer(mb.notes)
	themes := analyzer.AnalyzeContent()

	var mocs []MOCFile

	var included []ContentTheme

	for _, theme := range themes {
		if theme.NoteCount < mb.config.MinNotesPerMOC {
			continue
		}

		mocs = append(mocs, mb.createThemeMOC(theme))
		included = append(included, theme)
	}

	home := mb.createTopicHomeMOC(included)

	return append([]MOCFile{home}, mocs...), nil
}

func (mb *MOCBuilder) createTopicHomeMOC(themes []ContentTheme) MOCFile {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString("title: Home\n")
	sb.WriteString("tags: [moc, index]\n")
	sb.WriteString("---\n\n")
	sb.WriteString("# 🏠 Home\n\n")
	sb.WriteString("Welcome to your knowledge base!\n\n")

	if len(themes) > 0 {
		sb.WriteString("## 🎯 Content Themes\n\n")

		for _, theme := range themes {
			fmt.Fprintf(&sb, "- 📝 [[%s MOC]] (%d notes)\n", theme.Name, theme.NoteCount)
		}

		sb.WriteString("\n")
	} else {
		sb.WriteString("*No themes were found in these notes. Try `--moc-style flat` to build MOCs from tags instead.*\n\n")
	}

	if mb.config.IncludeStats {
		sb.WriteString("## 📊 Quick Stats\n\n")
		fmt.Fprintf(&sb, "- Total Notes: %d\n", mb.countNotes())
		fmt.Fprintf(&sb, "- Themes Found: %d\n\n", len(themes))
	}

	if mb.config.IncludeRecent {
		sb.WriteString("## 🔍 Recently Updated\n\n")

		for _, note := range mb.getRecentNotes(mb.config.RecentCount) {
			fmt.Fprintf(&sb, "- [[%s]]\n", note.Content.GetTitle())
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
			fmt.Fprintf(&sb, "- %s [[%s MOC]] (%d notes)\n", icon, tag, noteCount)
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
			fmt.Fprintf(&sb, "- %s [[%s MOC]] (%d notes)\n", icon, theme.Name, theme.NoteCount)
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
		fmt.Fprintf(&sb, "- Total Notes: %d\n", noteCount)
		fmt.Fprintf(&sb, "- Total Tags: %d\n", len(mb.tags))
	}

	if mb.config.IncludeRecent {
		sb.WriteString("\n## 🔍 Recently Updated\n\n")
		recentNotes := mb.getRecentNotes(mb.config.RecentCount)
		for _, note := range recentNotes {
			fmt.Fprintf(&sb, "- [[%s]]\n", note.Content.GetTitle())
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
	fmt.Fprintf(&sb, "title: %s\n", titleTag)
	fmt.Fprintf(&sb, "tags: [moc, %s]\n", tag)
	sb.WriteString("---\n\n")
	fmt.Fprintf(&sb, "# %s %s\n\n", mb.getIconForTag(tag), titleTag)

	// List all notes
	sb.WriteString("## Notes\n\n")
	for _, item := range notes {
		if note, ok := item.(*items.Note); ok {
			fmt.Fprintf(&sb, "- [[%s]]\n", note.Content.GetTitle())
		}
	}

	fmt.Fprintf(&sb, "\n---\n**Tagged Notes**: #%s (%d notes)\n", tag, len(notes))

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
	fmt.Fprintf(&sb, "title: %s\n", theme.Name)
	fmt.Fprintf(&sb, "tags: [moc, theme, %s]\n", strings.ToLower(theme.Name))
	sb.WriteString("---\n\n")
	fmt.Fprintf(&sb, "# 🎯 %s\n\n", theme.Name)
	fmt.Fprintf(&sb, "*Discovered theme based on content analysis (%d notes)*\n\n", theme.NoteCount)

	// Show key phrases if available
	if len(theme.Phrases) > 0 {
		sb.WriteString("## 🔑 Key Phrases\n\n")
		for i, phrase := range theme.Phrases {
			if i >= 5 {
				break
			}
			fmt.Fprintf(&sb, "- `%s`\n", phrase)
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
				fmt.Fprintf(&sb, "- [[%s]]\n", note.Content.GetTitle())
				break
			}
		}
	}

	fmt.Fprintf(&sb, "\n---\n**Content Theme**: %d notes connected by shared concepts\n", theme.NoteCount)

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
		"work":      "💼",
		"personal":  "🏠",
		"learning":  "📚",
		"projects":  "🚀",
		"ideas":     "💡",
		"reference": "📖",
		"meetings":  "🤝",
		"planning":  "📋",
		"security":  "🔐",
		"code":      "💻",
		"design":    "🎨",
		"research":  "🔬",
		"health":    "🏥",
		"finance":   "💰",
		"travel":    "✈️",
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
