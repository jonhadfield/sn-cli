package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jonhadfield/gosn-v2/cache"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	sncli "github.com/jonhadfield/sn-cli"
	"gopkg.in/yaml.v2"

	"github.com/jonhadfield/gosn-v2"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	msgAddSuccess      = "Added"
	msgAlreadyExisting = "Already existing"
	msgDeleted         = "Deleted"
	msgCreateSuccess   = "Created"
	msgRegisterSuccess = "Registered"
	msgTagSuccess      = "Tagged"
	msgItemsDeleted    = "Items deleted"
	msgNoMatches       = "No matches"
	snAppName          = "sn-cli"
)

var yamlAbbrevs = []string{"yml", "yaml"}

// overwritten at build time.
var version, versionOutput, tag, sha, buildDate string

func main() {
	msg, display, err := startCLI(os.Args)
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		os.Exit(1)
	}

	if display && msg != "" {
		fmt.Println(msg)
	}

	os.Exit(0)
}

type configOptsOutput struct {
	useStdOut  bool
	useSession bool
	sessKey    string
	server     string
	cacheDBDir string
	debug      bool
}

func getOpts(c *cli.Context) (out configOptsOutput, err error) {
	if !c.GlobalBool("no-stdout") {
		out.useStdOut = true
	}

	if c.GlobalBool("use-session") || viper.GetBool("use_session") {
		out.useSession = true
	}

	out.sessKey = c.GlobalString("session-key")

	out.server = c.GlobalString("server")
	if viper.GetString("server") != "" {
		out.server = viper.GetString("server")
	}

	out.cacheDBDir = viper.GetString("cachedb_dir")
	if out.cacheDBDir != "" {
		out.cacheDBDir = c.GlobalString("cachedb-dir")
	}

	if c.GlobalBool("debug") {
		out.debug = true
	}

	return
}

