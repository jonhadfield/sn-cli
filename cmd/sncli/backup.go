package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func cmdBackup() *cli.Command {
	return &cli.Command{
		Name:    "backup",
		Aliases: []string{"bak"},
		Usage:   "backup and restore operations",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "create a backup of all notes and tags",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "backup file path (.zip)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "incremental",
						Aliases: []string{"i"},
						Usage:   "create incremental backup (only changed items)",
					},
					&cli.StringFlag{
						Name:  "since",
						Usage: "last backup timestamp (for incremental)",
					},
					&cli.BoolFlag{
						Name:    "encrypt",
						Aliases: []string{"e"},
						Usage:   "encrypt the backup",
					},
				},
				Action: func(c *cli.Context) error {
					return runBackupCreate(c, getOpts(c))
				},
			},
			{
				Name:  "restore",
				Usage: "restore notes and tags from a backup file (overwrites matching UUIDs)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "backup file path (.zip)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "preview restore without making changes",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "restore without asking for confirmation",
					},
				},
				Action: func(c *cli.Context) error {
					return runBackupRestore(c, getOpts(c))
				},
			},
			{
				Name:  "info",
				Usage: "show backup file information",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "backup file path (.zip)",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return runBackupInfo(c)
				},
			},
		},
	}
}

func runBackupCreate(c *cli.Context, opts configOptsOutput) error {
	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Get password if encryption requested
	var password string
	if c.Bool("encrypt") {
		fmt.Print("Encryption password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = string(passwordBytes)

		if len(password) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}

		// Confirm password
		fmt.Print("Confirm password: ")
		confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		if string(confirmBytes) != password {
			return fmt.Errorf("passwords do not match")
		}
	}

	// Create backup config
	backupConfig := sncli.BackupConfig{
		Session:        &session,
		OutputFile:     c.String("output"),
		Incremental:    c.Bool("incremental"),
		LastBackupTime: c.String("since"),
		Encrypt:        c.Bool("encrypt"),
		Password:       password,
		Debug:          opts.debug,
	}

	// Show configuration
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).
		WithMargin(5).
		Println("💾 Create Backup")
	pterm.Println()

	pterm.Info.Println("Backup Configuration:")
	pterm.Printf("  Output File: %s\n", backupConfig.OutputFile)
	pterm.Printf("  Incremental: %v\n", backupConfig.Incremental)
	if backupConfig.Incremental && backupConfig.LastBackupTime != "" {
		pterm.Printf("  Since: %s\n", backupConfig.LastBackupTime)
	}
	pterm.Printf("  Encrypted: %v\n", backupConfig.Encrypt)
	pterm.Println()

	// Run backup
	spinner, _ := pterm.DefaultSpinner.Start("Creating backup...")

	if err := backupConfig.Run(); err != nil {
		spinner.Fail("Backup failed")
		return err
	}

	spinner.Success("Backup completed successfully")
	pterm.Success.Printf("Backup saved to: %s\n", backupConfig.OutputFile)

	return nil
}

