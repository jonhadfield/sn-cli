package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/pterm/pterm"
	sncli "github.com/jonhadfield/sn-cli"
)

// TagStats holds statistics about a tag
type TagStats struct {
	Title     string
	UUID      string
	NoteCount int
	CreatedAt string
}

// ShowTagCloud displays tags as a visual cloud
func ShowTagCloud(opts configOptsOutput) error {
	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Sync to get latest data
	si := cache.SyncInput{
		Session: &session,
		Close:   false,
	}

	so, err := cache.Sync(si)
	if err != nil {
		return err
	}
	defer so.DB.Close()

	// Get all tags and notes
	tagFilter := items.Filter{
		Type: common.SNItemTypeTag,
	}

	getTagConfig := sncli.GetTagConfig{
		Session: &session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  []items.Filter{tagFilter},
		},
		Debug: opts.debug,
	}

	rawTags, err := getTagConfig.Run()
	if err != nil {
		return err
	}

	// Get all notes to count references
	noteFilter := items.Filter{
		Type: common.SNItemTypeNote,
	}

	getNoteConfig := sncli.GetNoteConfig{
		Session: &session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  []items.Filter{noteFilter},
		},
		Debug: opts.debug,
	}

	rawNotes, err := getNoteConfig.Run()
	if err != nil {
		return err
	}

	// Build tag statistics
	tagStats := make(map[string]*TagStats)

	for _, item := range rawTags {
		tag := item.(*items.Tag)
		tagStats[tag.UUID] = &TagStats{
			Title:     tag.Content.GetTitle(),
			UUID:      tag.UUID,
			NoteCount: 0,
			CreatedAt: tag.CreatedAt,
		}
	}

	// Count note references for each tag
	for _, item := range rawNotes {
		note := item.(*items.Note)
		refs := note.Content.References()

		for _, ref := range refs {
			if ref.ContentType == common.SNItemTypeTag {
				if stats, ok := tagStats[ref.UUID]; ok {
					stats.NoteCount++
				}
			}
		}
	}

	// Convert to slice for sorting
	var stats []*TagStats
	for _, s := range tagStats {
		stats = append(stats, s)
	}

	// Sort by note count descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].NoteCount > stats[j].NoteCount
	})

	if len(stats) == 0 {
		pterm.Info.Println("No tags found")
		return nil
	}

	// Display cloud
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgMagenta)).
		WithMargin(10).
		Println("🏷️  Tag Cloud")
	pterm.Println()

	displayCloud(stats)

	pterm.Println()
	pterm.Info.Printf("Total: %d tag(s), %d note(s)\n", len(stats), len(rawNotes))

	return nil
}

// displayCloud renders the tag cloud
func displayCloud(stats []*TagStats) {
	if len(stats) == 0 {
		return
	}

	// Find max count for sizing
	maxCount := stats[0].NoteCount
	if maxCount == 0 {
		maxCount = 1
	}

	// Define size levels
	minSize := 1
	maxSize := 5

	// Group tags by size
	var cloudLines []string
	currentLine := ""
	lineWidth := 0
	maxLineWidth := 100

	for _, stat := range stats {
		if stat.NoteCount == 0 {
			continue
		}

		// Calculate size (1-5)
		ratio := float64(stat.NoteCount) / float64(maxCount)
		size := minSize + int(math.Round(ratio*float64(maxSize-minSize)))

		// Format tag with size and color
		tag := formatTagForCloud(stat.Title, stat.NoteCount, size)

		tagLen := len(stat.Title) + 4 // Approximate visual length

		// Check if we need a new line
		if lineWidth+tagLen > maxLineWidth && currentLine != "" {
			cloudLines = append(cloudLines, currentLine)
			currentLine = ""
			lineWidth = 0
		}

		// Add tag to current line
		if currentLine != "" {
			currentLine += "  "
			lineWidth += 2
		}
		currentLine += tag
		lineWidth += tagLen
	}

	// Add final line
	if currentLine != "" {
		cloudLines = append(cloudLines, currentLine)
	}

	// Display cloud
	for _, line := range cloudLines {
		fmt.Println(line)
	}

	// Show legend
	pterm.Println()
	pterm.DefaultSection.Println("Legend")
	pterm.Println("  Size indicates number of notes (larger = more notes)")
	pterm.Println("  Color: " + color.Red.Sprint("◼ 10+") + " " +
		color.Yellow.Sprint("◼ 5-9") + " " +
		color.Green.Sprint("◼ 3-4") + " " +
		color.Cyan.Sprint("◼ 1-2"))
}

