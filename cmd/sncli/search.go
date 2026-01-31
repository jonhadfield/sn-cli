package main

import (
	"fmt"
	"strings"

	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/pterm/pterm"
	"github.com/sahilm/fuzzy"
	"github.com/urfave/cli/v2"
)

func cmdSearch() *cli.Command {
	return &cli.Command{
		Name:    "search",
		Usage:   "search notes by title and content",
		Aliases: []string{"find"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "query",
				Aliases:  []string{"q"},
				Usage:    "search query",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "content",
				Aliases: []string{"c"},
				Usage:   "search in note content (slower but more thorough)",
				Value:   true,
			},
			&cli.BoolFlag{
				Name:    "fuzzy",
				Aliases: []string{"f"},
				Usage:   "enable fuzzy matching",
			},
			&cli.BoolFlag{
				Name:  "case-sensitive",
				Usage: "case-sensitive search",
			},
			&cli.StringFlag{
				Name:  "tag",
				Usage: "filter by tag",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "maximum number of results",
				Value:   0, // 0 means unlimited
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "table",
				Usage: "output format (table, rich, json, yaml)",
			},
		},
		Action: func(c *cli.Context) error {
			return processSearch(c, getOpts(c))
		},
	}
}

func processSearch(c *cli.Context, opts configOptsOutput) error {
	query := c.String("query")
	if query == "" {
		return fmt.Errorf("search query is required")
	}

	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Sync to get latest notes
	si := cache.SyncInput{
		Session: &session,
		Close:   false,
	}

	so, err := cache.Sync(si)
	if err != nil {
		return err
	}
	defer so.DB.Close()

	// Get all notes
	noteFilter := items.Filter{
		Type: common.SNItemTypeNote,
	}

	// Add tag filter if specified
	filters := []items.Filter{noteFilter}
	if c.String("tag") != "" {
		tagFilter := items.Filter{
			Type:       common.SNItemTypeNote,
			Key:        common.SNItemTypeTag,
			Comparison: "contains",
			Value:      c.String("tag"),
		}
		filters = append(filters, tagFilter)
	}

	// Don't include trash by default
	trashFilter := items.Filter{
		Type:       common.SNItemTypeNote,
		Key:        "Trash",
		Comparison: "!=",
		Value:      "true",
	}
	filters = append(filters, trashFilter)

	getNoteConfig := sncli.GetNoteConfig{
		Session: &session,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters:  filters,
		},
		Debug: opts.debug,
	}

	rawNotes, err := getNoteConfig.Run()
	if err != nil {
		return err
	}

	// Perform search
	results := searchNotes(rawNotes, query, c.Bool("content"), c.Bool("fuzzy"), c.Bool("case-sensitive"))

	// Apply limit if specified
	limit := c.Int("limit")
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	// Display results
	if len(results) == 0 {
		pterm.Info.Println("No matches found")
		return nil
	}

	output := c.String("output")
	switch output {
	case "rich":
		if len(results) == 1 {
			return RichNoteDisplay(results[0].Note, true)
		}
		return displaySearchResults(results, query)
	case "table":
		return displaySearchResults(results, query)
	case "json", "yaml":
		// Convert search results back to items.Items
		var items items.Items
		for _, r := range results {
			items = append(items, r.Note)
		}
		return outputNotesFormat(c, items, output)
	default:
		return displaySearchResults(results, query)
	}
}

// SearchResult represents a search match
type SearchResult struct {
	Note         *items.Note
	Score        int
	MatchInTitle bool
	MatchInBody  bool
	Preview      string
}

