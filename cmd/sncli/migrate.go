package main

import (
	"fmt"
	"time"

	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func cmdMigrate() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "migrate notes to other applications",
		Description: `Export your Standard Notes to other note-taking applications
with intelligent organization and automatic MOC generation.

Supported providers:
  - obsidian: Export to Obsidian vault (markdown + wikilinks)

Example:
  sn migrate obsidian --output ./my-vault --moc`,
		Subcommands: []*cli.Command{
			{
				Name:    "obsidian",
				Aliases: []string{"obs"},
				Usage:   "migrate to Obsidian vault",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "output directory for Obsidian vault",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "moc",
						Aliases: []string{"m"},
						Usage:   "generate Maps of Content (MOCs)",
						Value:   true,
					},
					&cli.StringFlag{
						Name:  "moc-style",
						Usage: "MOC generation style: flat, hierarchical, para, topic, auto",
						Value: "flat",
					},
					&cli.IntFlag{
						Name:  "moc-depth",
						Usage: "maximum MOC hierarchy depth",
						Value: 2,
					},
					&cli.StringFlag{
						Name:  "tag-filter",
						Usage: "only export notes with these tags (comma-separated)",
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "preview migration without writing files",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)
					return processMigrateObsidian(c, opts)
				},
			},
		},
	}
}

func processMigrateObsidian(c *cli.Context, opts configOptsOutput) error {
	// Show migration start
	pterm.Info.Println("Starting migration to Obsidian...")

	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	var cacheDBPath string
	cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}
	session.CacheDBPath = cacheDBPath

	// Parse tag filter
	var tagFilter []string
	if tagStr := c.String("tag-filter"); tagStr != "" {
		tagFilter = sncli.CommaSplit(tagStr)
	}

	// Create migration config
	migrateConfig := sncli.MigrateConfig{
		Session:      &session,
		Provider:     "obsidian",
		OutputDir:    c.String("output"),
		GenerateMOCs: c.Bool("moc"),
		MOCStyle:     sncli.MOCStyle(c.String("moc-style")),
		MOCDepth:     c.Int("moc-depth"),
		TagFilter:    tagFilter,
		DryRun:       c.Bool("dry-run"),
		Debug:        opts.debug,
	}

	// Show progress
	var spinner *pterm.SpinnerPrinter
	if !c.Bool("dry-run") {
		spinner, _ = pterm.DefaultSpinner.Start("Exporting notes...")
	} else {
		pterm.Info.Println("Running in dry-run mode (no files will be written)")
	}

	// Execute migration
	result, err := migrateConfig.Run()
	if spinner != nil {
		spinner.Stop()
	}

	if err != nil {
		pterm.Error.Printf("Migration failed: %v\n", err)
		return err
	}

	// Display results
	displayMigrationResults(c, result)

	return nil
}

func displayMigrationResults(c *cli.Context, result *sncli.MigrationResult) {
	pterm.Println()
	pterm.Success.Println("Migration completed successfully!")
	pterm.Println()

	// Create summary table
	tableData := pterm.TableData{
		{"Metric", "Value"},
		{"Notes Exported", fmt.Sprintf("%d", result.NotesExported)},
		{"MOCs Created", fmt.Sprintf("%d", result.MOCsCreated)},
		{"Tags Processed", fmt.Sprintf("%d", result.TagsProcessed)},
		{"Duration", formatDuration(result.Duration)},
		{"Output Path", result.OutputPath},
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	// Show warnings if any
	if len(result.Warnings) > 0 {
		pterm.Println()
		pterm.Warning.Println("Warnings:")
		for _, warning := range result.Warnings {
			fmt.Fprintf(c.App.Writer, "  - %s\n", warning)
		}
	}

	// Show errors if any
	if len(result.Errors) > 0 {
		pterm.Println()
		pterm.Error.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Fprintf(c.App.Writer, "  - %s\n", err)
		}
	}

	// Show next steps
	if !c.Bool("dry-run") {
		pterm.Println()
		pterm.Info.Printfln("Your Obsidian vault is ready at: %s", color.Cyan.Sprint(result.OutputPath))
		pterm.Info.Println("Next steps:")
		fmt.Fprintf(c.App.Writer, "  1. Open Obsidian\n")
		fmt.Fprintf(c.App.Writer, "  2. Click 'Open folder as vault'\n")
		fmt.Fprintf(c.App.Writer, "  3. Select: %s\n", result.OutputPath)
		fmt.Fprintf(c.App.Writer, "  4. Start exploring your notes!\n")
	}

	pterm.Println()
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
