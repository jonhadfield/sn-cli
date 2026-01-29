package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/pterm/pterm"
)

// RichNoteDisplay renders a note with beautiful markdown formatting
func RichNoteDisplay(note *items.Note, showMetadata bool) error {
	// Create glamour renderer with auto-detect theme based on terminal background
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	// Print separator
	pterm.Println()
	pterm.DefaultBox.WithTitle(note.Content.GetTitle()).
		WithTitleTopCenter().
		WithRightPadding(2).
		WithLeftPadding(2).
		Println()

	// Render metadata if requested
	if showMetadata {
		displayNoteMetadata(note)
		pterm.Println()
	}

	// Render the note content as markdown
	text := note.Content.GetText()
	if text == "" {
		pterm.Warning.Println("(Empty note)")
		return nil
	}

	rendered, err := r.Render(text)
	if err != nil {
		// If markdown rendering fails, just print the text
		fmt.Println(text)
		return nil
	}

	fmt.Print(rendered)
	pterm.Println()

	return nil
}

// displayNoteMetadata shows note metadata in a formatted way
func displayNoteMetadata(note *items.Note) {
	data := [][]string{
		{"UUID", note.UUID},
		{"Created", note.CreatedAt},
		{"Updated", note.UpdatedAt},
	}

	// Add tags if present
	refs := note.Content.References()
	var tags []string
	for _, ref := range refs {
		if ref.ContentType == "Tag" {
			tags = append(tags, ref.UUID)
		}
	}
	if len(tags) > 0 {
		data = append(data, []string{"Tags", fmt.Sprintf("%d tag(s)", len(tags))})
	}

	// Add trashed status
	if note.Content.Trashed != nil && *note.Content.Trashed {
		data = append(data, []string{"Status", color.Red.Sprint("🗑️  Trashed")})
	}

	// Create metadata table
	pterm.DefaultTable.WithHasHeader(false).
		WithData(data).
		WithBoxed(false).
		Render()
}

// RichNoteList displays notes in a beautiful table format
func RichNoteList(notes items.Items, showPreview bool) error {
	if len(notes) == 0 {
		pterm.Info.Println("No notes found")
		return nil
	}

	// Create table header
	header := []string{"#", "Title", "Updated"}
	if showPreview {
		header = append(header, "Preview")
	}

	// Build table data
	data := [][]string{header}
	for i, item := range notes {
		note := item.(*items.Note)

		// Format row
		row := []string{
			color.Gray.Sprint(fmt.Sprintf("%d", i+1)),
			truncateAndStyleTitle(note.Content.GetTitle(), note.Content.Trashed),
			formatTime(note.UpdatedAt),
		}

		if showPreview {
			preview := generatePreview(note.Content.GetText(), 50)
			row = append(row, color.Gray.Sprint(preview))
		}

		data = append(data, row)
	}

	// Render table
	pterm.DefaultTable.WithHasHeader(true).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(data).
		WithBoxed(true).
		Render()

	// Show count
	pterm.Info.Printf("Total: %d note(s)\n", len(notes))

	return nil
}

// truncateAndStyleTitle formats note title with appropriate styling
func truncateAndStyleTitle(title string, trashed *bool) string {
	if title == "" {
		title = "(Untitled)"
	}

	// Truncate if too long
	if len(title) > 60 {
		title = title[:57] + "..."
	}

	// Style based on status
	if trashed != nil && *trashed {
		return color.Gray.Sprint("🗑️  " + title)
	}

	return title
}

// generatePreview creates a short preview from note text
func generatePreview(text string, maxLen int) string {
	// Remove multiple newlines
	text = strings.ReplaceAll(text, "\n\n", " ")
	text = strings.ReplaceAll(text, "\n", " ")

	// Remove markdown formatting for preview
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "#", "")

	// Trim whitespace
	text = strings.TrimSpace(text)

	if len(text) == 0 {
		return "(empty)"
	}

	if len(text) > maxLen {
		return text[:maxLen-3] + "..."
	}

	return text
}

// formatTime formats timestamp for display
func formatTime(timestamp string) string {
	// Just show the date part for brevity
	if len(timestamp) > 10 {
		return color.Cyan.Sprint(timestamp[:10])
	}
	return timestamp
}

// ShowProgress displays a spinner with a message
func ShowProgress(message string) (*pterm.SpinnerPrinter, error) {
	spinner, err := pterm.DefaultSpinner.Start(message)
	if err != nil {
		return nil, err
	}
	return spinner, nil
}

// ProgressBar creates a progress bar for operations
func ProgressBar(title string, total int) *pterm.ProgressbarPrinter {
	pb, _ := pterm.DefaultProgressbar.
		WithTitle(title).
		WithTotal(total).
		WithShowCount(true).
		Start()
	return pb
}