// searchNotes performs the actual search
func searchNotes(notes items.Items, query string, searchContent bool, fuzzyMatch bool, caseSensitive bool) []SearchResult {
	var results []SearchResult

	// Prepare query for comparison
	searchQuery := query
	if !caseSensitive {
		searchQuery = strings.ToLower(query)
	}

	for _, item := range notes {
		note := item.(*items.Note)

		title := note.Content.GetTitle()
		text := note.Content.GetText()

		if !caseSensitive {
			title = strings.ToLower(title)
			text = strings.ToLower(text)
		}

		var matchInTitle, matchInBody bool
		var score int

		if fuzzyMatch {
			// Fuzzy matching using fuzzy.Find
			titleResults := fuzzy.Find(searchQuery, []string{title})
			matchInTitle = len(titleResults) > 0
			if matchInTitle {
				score += 200 // Weight title matches higher
			}

			if searchContent {
				bodyResults := fuzzy.Find(searchQuery, []string{text})
				matchInBody = len(bodyResults) > 0
				if matchInBody {
					score += 100
				}
			}
		} else {
			// Exact substring matching
			matchInTitle = strings.Contains(title, searchQuery)
			if matchInTitle {
				score += 100 // Higher score for title matches
			}

			if searchContent {
				matchInBody = strings.Contains(text, searchQuery)
				if matchInBody {
					score += 50
				}
			}
		}

		if matchInTitle || matchInBody {
			preview := generateSearchPreview(text, query, caseSensitive)
			results = append(results, SearchResult{
				Note:         note,
				Score:        score,
				MatchInTitle: matchInTitle,
				MatchInBody:  matchInBody,
				Preview:      preview,
			})
		}
	}

	// Sort by score descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// generateSearchPreview creates a preview showing context around the match
func generateSearchPreview(text, query string, caseSensitive bool) string {
	searchText := text
	searchQuery := query
	if !caseSensitive {
		searchText = strings.ToLower(text)
		searchQuery = strings.ToLower(query)
	}

	index := strings.Index(searchText, searchQuery)
	if index == -1 {
		// No match in body, show beginning
		if len(text) > 100 {
			return text[:100] + "..."
		}
		return text
	}

	// Show context around match
	start := index - 50
	if start < 0 {
		start = 0
	}

	end := index + len(query) + 50
	if end > len(text) {
		end = len(text)
	}

	preview := text[start:end]

	// Add ellipsis
	if start > 0 {
		preview = "..." + preview
	}
	if end < len(text) {
		preview = preview + "..."
	}

	// Clean up newlines for preview
	preview = strings.ReplaceAll(preview, "\n", " ")

	return preview
}

// displaySearchResults shows search results in a table
func displaySearchResults(results []SearchResult, query string) error {
	pterm.DefaultSection.Printf("Search Results for \"%s\"", query)
	pterm.Println()

	tableData := [][]string{
		{color.Cyan.Sprint("#"), color.Cyan.Sprint("Title"), color.Cyan.Sprint("Match"), color.Cyan.Sprint("Preview")},
	}

	for i, result := range results {
		var matchType string
		if result.MatchInTitle && result.MatchInBody {
			matchType = color.Green.Sprint("Title + Body")
		} else if result.MatchInTitle {
			matchType = color.Yellow.Sprint("Title")
		} else {
			matchType = color.Cyan.Sprint("Body")
		}

		title := truncateTitle(result.Note.Content.GetTitle(), 35)
		preview := truncateString(result.Preview, 50)

		tableData = append(tableData, []string{
			color.Gray.Sprint(fmt.Sprintf("%d", i+1)),
			title,
			matchType,
			color.Gray.Sprint(preview),
		})
	}

	pterm.DefaultTable.WithHasHeader(true).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(tableData).
		WithBoxed(true).
		Render()

	pterm.Info.Printf("Total: %d result(s)\n", len(results))

	return nil
}

// outputNotesFormat outputs notes in JSON or YAML format
func outputNotesFormat(c *cli.Context, notes items.Items, format string) error {
	// This would use the existing outputNotes function from note.go
	// For now, just print count
	fmt.Printf("Found %d notes (format: %s)\n", len(notes), format)
	return nil
}