func startCLI(args []string) (msg string, useStdOut bool, err error) {
	viper.SetEnvPrefix("sn")

	err = viper.BindEnv("email")
	if err != nil {
		return "", false, err
	}

	err = viper.BindEnv("password")
	if err != nil {
		return "", false, err
	}

	err = viper.BindEnv("server")
	if err != nil {
		return "", false, err
	}

	err = viper.BindEnv("cachedb_dir")
	if err != nil {
		return "", false, err
	}
	err = viper.BindEnv("use_session")
	if err != nil {
		return "", false, err
	}

	if tag != "" && buildDate != "" {
		versionOutput = fmt.Sprintf("[%s-%s] %s UTC", tag, sha, buildDate)
	} else {
		versionOutput = version
	}

	app := cli.NewApp()
	app.EnableBashCompletion = true

	app.Name = "sn"
	app.Version = versionOutput
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Jon Hadfield",
			Email: "jon@lessknown.co.uk",
		},
	}
	app.HelpName = "-"
	app.Usage = "Standard Notes CLI"
	app.Description = ""

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug"},
		cli.StringFlag{Name: "server"},
		cli.BoolFlag{Name: "use-session"},
		cli.StringFlag{Name: "session-key"},
		cli.BoolFlag{Name: "no-stdout"},
		cli.StringFlag{Name: "cachedb-dir", Value: viper.GetString("cachedb_dir")},
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		_, _ = fmt.Fprintf(c.App.Writer, "\ninvalid command: \"%s\" \n\n", command)
		cli.ShowAppHelpAndExit(c, 1)
	}
	app.Commands = []cli.Command{
		{
			Name:  "add",
			Usage: "add items",
			BashComplete: func(c *cli.Context) {
				addTasks := []string{"tag", "note"}
				if c.NArg() > 0 {
					return
				}
				for _, t := range addTasks {
					fmt.Println(t)
				}
			},
			Subcommands: []cli.Command{
				{
					Name:  "tag",
					Usage: "add tags",
					BashComplete: func(c *cli.Context) {
						if c.NArg() > 0 {
							return
						}
						fmt.Println("--title")
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "new tag title (separate multiple with commas)",
						},
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						// validate input
						tagInput := c.String("title")
						if strings.TrimSpace(tagInput) == "" {
							if cErr := cli.ShowSubcommandHelp(c); cErr != nil {
								panic(cErr)
							}
							return errors.New("tag title not defined")
						}

						// get session
						session, _, err := cache.GetSession(opts.useSession,
							opts.sessKey, opts.server)
						if err != nil {
							return err
						}

						session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
						if err != nil {
							return err
						}

						// prepare input
						tags := sncli.CommaSplit(tagInput)
						addTagInput := sncli.AddTagsInput{
							Session: session,
							Tags:    tags,
							Debug:   opts.debug,
						}

						// attempt to add tags
						var ato sncli.AddTagsOutput
						ato, err = addTagInput.Run()
						if err != nil {
							return fmt.Errorf(sncli.Red(err))
						}

						// present results
						if len(ato.Added) > 0 {
							msg = sncli.Green(msgAddSuccess+": ", strings.Join(ato.Added, ", "))
						}
						if len(ato.Existing) > 0 {
							// add line break if output already added
							if len(msg) > 0 {
								msg += "\n"
							}
							msg += sncli.Yellow(msgAlreadyExisting + ": " + strings.Join(ato.Existing, ", "))
						}

						return nil
					},
				},
				{
					Name:  "note",
					Usage: "add a note",
					BashComplete: func(c *cli.Context) {
						addNoteOpts := []string{"--title", "--text", "--tag", "--replace"}
						if c.NArg() > 0 {
							return
						}
						for _, ano := range addNoteOpts {
							fmt.Println(ano)
						}
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "new note title",
						},
						cli.StringFlag{
							Name:  "text",
							Usage: "new note text",
						},
						cli.StringFlag{
							Name:  "tag",
							Usage: "associate with tag (separate multiple with commas)",
						},
						cli.BoolFlag{
							Name:  "replace",
							Usage: "replace note with same title",
						},
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						// get input
						title := c.String("title")
						text := c.String("text")
						if strings.TrimSpace(title) == "" {
							if cErr := cli.ShowSubcommandHelp(c); cErr != nil {
								panic(cErr)
							}

							return errors.New("note title not defined")
						}
						if strings.TrimSpace(text) == "" {
							if cErr := cli.ShowSubcommandHelp(c); cErr != nil {
								panic(cErr)
							}

							return errors.New("note text not defined")
						}

						// get session
						session, _, err := cache.GetSession(opts.useSession,
							opts.sessKey, opts.server)
						if err != nil {
							return err
						}

						processedTags := sncli.CommaSplit(c.String("tag"))

						var cacheDBPath string
						cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
						if err != nil {
							return err
						}
						session.CacheDBPath = cacheDBPath
						AddNoteInput := sncli.AddNoteInput{
							Session: session,
							Title:   title,
							Text:    text,
							Tags:    processedTags,
							Replace: false,
							Debug:   opts.debug,
						}
						if err = AddNoteInput.Run(); err != nil {
							return fmt.Errorf("failed to add note. %+v", err)
						}

						msg = sncli.Green(msgAddSuccess + " note")

						return nil
					},
				},
			},
		},
		{
			Name:  "delete",
			Usage: "delete items",
			BashComplete: func(c *cli.Context) {
				addTasks := []string{"tag", "note"}
				if c.NArg() > 0 {
					return
				}
				for _, t := range addTasks {
					fmt.Println(t)
				}
			},
			Subcommands: []cli.Command{
				{
					Name:  "tag",
					Usage: "delete tag",
					BashComplete: func(c *cli.Context) {
						delTagOpts := []string{"--title", "--uuid"}
						if c.NArg() > 0 {
							return
						}
						for _, t := range delTagOpts {
							fmt.Println(t)
						}
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "title of note to delete (separate multiple with commas)",
						},
						cli.StringFlag{
							Name:  "uuid",
							Usage: "unique id of note to delete (separate multiple with commas)",
						},
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}

						useStdOut = opts.useStdOut
						msg, err = processDeleteTags(c, opts)
						if err != nil {
							return err
						}
						return nil
					},
				},
				{
					Name:  "note",
					Usage: "delete note",
					BashComplete: func(c *cli.Context) {
						delNoteOpts := []string{"--title", "--uuid"}
						if c.NArg() > 0 {
							return
						}
						for _, t := range delNoteOpts {
							fmt.Println(t)
						}
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "title of note to delete (separate multiple with commas)",
						},
						cli.StringFlag{
							Name:  "uuid",
							Usage: "unique id of note to delete (separate multiple with commas)",
						},
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						msg, err = processDeleteNote(c, opts)
						if err != nil {
							return err
						}

						return nil
					},
				},
			},
		},
		{
			Name:  "tag",
			Usage: "tag items",

			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "find-title",
					Usage: "match title",
				},
				cli.StringFlag{
					Name:  "find-text",
					Usage: "match text",
				},
				cli.StringFlag{
					Name:  "find-tag",
					Usage: "match tag",
				},
				cli.StringFlag{
					Name:  "title",
					Usage: "tag title to apply (separate multiple with commas)",
				},
				cli.BoolFlag{
					Name:  "purge",
					Usage: "delete other existing tags",
				},
				cli.BoolFlag{
					Name:  "ignore-case",
					Usage: "ignore case when matching",
				},
			},
			Action: func(c *cli.Context) error {
				opts, err := getOpts(c)
				if err != nil {
					return err
				}
				useStdOut = opts.useStdOut

				msg, err = processTagItems(c, opts)
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "get",
			Usage: "get items",
			BashComplete: func(c *cli.Context) {
				addTasks := []string{"tag", "note", "settings"}
				if c.NArg() > 0 {
					return
				}
				for _, t := range addTasks {
					fmt.Println(t)
				}
			},
			Subcommands: []cli.Command{
				{
					Name:    "settings",
					Aliases: []string{"setting"},
					Usage:   "get settings",
					Hidden:  true,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "count",
							Usage: "useStdOut count only",
						},
						cli.StringFlag{
							Name:  "output",
							Value: "json",
							Usage: "output format",
						},
					},
					OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
						return err
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}

						useStdOut = opts.useStdOut

						var matchAny bool
						if c.Bool("match-all") {
							matchAny = false
						}
						//regex := c.Bool("regex")
						count := c.Bool("count")

						getSettingssIF := gosn.ItemFilters{
							MatchAny: matchAny,
							Filters: []gosn.Filter{
								{Type: "Setting"},
							},
						}

						session, _, err := cache.GetSession(opts.useSession,
							opts.sessKey, opts.server)
						if err != nil {
							return err
						}

						var cacheDBPath string
						cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
						if err != nil {
							return err
						}

						session.CacheDBPath = cacheDBPath

						// TODO: validate output
						output := c.String("output")
						appGetSettingsConfig := sncli.GetSettingsConfig{
							Session: session,
							Filters: getSettingssIF,
							Output:  output,
							Debug:   opts.debug,
						}
						var rawSettings gosn.Items
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
									ClientUpdatedAt: rt.(*gosn.Component).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								tagContentAppDataContent := sncli.AppDataContentYAML{
									OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailYAML,
								}

								settingContentYAML := sncli.SettingContentYAML{
									Title:          rt.(*gosn.Component).Content.GetTitle(),
									ItemReferences: sncli.ItemRefsToYaml(rt.(*gosn.Component).Content.References()),
									AppData:        tagContentAppDataContent,
								}

								settingsYAML = append(settingsYAML, sncli.SettingYAML{
									UUID:        rt.(*gosn.Component).UUID,
									ContentType: rt.(*gosn.Component).ContentType,
									Content:     settingContentYAML,
									UpdatedAt:   rt.(*gosn.Component).UpdatedAt,
									CreatedAt:   rt.(*gosn.Component).CreatedAt,
								})
							}
							if !count && strings.ToLower(output) == "json" {
								settingContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
									ClientUpdatedAt: rt.(*gosn.Component).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								settingContentAppDataContent := sncli.AppDataContentJSON{
									OrgStandardNotesSN: settingContentOrgStandardNotesSNDetailJSON,
								}

								settingContentJSON := sncli.SettingContentJSON{
									Title:          rt.(*gosn.Component).Content.GetTitle(),
									ItemReferences: sncli.ItemRefsToJSON(rt.(*gosn.Component).Content.References()),
									AppData:        settingContentAppDataContent,
								}

								settingsJSON = append(settingsJSON, sncli.SettingJSON{
									UUID:        rt.(*gosn.Component).UUID,
									ContentType: rt.(*gosn.Component).ContentType,
									Content:     settingContentJSON,
									UpdatedAt:   rt.(*gosn.Component).UpdatedAt,
									CreatedAt:   rt.(*gosn.Component).CreatedAt,
								})
							}
						}
						if numResults <= 0 {
							if count {
								msg = "0"
							} else {
								msg = msgNoMatches
							}
						} else if count {
							msg = strconv.Itoa(numResults)
						} else {
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
						}
						return err
					},
				},
				{
					Name:    "tag",
					Aliases: []string{"tags"},
					Usage:   "get tags",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "find by title (separate multiple by commas)",
						},
						cli.StringFlag{
							Name:  "uuid",
							Usage: "find by uuid (separate multiple by commas)",
						},
						cli.BoolFlag{
							Name:  "regex",
							Usage: "enable regular expressions",
						},
						cli.BoolFlag{
							Name:  "match-all",
							Usage: "match all search criteria (default: match any)",
						},
						cli.BoolFlag{
							Name:  "count",
							Usage: "useStdOut count only",
						},
						cli.StringFlag{
							Name:  "output",
							Value: "json",
							Usage: "output format",
						},
					},
					OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
						return err
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						inTitle := strings.TrimSpace(c.String("title"))
						inUUID := strings.TrimSpace(c.String("uuid"))

						var matchAny bool
						if c.Bool("match-all") {
							matchAny = false
						}
						regex := c.Bool("regex")
						count := c.Bool("count")

						getTagsIF := gosn.ItemFilters{
							MatchAny: matchAny,
						}

						// add uuid filters
						if inUUID != "" {
							for _, uuid := range sncli.CommaSplit(inUUID) {
								titleFilter := gosn.Filter{
									Type:       "Tag",
									Key:        "uuid",
									Comparison: "==",
									Value:      uuid,
								}
								getTagsIF.Filters = append(getTagsIF.Filters, titleFilter)
							}
						}

						comparison := "contains"
						if regex {
							comparison = "~"
						}

						if inTitle != "" {
							for _, title := range sncli.CommaSplit(inTitle) {
								titleFilter := gosn.Filter{
									Type:       "Tag",
									Key:        "Title",
									Comparison: comparison,
									Value:      title,
								}
								getTagsIF.Filters = append(getTagsIF.Filters, titleFilter)
							}
						}

						if inTitle == "" && inUUID == "" {
							getTagsIF.Filters = append(getTagsIF.Filters, gosn.Filter{
								Type: "Tag",
							})
						}

						session, _, err := cache.GetSession(opts.useSession,
							opts.sessKey, opts.server)
						if err != nil {
							return err
						}

						// TODO: validate output
						output := c.String("output")

						var cacheDBPath string
						cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
						if err != nil {
							return err
						}
						session.CacheDBPath = cacheDBPath

						appGetTagConfig := sncli.GetTagConfig{
							Session: session,
							Filters: getTagsIF,
							Output:  output,
							Debug:   opts.debug,
						}
						var rawTags gosn.Items
						rawTags, err = appGetTagConfig.Run()
						if err != nil {
							return err
						}

						// strip deleted items
						rawTags = sncli.RemoveDeleted(rawTags)

						var tagsYAML []sncli.TagYAML
						var tagsJSON []sncli.TagJSON
						var numResults int
						for _, rt := range rawTags {
							numResults++
							if !count && sncli.StringInSlice(output, yamlAbbrevs, false) {
								tagContentOrgStandardNotesSNDetailYAML := sncli.OrgStandardNotesSNDetailYAML{
									ClientUpdatedAt: rt.(*gosn.Tag).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								tagContentAppDataContent := sncli.AppDataContentYAML{
									OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailYAML,
								}

								tagContentYAML := sncli.TagContentYAML{
									Title:          rt.(*gosn.Tag).Content.GetTitle(),
									ItemReferences: sncli.ItemRefsToYaml(rt.(*gosn.Tag).Content.References()),
									AppData:        tagContentAppDataContent,
								}

								tagsYAML = append(tagsYAML, sncli.TagYAML{
									UUID:        rt.(*gosn.Tag).UUID,
									ContentType: rt.(*gosn.Tag).ContentType,
									Content:     tagContentYAML,
									UpdatedAt:   rt.(*gosn.Tag).UpdatedAt,
									CreatedAt:   rt.(*gosn.Tag).CreatedAt,
								})
							}
							if !count && strings.ToLower(output) == "json" {
								tagContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
									ClientUpdatedAt: rt.(*gosn.Tag).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								tagContentAppDataContent := sncli.AppDataContentJSON{
									OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailJSON,
								}

								tagContentJSON := sncli.TagContentJSON{
									Title:          rt.(*gosn.Tag).Content.GetTitle(),
									ItemReferences: sncli.ItemRefsToJSON(rt.(*gosn.Tag).Content.References()),
									AppData:        tagContentAppDataContent,
								}

								tagsJSON = append(tagsJSON, sncli.TagJSON{
									UUID:        rt.(*gosn.Tag).UUID,
									ContentType: rt.(*gosn.Tag).ContentType,
									Content:     tagContentJSON,
									UpdatedAt:   rt.(*gosn.Tag).UpdatedAt,
									CreatedAt:   rt.(*gosn.Tag).CreatedAt,
								})
							}
						}
						if numResults <= 0 {
							if count {
								msg = "0"
							} else {
								msg = msgNoMatches
							}
						} else if count {
							msg = strconv.Itoa(numResults)
						} else {
							output = c.String("output")
							var bOutput []byte
							switch strings.ToLower(output) {
							case "json":
								bOutput, err = json.MarshalIndent(tagsJSON, "", "    ")
							case "yaml":
								bOutput, err = yaml.Marshal(tagsYAML)
							}
							if len(bOutput) > 0 {
								fmt.Println(string(bOutput))
							}
						}
						return err
					},
				},
				{
					Name:    "note",
					Aliases: []string{"notes"},
					Usage:   "get notes",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "find by title",
						},
						cli.StringFlag{
							Name:  "text",
							Usage: "find by text",
						},
						cli.StringFlag{
							Name:  "tag",
							Usage: "find by tag",
						},
						cli.StringFlag{
							Name:  "uuid",
							Usage: "find by uuid",
						},
						cli.BoolFlag{
							Name:  "count",
							Usage: "useStdOut countonly",
						},
						cli.StringFlag{
							Name:  "output",
							Value: "json",
							Usage: "output format",
						},
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						uuid := c.String("uuid")
						title := c.String("title")
						text := c.String("text")
						count := c.Bool("count")
						output := c.String("output")
						noteFilter := gosn.Filter{
							Type: "Note",
						}
						getNotesIF := gosn.ItemFilters{
							MatchAny: false,
							Filters:  []gosn.Filter{noteFilter},
						}

						if uuid != "" {
							titleFilter := gosn.Filter{
								Type:       "Note",
								Key:        "uuid",
								Comparison: "==",
								Value:      uuid,
							}
							getNotesIF.Filters = append(getNotesIF.Filters, titleFilter)
						}
						if title != "" {
							titleFilter := gosn.Filter{
								Type:       "Note",
								Key:        "Title",
								Comparison: "contains",
								Value:      title,
							}
							getNotesIF.Filters = append(getNotesIF.Filters, titleFilter)
						}
						if text != "" {
							titleFilter := gosn.Filter{
								Type:       "Note",
								Key:        "Text",
								Comparison: "contains",
								Value:      text,
							}
							getNotesIF.Filters = append(getNotesIF.Filters, titleFilter)
						}
						processedTags := sncli.CommaSplit(c.String("tag"))

						if len(processedTags) > 0 {
							for _, t := range processedTags {
								titleFilter := gosn.Filter{
									Type:       "Note",
									Key:        "Tag",
									Comparison: "contains",
									Value:      t,
								}
								getNotesIF.Filters = append(getNotesIF.Filters, titleFilter)
							}
						}

						session, _, err := cache.GetSession(opts.useSession,
							opts.sessKey, opts.server)
						if err != nil {
							return err
						}

						var cacheDBPath string
						cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
						if err != nil {
							return err
						}

						session.CacheDBPath = cacheDBPath

						getNoteConfig := sncli.GetNoteConfig{
							Session: session,
							Filters: getNotesIF,
							Debug:   opts.debug,
						}

						var rawNotes gosn.Items
						rawNotes, err = getNoteConfig.Run()
						if err != nil {
							return err
						}
						// strip deleted items
						rawNotes = sncli.RemoveDeleted(rawNotes)

						var numResults int
						var notesYAML []sncli.NoteYAML
						var notesJSON []sncli.NoteJSON
						for _, rt := range rawNotes {
							numResults++
							if !count && sncli.StringInSlice(output, yamlAbbrevs, false) {
								noteContentOrgStandardNotesSNDetailYAML := sncli.OrgStandardNotesSNDetailYAML{
									ClientUpdatedAt: rt.(*gosn.Note).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								noteContentAppDataContent := sncli.AppDataContentYAML{
									OrgStandardNotesSN: noteContentOrgStandardNotesSNDetailYAML,
								}
								noteContentYAML := sncli.NoteContentYAML{
									Title:          rt.(*gosn.Note).Content.GetTitle(),
									Text:           rt.(*gosn.Note).Content.GetText(),
									ItemReferences: sncli.ItemRefsToYaml(rt.(*gosn.Note).Content.References()),
									AppData:        noteContentAppDataContent,
								}

								notesYAML = append(notesYAML, sncli.NoteYAML{
									UUID:        rt.(*gosn.Note).UUID,
									ContentType: rt.(*gosn.Note).ContentType,
									Content:     noteContentYAML,
									UpdatedAt:   rt.(*gosn.Note).UpdatedAt,
									CreatedAt:   rt.(*gosn.Note).CreatedAt,
								})
							}
							if !count && strings.ToLower(output) == "json" {
								noteContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
									ClientUpdatedAt: rt.(*gosn.Note).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								noteContentAppDataContent := sncli.AppDataContentJSON{
									OrgStandardNotesSN: noteContentOrgStandardNotesSNDetailJSON,
								}
								noteContentJSON := sncli.NoteContentJSON{
									Title:          rt.(*gosn.Note).Content.GetTitle(),
									Text:           rt.(*gosn.Note).Content.GetText(),
									ItemReferences: sncli.ItemRefsToJSON(rt.(*gosn.Note).Content.References()),
									AppData:        noteContentAppDataContent,
								}

								notesJSON = append(notesJSON, sncli.NoteJSON{
									UUID:        rt.(*gosn.Note).UUID,
									ContentType: rt.(*gosn.Note).ContentType,
									Content:     noteContentJSON,
									UpdatedAt:   rt.(*gosn.Note).UpdatedAt,
									CreatedAt:   rt.(*gosn.Note).CreatedAt,
								})
							}
						}

						if numResults <= 0 {
							if count {
								msg = "0"
							} else {
								msg = msgNoMatches
							}
						} else if count {
							msg = strconv.Itoa(numResults)
						} else {
							output = c.String("output")
							var bOutput []byte
							switch strings.ToLower(output) {
							case "json":
								bOutput, err = json.MarshalIndent(notesJSON, "", "    ")
							case "yaml":
								bOutput, err = yaml.Marshal(notesYAML)
							}
							if len(bOutput) > 0 {
								fmt.Println(string(bOutput))
							}
						}
						return err
					},
				},
			},
		},
		{
			Name:  "export",
			Usage: "export data",
			Flags: []cli.Flag{
				cli.BoolTFlag{
					Name:  "encrypted (default: true)",
					Usage: "encrypt the exported data",
				},
				cli.StringFlag{
					Name:   "format",
					Usage:  "hidden whilst gob is the only supported format",
					Value:  "gob",
					Hidden: true,
				},
				cli.StringFlag{
					Name:  "output (default: current directory)",
					Usage: "output path",
				},
			},
			Action: func(c *cli.Context) error {
				opts, err := getOpts(c)
				if err != nil {
					return err
				}
				useStdOut = opts.useStdOut

				outputPath := strings.TrimSpace(c.String("output"))
				if outputPath == "" {
					currDir, err := os.Getwd()
					if err != nil {
						return err
					}
					timeStamp := time.Now().UTC().Format("20060102150405")
					filePath := fmt.Sprintf("standard_notes_export_%s.gob", timeStamp)
					outputPath = currDir + string(os.PathSeparator) + filePath
				}
				session, _, err := cache.GetSession(opts.useSession, opts.sessKey, opts.server)
				if err != nil {
					return err
				}

				var cacheDBPath string
				cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
				if err != nil {
					return err
				}

				session.CacheDBPath = cacheDBPath
				appExportConfig := sncli.ExportConfig{
					Session: session,
					//CacheDBPath: cacheDBPath,
					File:  outputPath,
					Debug: opts.debug,
				}
				err = appExportConfig.Run()
				if err == nil {
					msg = fmt.Sprintf("encrypted export written to: %s", outputPath)
				}
				return err
			},
		},
		{
			Name:  "import",
			Usage: "import data",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "format",
					Usage:  "hidden whilst gob is the only supported format",
					Value:  "gob",
					Hidden: true,
				},
				cli.StringFlag{
					Name:  "file",
					Usage: "path of file to import",
				},
			},
			Action: func(c *cli.Context) error {
				opts, err := getOpts(c)
				if err != nil {
					return err
				}

				inputPath := strings.TrimSpace(c.String("file"))
				if inputPath == "" {
					return errors.New("please specify path using --file")
				}

				session, _, err := cache.GetSession(opts.useSession, opts.sessKey, opts.server)
				if err != nil {
					return err
				}

				session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
				if err != nil {
					return err
				}

				appImportConfig := sncli.ImportConfig{
					Session: session,
					File:    inputPath,
					Debug:   opts.debug,
				}
				err = appImportConfig.Run()
				if err == nil {
					msg = fmt.Sprintf("import successful")
				}
				return err
			},
		},
		{
			Name:  "register",
			Usage: "register a new user",
			BashComplete: func(c *cli.Context) {
				if c.NArg() > 0 {
					return
				}
				fmt.Println("--email")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "email",
					Usage: "email address",
				},
			},
			Action: func(c *cli.Context) error {
				opts, err := getOpts(c)
				if err != nil {
					return err
				}
				useStdOut = opts.useStdOut

				if strings.TrimSpace(c.String("email")) == "" {
					if cErr := cli.ShowCommandHelp(c, "register"); cErr != nil {
						panic(cErr)
					}
					return errors.New("email required")
				}
				var password string
				fmt.Print("password: ")
				bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if err == nil {
					password = string(bytePassword)
				}
				if len(strings.TrimSpace(password)) == 0 {
					return errors.New("password required")
				}
				registerConfig := sncli.RegisterConfig{
					Email:     c.String("email"),
					Password:  password,
					APIServer: opts.server,
				}
				err = registerConfig.Run()
				if err != nil {
					return err
				}
				fmt.Println(msgRegisterSuccess)
				return nil
			},
		},
		{
			Name:  "stats",
			Usage: "show statistics",
			Action: func(c *cli.Context) error {
				opts, err := getOpts(c)
				if err != nil {
					return err
				}

				session, _, err := cache.GetSession(opts.useSession,
					opts.sessKey, opts.server)
				if err != nil {
					return err
				}

				session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
				if err != nil {
					return err
				}

				statsConfig := sncli.StatsConfig{
					Session: session,
					Debug:   opts.debug,
				}
				err = statsConfig.Run()
				return err
			},
		},
		{
			Name:  "wipe",
			Usage: "deletes all tags and notes",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "yes",
					Usage: "ignore warning",
				},
				cli.BoolFlag{
					Name:  "settings",
					Usage: "wipe settings also",
				},
			},
			Action: func(c *cli.Context) error {
				opts, err := getOpts(c)
				if err != nil {
					return err
				}
				useStdOut = opts.useStdOut

				session, email, err := cache.GetSession(opts.useSession, opts.sessKey, opts.server)
				if err != nil {

					return err
				}

				var cacheDBPath string
				cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
				if err != nil {
					return err
				}

				session.CacheDBPath = cacheDBPath
				wipeConfig := sncli.WipeConfig{
					Session:  session,
					Settings: c.Bool("settings"),
					Debug:    opts.debug,
				}
				var numWiped int

				var proceed bool
				if c.Bool("yes") {
					proceed = true
				} else {
					fmt.Printf("wipe all items for account %s? ", email)
					var input string
					_, err = fmt.Scanln(&input)
					if err == nil && sncli.StringInSlice(input, []string{"y", "yes"}, false) {
						proceed = true
					}
				}
				if proceed {
					numWiped, err = wipeConfig.Run()
					if err != nil {
						return err
					}
					msg = fmt.Sprintf("%d %s", numWiped, msgItemsDeleted)
				} else {
					return nil
				}
				return err
			},
		},
		{
			Name:  "session",
			Usage: "manage session credentials",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "add",
					Usage: "add session to keychain",
				},
				cli.BoolFlag{
					Name:  "remove",
					Usage: "remove session from keychain",
				},
				cli.BoolFlag{
					Name:  "status",
					Usage: "get session details",
				},
				cli.StringFlag{
					Name:     "session-key",
					Usage:    "[optional] key to encrypt/decrypt session",
					Required: false,
				},
			},
			Hidden: false,
			Action: func(c *cli.Context) error {
				opts, err := getOpts(c)
				if err != nil {
					return err
				}
				useStdOut = opts.useStdOut

				sAdd := c.Bool("add")
				sRemove := c.Bool("remove")
				sStatus := c.Bool("status")
				sessKey := c.String("session-key")
				if sStatus || sRemove {
					if err = gosn.SessionExists(nil); err != nil {
						return err
					}
				}
				nTrue := numTrue(sAdd, sRemove, sStatus)
				if nTrue == 0 || nTrue > 1 {
					_ = cli.ShowCommandHelp(c, "session")
					os.Exit(1)
				}
				if sAdd {
					msg, err = gosn.AddSession(opts.server, sessKey, nil)
					return err
				}
				if sRemove {
					msg = gosn.RemoveSession(nil)
					return nil
				}
				if sStatus {
					msg, err = gosn.SessionStatus(sessKey, nil)
				}
				return err
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))

	return msg, useStdOut, app.Run(args)
}

func numTrue(in ...bool) (total int) {
	for _, i := range in {
		if i {
			total++
		}
	}

	return
}