func runBackupRestore(c *cli.Context, opts configOptsOutput) error {
	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Get backup info first
	manifest, err := sncli.GetBackupInfo(c.String("input"))
	if err != nil {
		return fmt.Errorf("failed to read backup info: %w", err)
	}

	// Get password if encrypted
	var password string
	if manifest.Encrypted {
		fmt.Print("Decryption password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = string(passwordBytes)
	}

	dryRun := c.Bool("dry-run")

	// Create restore config
	restoreConfig := sncli.RestoreConfig{
		Session:   &session,
		InputFile: c.String("input"),
		DryRun:    dryRun,
		Password:  password,
		Debug:     opts.debug,
	}

	// Show header
	if dryRun {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgYellow)).
			WithMargin(5).
			Println("🔍 Preview Restore (Dry Run)")
	} else {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).
			WithMargin(5).
			Println("⚠️  Restore Backup")
	}
	pterm.Println()

	pterm.Info.Println("Notes and tags in the backup are written back with their original UUIDs.")
	pterm.Info.Println("Any existing item with the same UUID will be overwritten.")
	pterm.Println()

	// Preview first so the user sees counts before confirming a real restore.
	previewConfig := restoreConfig
	previewConfig.DryRun = true

	spinner, _ := pterm.DefaultSpinner.Start("Reading backup...")

	result, err := previewConfig.Run()
	if err != nil {
		spinner.Fail("Restore failed")
		return err
	}

	spinner.Success("Backup read")
	pterm.Println()

	// Show results
	pterm.DefaultSection.Println("Backup Information")
	tableData := [][]string{
		{"Timestamp", result.Manifest.Timestamp},
		{"Incremental", fmt.Sprintf("%v", result.Manifest.Incremental)},
		{"Encrypted", fmt.Sprintf("%v", result.Manifest.Encrypted)},
		{"Version", result.Manifest.Version},
	}
	pterm.DefaultTable.WithHasHeader(false).
		WithData(tableData).
		Render()

	pterm.Println()
	pterm.DefaultSection.Println("Items to Restore")
	itemData := [][]string{
		{color.Cyan.Sprint("Type"), color.Cyan.Sprint("Count")},
		{"Notes", fmt.Sprintf("%d", result.NotesCount)},
		{"Tags", fmt.Sprintf("%d", result.TagsCount)},
	}
	pterm.DefaultTable.WithHasHeader(true).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(itemData).
		WithBoxed(true).
		Render()

	if dryRun {
		pterm.Println()
		pterm.Info.Println("This was a dry run. No changes were made.")
		pterm.Info.Println("To perform the actual restore, run without --dry-run flag")

		return nil
	}

	if result.NotesCount == 0 && result.TagsCount == 0 {
		pterm.Println()
		pterm.Warning.Println("Backup contains no notes or tags to restore")

		return nil
	}

	if !c.Bool("yes") && !confirmBackupRestore(c, result.NotesCount, result.TagsCount) {
		pterm.Println()
		pterm.Info.Println("Restore cancelled")

		return nil
	}

	spinner, _ = pterm.DefaultSpinner.Start("Restoring notes and tags...")

	result, err = restoreConfig.Run()
	if err != nil {
		spinner.Fail("Restore failed")
		return err
	}

	spinner.Success("Restore completed")
	pterm.Println()
	pterm.Success.Printf("Restored %d notes and %d tags\n", result.NotesRestored, result.TagsRestored)

	return nil
}

func confirmBackupRestore(c *cli.Context, notes, tags int) bool {
	_, _ = fmt.Fprintf(c.App.Writer, "restore %d notes and %d tags (overwriting matching UUIDs)? ", notes, tags)

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return false
	}

	return sncli.StringInSlice(input, []string{"y", "yes"}, false)
}

func runBackupInfo(c *cli.Context) error {
	filename := c.String("file")

	manifest, err := sncli.GetBackupInfo(filename)
	if err != nil {
		return err
	}

	// Show backup information
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
		WithMargin(5).
		Println("📋 Backup Information")
	pterm.Println()

	// Parse timestamp
	timestamp, _ := time.Parse(time.RFC3339, manifest.Timestamp)

	tableData := [][]string{
		{"File", filename},
		{"Created", timestamp.Format("2006-01-02 15:04:05")},
		{"Incremental", fmt.Sprintf("%v", manifest.Incremental)},
		{"Encrypted", fmt.Sprintf("%v", manifest.Encrypted)},
		{"Version", manifest.Version},
	}

	pterm.DefaultTable.WithHasHeader(false).
		WithData(tableData).
		WithBoxed(true).
		Render()

	pterm.Println()
	pterm.DefaultSection.Println("Item Counts")

	itemData := [][]string{
		{color.Cyan.Sprint("Type"), color.Cyan.Sprint("Count")},
	}

	for itemType, count := range manifest.ItemCounts {
		itemData = append(itemData, []string{
			itemType,
			fmt.Sprintf("%d", count),
		})
	}

	pterm.DefaultTable.WithHasHeader(true).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(itemData).
		WithBoxed(true).
		Render()

	return nil
}
