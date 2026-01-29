package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
)

func cmdOrganize() *cli.Command {
	return &cli.Command{
		Name:  "organize",
		Usage: "organize notes using AI to suggest tags and improve titles",
		Description: `Use Google Gemini AI to automatically organize your notes by:
  - Suggesting relevant tags (preferring existing tags)
  - Improving vague or generic titles
  - Providing explanations for proposed changes

The command shows a preview of changes before applying them.

Note: Your note content will be sent to Google Gemini AI for analysis.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "gemini-key",
				Usage:   "Google Gemini API key (or set GEMINI_API_KEY env var)",
				EnvVars: []string{"GEMINI_API_KEY"},
			},
			&cli.StringFlag{
				Name:  "since",
				Usage: "only process notes created after this date (RFC3339: 2024-01-01T00:00:00Z)",
			},
			&cli.StringFlag{
				Name:  "until",
				Usage: "only process notes created before this date (RFC3339: 2024-12-31T23:59:59Z)",
			},
			&cli.StringFlag{
				Name:  "uuid",
				Usage: "specific note UUID(s) to process (comma-separated)",
			},
			&cli.StringFlag{
				Name:  "title",
				Usage: "process notes with title containing this text (comma-separated)",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "skip confirmation and apply changes automatically",
			},
		},
		Action: func(c *cli.Context) error {
			opts := getOpts(c)
			return processOrganize(c, opts)
		},
	}
}

func processOrganize(c *cli.Context, opts configOptsOutput) error {
	// 1. Validate Gemini API key
	geminiKey := c.String("gemini-key")
	if geminiKey == "" {
		return errors.New("GEMINI_API_KEY is required (use --gemini-key or set environment variable)")
	}

	// 2. Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// 3. Parse date filters
	sinceDate := c.String("since")
	if sinceDate != "" {
		if _, err := time.Parse(time.RFC3339, sinceDate); err != nil {
			return fmt.Errorf("invalid --since date format (use RFC3339: 2024-01-01T00:00:00Z): %w", err)
		}
	}

	untilDate := c.String("until")
	if untilDate != "" {
		if _, err := time.Parse(time.RFC3339, untilDate); err != nil {
			return fmt.Errorf("invalid --until date format (use RFC3339: 2024-12-31T23:59:59Z): %w", err)
		}
	}

	// 4. Parse UUID and title filters
	var uuids []string
	if c.String("uuid") != "" {
		uuids = sncli.CommaSplit(c.String("uuid"))
	}

	var titles []string
	if c.String("title") != "" {
		titles = sncli.CommaSplit(c.String("title"))
	}

	// 5. Build config
	config := sncli.OrganizeConfig{
		Session:   &session,
		GeminiKey: geminiKey,
		Since:     sinceDate,
		Until:     untilDate,
		UUIDs:     uuids,
		Titles:    titles,
		AutoApply: c.Bool("yes"),
		Debug:     opts.debug,
	}

	// 6. Show processing message
	fmt.Println(color.Cyan.Sprint("Analyzing notes with Gemini AI..."))

	// 7. Run organize operation
	output, err := config.Run()
	if err != nil {
		return err
	}

	// 8. Handle results
	if len(output.Changes) == 0 {
		fmt.Println(color.Green.Sprint("No changes suggested - your notes look good!"))
		return nil
	}

	// 9. Display preview table
	displayOrganizePreview(output.Changes)

	// 10. Get confirmation
	if !config.AutoApply {
		fmt.Printf("\n%s ", color.Yellow.Sprint("Apply these changes? [y/N]:"))
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println(color.Yellow.Sprint("Changes cancelled."))
			return nil
		}
	}

	// 11. Apply changes
	fmt.Println(color.Cyan.Sprint("\nApplying changes..."))

	if err := sncli.ApplyOrganizeChanges(&session, output.Changes, opts.debug); err != nil {
		return fmt.Errorf("failed to apply changes: %w", err)
	}

	// 12. Report success
	titleChanges := 0
	tagChanges := 0
	for _, change := range output.Changes {
		if change.TitleChanged {
			titleChanges++
		}
		if change.TagsChanged {
			tagChanges++
		}
	}

	fmt.Printf("\n%s\n", color.Green.Sprint("✓ Organization complete!"))
	fmt.Printf("  - %d note titles updated\n", titleChanges)
	fmt.Printf("  - %d notes retagged\n", tagChanges)

	return nil
}

func displayOrganizePreview(changes []sncli.ProposedChange) {
	fmt.Println(color.Cyan.Sprint("\nProposed Changes:\n"))

	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Current Title"},
			{Align: simpletable.AlignCenter, Text: "New Title"},
			{Align: simpletable.AlignCenter, Text: "Current Tags"},
			{Align: simpletable.AlignCenter, Text: "Proposed Tags"},
			{Align: simpletable.AlignCenter, Text: "Reason"},
		},
	}

	for _, change := range changes {
		currentTitle := change.NoteTitle
		newTitle := change.NoteTitle
		if change.TitleChanged {
			newTitle = color.Green.Sprint(change.NewTitle)
		} else {
			newTitle = color.Gray.Sprint("(no change)")
		}

		currentTags := strings.Join(change.ExistingTags, ", ")
		if currentTags == "" {
			currentTags = color.Gray.Sprint("(none)")
		}

		proposedTags := strings.Join(change.ProposedTags, ", ")
		if change.TagsChanged {
			proposedTags = color.Green.Sprint(proposedTags)
		} else {
			proposedTags = color.Gray.Sprint("(no change)")
		}

		// Truncate reason if too long
		reason := change.Reason
		if len(reason) > 60 {
			reason = reason[:57] + "..."
		}

		r := []*simpletable.Cell{
			{Text: truncateString(currentTitle, 30)},
			{Text: truncateString(newTitle, 30)},
			{Text: truncateString(currentTags, 25)},
			{Text: truncateString(proposedTags, 25)},
			{Text: truncateString(reason, 60)},
		}

		table.Body.Cells = append(table.Body.Cells, r)
	}

	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
