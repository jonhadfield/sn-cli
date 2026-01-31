package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gookit/color"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/pterm/pterm"
)

// ShowVisualStats displays stats with beautiful charts and visuals
func ShowVisualStats(data sncli.StatsData) error {
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
		WithMargin(10).
		Println("📊 Standard Notes Statistics")

	pterm.Println()

	// Item counts section
	showItemCountsVisual(data)
	pterm.Println()

	// Activity section
	showActivityVisual(data)
	pterm.Println()

	// Top notes section
	showTopNotesVisual(data)
	pterm.Println()

	// Duplicates section if any
	if len(data.DuplicateNotes) > 0 {
		showDuplicatesVisual(data)
		pterm.Println()
	}

	return nil
}

// showItemCountsVisual displays item counts as a bar chart
func showItemCountsVisual(data sncli.StatsData) {
	pterm.DefaultSection.Println("Item Counts")

	// Prepare data for bar chart
	var bars []pterm.Bar

	// Add core items
	if count, ok := data.CoreTypeCounter.Counts()["Note"]; ok && count > 0 {
		bars = append(bars, pterm.Bar{
			Label: "📝 Notes",
			Value: int(count),
			Style: pterm.NewStyle(pterm.FgLightBlue),
		})
	}

	if count, ok := data.CoreTypeCounter.Counts()["Tag"]; ok && count > 0 {
		bars = append(bars, pterm.Bar{
			Label: "🏷️  Tags",
			Value: int(count),
			Style: pterm.NewStyle(pterm.FgLightMagenta),
		})
	}

	// Add other items
	for itemType, count := range data.OtherTypeCounter.Counts() {
		if count > 0 {
			bars = append(bars, pterm.Bar{
				Label: fmt.Sprintf("📦 %s", itemType),
				Value: int(count),
				Style: pterm.NewStyle(pterm.FgLightCyan),
			})
		}
	}

	if len(bars) > 0 {
		// Sort by value descending
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Value > bars[j].Value
		})

		pterm.DefaultBarChart.WithBars(bars).
			WithShowValue(true).
			WithHorizontal(true).
			WithHeight(10).
			Render()
	} else {
		pterm.Info.Println("No items found")
	}
}

// showActivityVisual displays recent activity
func showActivityVisual(data sncli.StatsData) {
	pterm.DefaultSection.Println("Recent Activity")

	tableData := [][]string{
		{color.Cyan.Sprint("Metric"), color.Cyan.Sprint("Value"), color.Cyan.Sprint("When")},
	}

	if data.NewestNote != nil {
		title := truncateTitle(data.NewestNote.Content.GetTitle(), 40)
		tableData = append(tableData, []string{
			"🆕 Newest Note",
			title,
			formatDate(data.NewestNote.CreatedAt),
		})
	}

	if data.LastUpdatedNote != nil {
		title := truncateTitle(data.LastUpdatedNote.Content.GetTitle(), 40)
		tableData = append(tableData, []string{
			"✏️  Last Updated",
			title,
			formatDate(data.LastUpdatedNote.UpdatedAt),
		})
	}

	if data.OldestNote != nil {
		title := truncateTitle(data.OldestNote.Content.GetTitle(), 40)
		tableData = append(tableData, []string{
			"📜 Oldest Note",
			title,
			formatDate(data.OldestNote.CreatedAt),
		})
	}

	pterm.DefaultTable.WithHasHeader(true).
		WithData(tableData).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		Render()
}

// showTopNotesVisual displays largest notes
func showTopNotesVisual(data sncli.StatsData) {
	if len(data.LargestNotes) == 0 {
		return
	}

	pterm.DefaultSection.Println("📏 Largest Notes")

	tableData := [][]string{
		{color.Cyan.Sprint("Title"), color.Cyan.Sprint("Size"), color.Cyan.Sprint("Words")},
	}

	for i, note := range data.LargestNotes {
		if i >= 5 {
			break
		}

		title := truncateTitle(note.Content.GetTitle(), 45)
		size := formatSize(len(note.Content.GetText()))
		words := estimateWords(note.Content.GetText())

		tableData = append(tableData, []string{
			title,
			color.Yellow.Sprint(size),
			color.Green.Sprint(fmt.Sprintf("%d", words)),
		})
	}

	pterm.DefaultTable.WithHasHeader(true).
		WithData(tableData).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		Render()
}

// showDuplicatesVisual displays duplicate notes warning
func showDuplicatesVisual(data sncli.StatsData) {
	pterm.Warning.Printf("⚠️  Found %d duplicate note(s)\n", len(data.DuplicateNotes))

	if len(data.DuplicateNotes) <= 5 {
		for _, note := range data.DuplicateNotes {
			pterm.Println("  • " + truncateTitle(note.Content.GetTitle(), 60))
		}
	} else {
		for i := 0; i < 5; i++ {
			pterm.Println("  • " + truncateTitle(data.DuplicateNotes[i].Content.GetTitle(), 60))
		}
		pterm.Println(color.Gray.Sprintf("  ... and %d more", len(data.DuplicateNotes)-5))
	}
}

// Helper functions

func truncateTitle(title string, maxLen int) string {
	if title == "" {
		return color.Gray.Sprint("(Untitled)")
	}
	if len(title) > maxLen {
		return title[:maxLen-3] + "..."
	}
	return title
}

func formatDate(dateStr string) string {
	// Extract date part
	if len(dateStr) >= 10 {
		return color.Gray.Sprint(dateStr[:10])
	}
	return dateStr
}

func formatSize(bytes int) string {
	const (
		KB = 1024
		MB = KB * 1024
	)

	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func estimateWords(text string) int {
	if text == "" {
		return 0
	}
	words := strings.Fields(text)
	return len(words)
}