// formatTagForCloud formats a tag for cloud display
func formatTagForCloud(title string, count int, size int) string {
	// Choose color based on count
	var colorFunc func(...interface{}) string

	switch {
	case count >= 10:
		colorFunc = color.Red.Sprint
	case count >= 5:
		colorFunc = color.Yellow.Sprint
	case count >= 3:
		colorFunc = color.Green.Sprint
	default:
		colorFunc = color.Cyan.Sprint
	}

	// Format with size
	tag := fmt.Sprintf("%s(%d)", title, count)

	// Apply size styling
	switch size {
	case 5:
		return colorFunc(strings.ToUpper(tag))
	case 4:
		return color.Bold.Sprint(colorFunc(tag))
	case 3:
		return colorFunc(tag)
	case 2:
		return colorFunc(tag)
	default:
		return color.Gray.Sprint(colorFunc(tag))
	}
}

// ShowTagStats displays detailed tag statistics
func ShowTagStats(opts configOptsOutput) error {
	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Sync to get latest data
	si := cache.SyncInput{
		Session: &session,
		Close:   false,
	}

	so, err := cache.Sync(si)
	if err != nil {
		return err
	}
	defer so.DB.Close()

	// Get all tags
	tagFilter := items.Filter{
		Type: common.SNItemTypeTag,
	}

	getTagConfig := sncli.GetTagConfig{
		Session: &session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  []items.Filter{tagFilter},
		},
		Debug: opts.debug,
	}

	rawTags, err := getTagConfig.Run()
	if err != nil {
		return err
	}

	// Get all notes
	noteFilter := items.Filter{
		Type: common.SNItemTypeNote,
	}

	getNoteConfig := sncli.GetNoteConfig{
		Session: &session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  []items.Filter{noteFilter},
		},
		Debug: opts.debug,
	}

	rawNotes, err := getNoteConfig.Run()
	if err != nil {
		return err
	}

	// Build statistics
	tagStats := make(map[string]*TagStats)

	for _, item := range rawTags {
		tag := item.(*items.Tag)
		tagStats[tag.UUID] = &TagStats{
			Title:     tag.Content.GetTitle(),
			UUID:      tag.UUID,
			NoteCount: 0,
			CreatedAt: tag.CreatedAt,
		}
	}

	// Count references
	for _, item := range rawNotes {
		note := item.(*items.Note)
		refs := note.Content.References()

		for _, ref := range refs {
			if ref.ContentType == common.SNItemTypeTag {
				if stats, ok := tagStats[ref.UUID]; ok {
					stats.NoteCount++
				}
			}
		}
	}

	// Convert to slice and sort
	var stats []*TagStats
	for _, s := range tagStats {
		stats = append(stats, s)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].NoteCount > stats[j].NoteCount
	})

	// Display table
	pterm.DefaultSection.Println("Tag Statistics")

	tableData := [][]string{
		{color.Cyan.Sprint("#"), color.Cyan.Sprint("Tag"), color.Cyan.Sprint("Notes"), color.Cyan.Sprint("Created")},
	}

	for i, stat := range stats {
		noteCount := fmt.Sprintf("%d", stat.NoteCount)
		if stat.NoteCount == 0 {
			noteCount = color.Gray.Sprint("0")
		} else if stat.NoteCount >= 10 {
			noteCount = color.Red.Sprint(noteCount)
		} else if stat.NoteCount >= 5 {
			noteCount = color.Yellow.Sprint(noteCount)
		} else {
			noteCount = color.Green.Sprint(noteCount)
		}

		created := ""
		if len(stat.CreatedAt) >= 10 {
			created = stat.CreatedAt[:10]
		}

		tableData = append(tableData, []string{
			color.Gray.Sprint(fmt.Sprintf("%d", i+1)),
			stat.Title,
			noteCount,
			color.Gray.Sprint(created),
		})
	}

	pterm.DefaultTable.WithHasHeader(true).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(tableData).
		WithBoxed(true).
		Render()

	// Show summary
	pterm.Println()
	totalNotes := len(rawNotes)
	avgNotesPerTag := 0.0
	if len(stats) > 0 {
		totalTaggedNotes := 0
		for _, s := range stats {
			totalTaggedNotes += s.NoteCount
		}
		avgNotesPerTag = float64(totalTaggedNotes) / float64(len(stats))
	}

	pterm.Info.Printf("Total: %d tag(s), %d note(s)\n", len(stats), totalNotes)
	pterm.Info.Printf("Average: %.1f notes per tag\n", avgNotesPerTag)

	// Find unused tags
	unusedCount := 0
	for _, s := range stats {
		if s.NoteCount == 0 {
			unusedCount++
		}
	}
	if unusedCount > 0 {
		pterm.Warning.Printf("%d unused tag(s)\n", unusedCount)
	}

	return nil
}
