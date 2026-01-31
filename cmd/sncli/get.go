package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func cmdGet() *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "get items",
		BashComplete: func(c *cli.Context) {
			addTasks := []string{"tag", "note"}
			if c.NArg() > 0 {
				return
			}
			for _, t := range addTasks {
				fmt.Println(t)
			}
		},
		Subcommands: []*cli.Command{
			{
				Name:    "settings",
				Aliases: []string{"setting"},
				Usage:   "get settings",
				Hidden:  true,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "count",
						Usage: "useStdOut count only",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: "json",
						Usage: "output format",
					},
				},
				OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
					return err
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					// useStdOut = opts.useStdOut

					var matchAny bool
					if c.Bool("match-all") {
						matchAny = false
					}

					count := c.Bool("count")

					getSettingssIF := items.ItemFilters{
						MatchAny: matchAny,
						Filters: []items.Filter{
							{Type: "Setting"},
						},
					}

					sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
					if err != nil {
						return err
					}
					ss := sess.Gosn()
					// sync to get keys
					gsi := items.SyncInput{
						Session: &ss,
					}
					_, err = items.Sync(gsi)
					if err != nil {
						return err
					}
					var cacheDBPath string
					cacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
					if err != nil {
						return err
					}

					sess.CacheDBPath = cacheDBPath

					// TODO: validate output
					output := c.String("output")
					appGetSettingsConfig := sncli.GetSettingsConfig{
						Session: &sess,
						Filters: getSettingssIF,
						Output:  output,
						Debug:   opts.debug,
					}
					var rawSettings items.Items
					rawSettings, err = appGetSettingsConfig.Run()
					if err != nil {
						return err
					}
					var settingsYAML []sncli.SettingYAML
					var settingsJSON []sncli.SettingJSON
					var numResults int
					for _, rt := range rawSettings {
						numResults++
						if !count && sncli.StringInSlice(output, yamlAbbrevs, false) {
							tagContentOrgStandardNotesSNDetailYAML := sncli.OrgStandardNotesSNDetailYAML{
								ClientUpdatedAt: rt.(*items.Component).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
							}
							tagContentAppDataContent := sncli.AppDataContentYAML{
								OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailYAML,
							}

							settingContentYAML := sncli.SettingContentYAML{
								Title:          rt.(*items.Component).Content.GetTitle(),
								ItemReferences: sncli.ItemRefsToYaml(rt.(*items.Component).Content.References()),
								AppData:        tagContentAppDataContent,
							}

							settingsYAML = append(settingsYAML, sncli.SettingYAML{
								UUID:        rt.(*items.Component).UUID,
								ContentType: rt.(*items.Component).ContentType,
								Content:     settingContentYAML,
								UpdatedAt:   rt.(*items.Component).UpdatedAt,
								CreatedAt:   rt.(*items.Component).CreatedAt,
							})
						}
						if !count && strings.ToLower(output) == "json" {
							settingContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
								ClientUpdatedAt: rt.(*items.Component).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
							}
							settingContentAppDataContent := sncli.AppDataContentJSON{
								OrgStandardNotesSN: settingContentOrgStandardNotesSNDetailJSON,
							}

							settingContentJSON := sncli.SettingContentJSON{
								Title:          rt.(*items.Component).Content.GetTitle(),
								ItemReferences: sncli.ItemRefsToJSON(rt.(*items.Component).Content.References()),
								AppData:        settingContentAppDataContent,
							}

							settingsJSON = append(settingsJSON, sncli.SettingJSON{
								UUID:        rt.(*items.Component).UUID,
								ContentType: rt.(*items.Component).ContentType,
								Content:     settingContentJSON,
								UpdatedAt:   rt.(*items.Component).UpdatedAt,
								CreatedAt:   rt.(*items.Component).CreatedAt,
							})
						}
					}
					// if numResults <= 0 {
					// 	if count {
					// 		msg = "0"
					// 	} else {
					// 		msg = msgNoMatches
					// 	}
					// } else if count {
					// 	msg = strconv.Itoa(numResults)
					// } else {
					output = c.String("output")
					var bOutput []byte
					switch strings.ToLower(output) {
					case "json":
						bOutput, err = json.MarshalIndent(settingsJSON, "", "    ")
					case "yaml":
						bOutput, err = yaml.Marshal(settingsYAML)
					}
					if len(bOutput) > 0 {
						fmt.Println(string(bOutput))
					}

					return err
				},
			},
			{
				Name:    "tag",
				Aliases: []string{"tags"},
				Usage:   "get tags",
				BashComplete: func(c *cli.Context) {
					tagTasks := []string{
						"--title", "--uuid", "--regex", "--match-all", "--count", "--output",
					}
					if c.NArg() > 0 {
						return
					}
					for _, t := range tagTasks {
						fmt.Println(t)
					}
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "title",
						Usage: "find by title (separate multiple by commas)",
					},
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "find by uuid (separate multiple by commas)",
					},
					&cli.BoolFlag{
						Name:  "regex",
						Usage: "enable regular expressions",
					},
					&cli.BoolFlag{
						Name:  "match-all",
						Usage: "match all search criteria (default: match any)",
					},
					&cli.BoolFlag{
						Name:  "count",
						Usage: "useStdOut count only",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: "json",
						Usage: "output format",
					},
				},
				OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
					return err
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					return processGetTags(c, opts)
				},
			},
			{
				Name:    "note",
				Aliases: []string{"notes"},
				Usage:   "get notes",
				BashComplete: func(c *cli.Context) {
					addTasks := []string{"--title", "--text", "--tag", "--uuid", "--editor", "--include-trash", "--count"}
					if c.NArg() > 0 {
						return
					}
					for _, t := range addTasks {
						fmt.Println(t)
					}
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "title",
						Usage: "find by title",
					},
					&cli.StringFlag{
						Name:  "text",
						Usage: "find by text",
					},
					&cli.StringFlag{
						Name:  "tag",
						Usage: "find by tag",
					},
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "find by uuid",
					},
					&cli.StringFlag{
						Name:  "editor",
						Usage: "find by associated editor",
					},
					&cli.BoolFlag{
						Name:  "include-trash",
						Usage: "include notes in trash",
					},
					&cli.BoolFlag{
						Name:  "count",
						Usage: "number of notes",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: "json",
						Usage: "output format (json, yaml, table, rich)",
					},
					&cli.BoolFlag{
						Name:    "rich",
						Aliases: []string{"r"},
						Usage:   "display notes with rich markdown formatting",
					},
					&cli.BoolFlag{
						Name:    "preview",
						Aliases: []string{"p"},
						Usage:   "show preview in table view",
					},
					&cli.BoolFlag{
						Name:  "metadata",
						Usage: "show metadata in rich view",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)
					// useStdOut = opts.useStdOut
					return processGetNotes(c, opts)
				},
			},
			{
				Name:    "item",
				Aliases: []string{"items"},
				Usage:   "get any standard notes item",
				BashComplete: func(c *cli.Context) {
					getItemOpts := []string{"--uuid"}
					if c.NArg() > 0 {
						return
					}
					for _, t := range getItemOpts {
						fmt.Println(t)
					}
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "unique id of item to return (separate multiple with commas)",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: "json",
						Usage: "output format",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					// useStdOut = opts.useStdOut

					return processGetItems(c, opts)
				},
			},
			{
				Name:    "trash",
				Aliases: []string{"trashed"},
				Usage:   "get notes in trash",
				Hidden:  true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "title",
						Usage: "find by title",
					},
					&cli.StringFlag{
						Name:  "text",
						Usage: "find by text",
					},
					&cli.StringFlag{
						Name:  "tag",
						Usage: "find by tag",
					},
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "find by uuid",
					},
					&cli.BoolFlag{
						Name:  "count",
						Usage: "useStdOut countonly",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: "json",
						Usage: "output format",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)
					// useStdOut = opts.useStdOut

					return processGetTrash(c, opts)
				},
			},
		},
	}
}
