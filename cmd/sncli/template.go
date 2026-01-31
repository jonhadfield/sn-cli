package main

import (
	"fmt"
	"strings"

	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func cmdTemplate() *cli.Command {
	return &cli.Command{
		Name:    "template",
		Aliases: []string{"tpl"},
		Usage:   "manage note templates",
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list available templates",
				Action: func(c *cli.Context) error {
					return listTemplates(c)
				},
			},
			{
				Name:  "create",
				Usage: "create a new custom template",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Required: true,
						Usage:    "template name",
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "template description",
					},
					&cli.StringFlag{
						Name:  "title",
						Usage: "note title template (supports variables like {{date}})",
					},
					&cli.StringFlag{
						Name:  "content",
						Usage: "note content template",
					},
					&cli.StringFlag{
						Name:  "tags",
						Usage: "comma-separated tags to apply",
					},
				},
				Action: func(c *cli.Context) error {
					return createTemplate(c)
				},
			},
			{
				Name:  "show",
				Usage: "show template details",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Required: true,
						Usage:    "template name",
					},
				},
				Action: func(c *cli.Context) error {
					return showTemplate(c)
				},
			},
			{
				Name:  "use",
				Usage: "create a note from template",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "template",
						Aliases:  []string{"t"},
						Required: true,
						Usage:    "template name to use",
					},
					&cli.StringFlag{
						Name:  "title",
						Usage: "value for {{title}} variable",
					},
					&cli.StringFlag{
						Name:  "var",
						Usage: "custom variables (format: key=value, repeat flag for multiple)",
					},
					&cli.StringSliceFlag{
						Name:  "vars",
						Usage: "custom variables (format: key=value)",
					},
				},
				Action: func(c *cli.Context) error {
					return useTemplate(c, getOpts(c))
				},
			},
		},
	}
}

func listTemplates(c *cli.Context) error {
	templateDir, err := sncli.GetTemplateDir()
	if err != nil {
		return err
	}

	templates := sncli.ListTemplates(templateDir)

	if len(templates) == 0 {
		pterm.Info.Println("No templates available")
		return nil
	}

	// Display templates in a table
	pterm.DefaultSection.Println("Available Templates")

	tableData := [][]string{
		{color.Cyan.Sprint("Name"), color.Cyan.Sprint("Description"), color.Cyan.Sprint("Tags")},
	}

	for _, tpl := range templates {
		// Mark built-in templates
		name := tpl.Name
		if _, ok := sncli.BuiltInTemplates()[tpl.Name]; ok {
			name = color.Green.Sprint(tpl.Name + " (built-in)")
		}

		tags := strings.Join(tpl.Tags, ", ")
		if tags == "" {
			tags = color.Gray.Sprint("(none)")
		}

		tableData = append(tableData, []string{
			name,
			truncateString(tpl.Description, 50),
			tags,
		})
	}

	pterm.DefaultTable.WithHasHeader(true).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(tableData).
		WithBoxed(true).
		Render()

	pterm.Info.Printf("Total: %d template(s)\n", len(templates))
	pterm.Println()
	pterm.Info.Println("Use 'sncli template use --template <name>' to create a note from a template")

	return nil
}

func showTemplate(c *cli.Context) error {
	name := c.String("name")

	templateDir, err := sncli.GetTemplateDir()
	if err != nil {
		return err
	}

	tpl, err := sncli.GetTemplate(name, templateDir)
	if err != nil {
		return err
	}

	// Display template details
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
		WithMargin(5).
		Printf("Template: %s", tpl.Name)
	pterm.Println()

	if tpl.Description != "" {
		pterm.Printf("Description: %s\n\n", tpl.Description)
	}

	// Show metadata
	pterm.DefaultSection.Println("Template Details")
	metadata := [][]string{
		{"Title Template", tpl.Title},
	}
	if len(tpl.Tags) > 0 {
		metadata = append(metadata, []string{"Tags", strings.Join(tpl.Tags, ", ")})
	}

	pterm.DefaultTable.WithHasHeader(false).
		WithData(metadata).
		Render()

	// Show content preview
	pterm.Println()
	pterm.DefaultSection.Println("Content Template")
	fmt.Println(tpl.Content)

	// Show available variables
	pterm.Println()
	pterm.DefaultSection.Println("Available Variables")
	vars := sncli.GetDefaultVariables()
	varList := []string{}
	for k, v := range vars {
		varList = append(varList, fmt.Sprintf("{{%s}} = %s", k, v))
	}
	for _, v := range varList {
		pterm.Println("  • " + v)
	}

	return nil
}

func createTemplate(c *cli.Context) error {
	name := c.String("name")
	description := c.String("description")
	title := c.String("title")
	content := c.String("content")
	tagsStr := c.String("tags")

	if title == "" {
		title = "{{title}}"
	}

	if content == "" {
		content = "# {{title}}\n\nCreated: {{datetime}}\n\n"
	}

	var tags []string
	if tagsStr != "" {
		tags = sncli.CommaSplit(tagsStr)
	}

	tpl := sncli.Template{
		Name:        name,
		Description: description,
		Title:       title,
		Content:     content,
		Tags:        tags,
	}

	templateDir, err := sncli.GetTemplateDir()
	if err != nil {
		return err
	}

	if err := sncli.SaveTemplate(tpl, templateDir); err != nil {
		return err
	}

	pterm.Success.Printf("Template '%s' created successfully\n", name)
	pterm.Info.Printf("Template saved to: %s/%s.yaml\n", templateDir, name)

	return nil
}

func useTemplate(c *cli.Context, opts configOptsOutput) error {
	templateName := c.String("template")

	// Get template
	templateDir, err := sncli.GetTemplateDir()
	if err != nil {
		return err
	}

	tpl, err := sncli.GetTemplate(templateName, templateDir)
	if err != nil {
		return err
	}

	// Build variables map
	vars := make(map[string]string)

	// Add title if provided
	if c.String("title") != "" {
		vars["title"] = c.String("title")
	}

	// Add custom variables
	for _, varStr := range c.StringSlice("vars") {
		parts := strings.SplitN(varStr, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}

	// Process template
	title, content, tags := sncli.ProcessTemplate(tpl, vars)

	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Create note
	addNoteInput := sncli.AddNoteInput{
		Session: &session,
		Title:   title,
		Text:    content,
		Tags:    tags,
		Debug:   opts.debug,
	}

	if err = addNoteInput.Run(); err != nil {
		return fmt.Errorf("failed to create note from template: %w", err)
	}

	pterm.Success.Printf("Note created from template '%s'\n", templateName)
	pterm.Info.Printf("Title: %s\n", title)
	if len(tags) > 0 {
		pterm.Info.Printf("Tags: %s\n", strings.Join(tags, ", "))
	}

	return nil
}
