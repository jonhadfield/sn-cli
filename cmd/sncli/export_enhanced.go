package main

import (
	"fmt"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func cmdExport() *cli.Command {
	return &cli.Command{
		Name:    "export",
		Aliases: []string{"exp"},
		Usage:   "export notes to various formats",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "output directory path",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Value:   "markdown",
				Usage:   "export format: markdown, html, json",
			},
			&cli.BoolFlag{
				Name:    "by-tags",
				Aliases: []string{"t"},
				Usage:   "organize exported notes by tags into folders",
			},
			&cli.BoolFlag{
				Name:    "metadata",
				Aliases: []string{"m"},
				Usage:   "include metadata frontmatter",
			},
			&cli.StringFlag{
				Name:  "static-site",
				Usage: "format for static site generator: hugo, jekyll",
			},
			&cli.BoolFlag{
				Name:  "include-trashed",
				Usage: "include trashed notes in export",
			},
		},
		Action: func(c *cli.Context) error {
			return runExport(c, getOpts(c))
		},
	}
}

func runExport(c *cli.Context, opts configOptsOutput) error {
	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Parse export format
	formatStr := c.String("format")
	var format sncli.ExportFormat
	switch formatStr {
	case "markdown", "md":
		format = sncli.FormatMarkdown
	case "html":
		format = sncli.FormatHTML
	case "json":
		format = sncli.FormatJSON
	default:
		return fmt.Errorf("unsupported format: %s (supported: markdown, html, json)", formatStr)
	}

	// Validate static site format
	staticSite := c.String("static-site")
	if staticSite != "" && staticSite != "hugo" && staticSite != "jekyll" {
		return fmt.Errorf("unsupported static site generator: %s (supported: hugo, jekyll)", staticSite)
	}

	// Create export config
	exportConfig := sncli.ExportEnhancedConfig{
		Session:        &session,
		OutputDir:      c.String("output"),
		Format:         format,
		ByTags:         c.Bool("by-tags"),
		WithMetadata:   c.Bool("metadata") || staticSite != "",
		StaticSite:     staticSite,
		IncludeTrashed: c.Bool("include-trashed"),
		Debug:          opts.debug,
	}

	// Show configuration
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).
		WithMargin(5).
		Println("📦 Export Notes")
	pterm.Println()

	pterm.Info.Println("Export Configuration:")
	pterm.Printf("  Output Directory: %s\n", exportConfig.OutputDir)
	pterm.Printf("  Format: %s\n", format)
	pterm.Printf("  Organize by Tags: %v\n", exportConfig.ByTags)
	pterm.Printf("  Include Metadata: %v\n", exportConfig.WithMetadata)
	if staticSite != "" {
		pterm.Printf("  Static Site: %s\n", staticSite)
	}
	pterm.Printf("  Include Trashed: %v\n", exportConfig.IncludeTrashed)
	pterm.Println()

	// Run export
	spinner, _ := pterm.DefaultSpinner.Start("Exporting notes...")

	if err := exportConfig.Run(); err != nil {
		spinner.Fail("Export failed")
		return err
	}

	spinner.Success("Export completed successfully")
	pterm.Success.Printf("Notes exported to: %s\n", exportConfig.OutputDir)

	return nil
}
