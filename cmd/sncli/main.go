package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

const (
	msgAddSuccess      = "added"
	msgAlreadyExisting = "already existing"
	msgDeleted         = "deleted"
	msgRegisterSuccess = "registered"
	msgTagSuccess      = "tagged"
	msgItemsDeleted    = "items deleted"
	msgNoMatches       = "no matches"
	snAppName          = "sn-cli"
)

var yamlAbbrevs = []string{"yml", "yaml"}

// overwritten at build time.
var version, versionOutput, tag, sha, buildDate string

func main() {
	msg, display, err := startCLI(os.Args)
	if err != nil {
		fmt.Println("err:", err)
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
	out.useStdOut = true

	if c.GlobalBool("no-stdout") {
		out.useStdOut = false
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
		versionOutput = fmt.Sprintf("[%s-%s] %s", tag, sha, buildDate)
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
		cli.BoolFlag{Name: "no-stdout", Hidden: true},
		cli.StringFlag{Name: "cachedb-dir", Value: viper.GetString("cachedb_dir")},
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		_, _ = fmt.Fprintf(c.App.Writer, "\ninvalid command: \"%s\" \n\n", command)
		cli.ShowAppHelpAndExit(c, 1)
	}
	app.Commands = []cli.Command{
		{
			Name:  "edit",
			Usage: "edit items",
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
					Usage: "edit a tag",
					BashComplete: func(c *cli.Context) {
						addNoteOpts := []string{"--title", "--uuid"}
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
							Usage: "title of the tag",
						},
						cli.StringFlag{
							Name:  "uuid",
							Usage: "uuid of the tag",
						},
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						msg, err = processEditTag(c, opts)

						return err
					},
				},
				{
					Name:  "note",
					Usage: "edit a note",
					BashComplete: func(c *cli.Context) {
						addNoteOpts := []string{"--title", "--uuid", "--editor"}
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
							Usage: "title of the note",
						},
						cli.StringFlag{
							Name:  "uuid",
							Usage: "uuid of the note",
						},
						cli.StringFlag{
							Name:   "editor",
							Usage:  "path to editor",
							EnvVar: "EDITOR",
						},
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						msg, err = processEditNote(c, opts)

						return err
					},
				},
			},
		},
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
						fmt.Println("--title", "--parent", "--parent-uuid")
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "new tag title (separate multiple with commas)",
						},
						cli.StringFlag{
							Name:  "parent",
							Usage: "parent tag title to make a sub-tag of",
						},
						cli.StringFlag{
							Name:  "parent-uuid",
							Usage: "parent tag uuid to make a sub-tag of",
						},
					},
					Action: func(c *cli.Context) error {
						var opts configOptsOutput
						opts, err = getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut
						msg, err = processAddTags(c, opts)

						return err
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
							Name:  "file",
							Usage: "path to file with note content (specify --title or leave blank to use filename)",
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

						msg, err = processAddNotes(c, opts)

						return err
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

						return err
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

						return err
					},
				},
				{
					Name:  "item",
					Usage: "delete any standard notes item",
					BashComplete: func(c *cli.Context) {
						delNoteOpts := []string{"--uuid"}
						if c.NArg() > 0 {
							return
						}
						for _, t := range delNoteOpts {
							fmt.Println(t)
						}
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "uuid",
							Usage: "unique id of item to delete (separate multiple with commas)",
						},
					},
					Action: func(c *cli.Context) error {
						opts, err := getOpts(c)
						if err != nil {
							return err
						}

						useStdOut = opts.useStdOut

						msg, err = processDeleteItems(c, opts)

						return err
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

				return err
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

						count := c.Bool("count")

						getSettingssIF := items.ItemFilters{
							MatchAny: matchAny,
							Filters: []items.Filter{
								{Type: "Setting"},
							},
						}

						sess, _, err := cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
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

						msg, err = processGetTags(c, opts)

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
							Name:  "include-trash",
							Usage: "include notes in trash",
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

						msg, err = processGetNotes(c, opts)

						return err
					},
				},
				{
					Name:    "trash",
					Aliases: []string{"trashed"},
					Usage:   "get notes in trash",
					Hidden:  true,
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

						msg, err = processGetTrash(c, opts)

						return err
					},
				},
			},
		},
		// {
		// 	Name:   "export",
		// 	Usage:  "export data",
		// 	Hidden: true,
		// 	Flags: []cli.Flag{
		// 		cli.StringFlag{
		// 			Name:  "path",
		// 			Usage: "choose directory to place export in (default: current directory)",
		// 		},
		// 	},
		// 	Action: func(c *cli.Context) error {
		// 		var opts configOptsOutput
		// 		opts, err = getOpts(c)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		useStdOut = opts.useStdOut
		//
		// 		outputPath := strings.TrimSpace(c.String("output"))
		// 		if outputPath == "" {
		// 			outputPath, err = os.Getwd()
		// 			if err != nil {
		// 				return err
		// 			}
		// 		}
		//
		// 		timeStamp := time.Now().UTC().Format("20060102150405")
		// 		filePath := fmt.Sprintf("standard_notes_export_%s.json", timeStamp)
		// 		outputPath += string(os.PathSeparator) + filePath
		//
		// 		var sess cache.Session
		// 		sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
		// 		if err != nil {
		// 			return err
		// 		}
		//
		// 		var cacheDBPath string
		// 		cacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
		// 		if err != nil {
		// 			return err
		// 		}
		//
		// 		sess.Debug = opts.debug
		//
		// 		sess.CacheDBPath = cacheDBPath
		// 		appExportConfig := sncli.ExportConfig{
		// 			Session:   &sess,
		// 			Decrypted: c.Bool("decrypted"),
		// 			File:      outputPath,
		// 		}
		// 		err = appExportConfig.Run()
		// 		if err == nil {
		// 			msg = fmt.Sprintf("encrypted export written to: %s", outputPath)
		// 		}
		//
		// 		return err
		// 	},
		// },
		// {
		// 	Name:   "import",
		// 	Usage:  "import data",
		// 	Hidden: true,
		// 	Flags: []cli.Flag{
		// 		cli.StringFlag{
		// 			Name:  "file",
		// 			Usage: "path of file to import",
		// 		},
		// 		cli.BoolFlag{
		// 			Name:  "experiment",
		// 			Usage: "test import functionality - only use after taking backup as this is experimental",
		// 		},
		// 	},
		// 	Action: func(c *cli.Context) error {
		// 		var opts configOptsOutput
		// 		opts, err = getOpts(c)
		// 		if err != nil {
		// 			return err
		// 		}
		//
		// 		useStdOut = opts.useStdOut
		//
		// 		inputPath := strings.TrimSpace(c.String("file"))
		// 		if inputPath == "" {
		// 			return errors.New("please specify path using --file")
		// 		}
		//
		// 		if !c.Bool("experiment") {
		// 			fmt.Printf("\nWARNING: The import functionality is currently for testing only\nDo not use unless you have a backup of your data and intend to restore after testing\nTo proceed run the command with flag --experiment\n")
		// 			return nil
		// 		}
		//
		// 		var session cache.Session
		// 		session, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
		// 		if err != nil {
		// 			return err
		// 		}
		//
		// 		session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
		// 		if err != nil {
		// 			return err
		// 		}
		//
		// 		appImportConfig := sncli.ImportConfig{
		// 			Session:   &session,
		// 			File:      inputPath,
		// 			Format:    c.String("format"),
		// 			Debug:     opts.debug,
		// 			UseStdOut: opts.useStdOut,
		// 		}
		//
		// 		var imported int
		// 		imported, err = appImportConfig.Run()
		// 		if err == nil {
		// 			msg = fmt.Sprintf("imported %d items", imported)
		// 		} else {
		// 			msg = "import failed"
		// 		}
		//
		// 		return err
		// 	},
		// },
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
				var opts configOptsOutput
				opts, err = getOpts(c)
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
				if password, err = getPassword(); err != nil {
					return err
				}

				registerConfig := sncli.RegisterConfig{
					Email:     c.String("email"),
					Password:  password,
					APIServer: opts.server,
					Debug:     opts.debug,
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
			Name:  "resync",
			Usage: "purge cache and resync content",
			Action: func(c *cli.Context) error {
				var opts configOptsOutput
				opts, err = getOpts(c)
				if err != nil {
					return err
				}
				var session cache.Session
				session, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
				if err != nil {
					return err
				}

				return sncli.Resync(&session, opts.cacheDBDir, snAppName)
			},
		},
		{
			Name:  "stats",
			Usage: "show statistics",
			Action: func(c *cli.Context) error {
				var opts configOptsOutput
				opts, err = getOpts(c)
				if err != nil {
					return err
				}
				var sess cache.Session
				sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
				if err != nil {
					return err
				}
				sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
				if err != nil {
					return err
				}
				statsConfig := sncli.StatsConfig{
					Session: sess,
				}

				return statsConfig.Run()
			},
		},
		{
			Name:  "wipe",
			Usage: "deletes all supported content",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "yes",
					Usage: "ignore warning",
				},
				cli.BoolFlag{
					Name:  "everything",
					Usage: "wipe settings also",
				},
			},
			Action: func(c *cli.Context) error {
				var opts configOptsOutput
				opts, err = getOpts(c)
				if err != nil {
					return err
				}
				useStdOut = opts.useStdOut

				var cacheSession cache.Session
				cacheSession, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
				if err != nil {
					return err
				}

				var cacheDBPath string
				cacheDBPath, err = cache.GenCacheDBPath(cacheSession, opts.cacheDBDir, snAppName)
				if err != nil {
					return err
				}

				cacheSession.CacheDBPath = cacheDBPath
				wipeConfig := sncli.WipeConfig{
					Session:    &cacheSession,
					UseStdOut:  useStdOut,
					Everything: c.Bool("everything"),
					Debug:      opts.debug,
				}
				var numWiped int

				var proceed bool
				if c.Bool("yes") {
					proceed = true
				} else {
					fmt.Printf("wipe all items for account %s? ", cacheSession.Session.KeyParams.Identifier)
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
				var opts configOptsOutput
				opts, err = getOpts(c)
				if err != nil {
					return err
				}
				useStdOut = opts.useStdOut

				msg, err = processSession(c, opts)

				return err
			},
		},
		{
			Name:   "healthcheck",
			Usage:  "find and fix account data errors",
			Hidden: true,
			BashComplete: func(c *cli.Context) {
				addTasks := []string{"keys"}
				if c.NArg() > 0 {
					return
				}
				for _, t := range addTasks {
					fmt.Println(t)
				}
			},
			Subcommands: []cli.Command{
				{
					Name:  "keys",
					Usage: "find issues relating to ItemsKeys",
					BashComplete: func(c *cli.Context) {
						hcKeysOpts := []string{"--delete-invalid"}
						if c.NArg() > 0 {
							return
						}
						for _, ano := range hcKeysOpts {
							fmt.Println(ano)
						}
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Hidden: true,
							Name:   "delete-invalid",
							Usage:  "delete items that cannot be decrypted",
						},
					},
					Action: func(c *cli.Context) error {
						var opts configOptsOutput
						opts, err = getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						var sess session.Session

						sess, _, err = session.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)

						if err != nil {
							return err
						}
						err = sncli.ItemKeysHealthcheck(sncli.ItemsKeysHealthcheckInput{
							Session:       sess,
							UseStdOut:     useStdOut,
							DeleteInvalid: c.Bool("delete-invalid"),
						})

						return err
					},
				},
			},
		},
		{
			Name:   "debug",
			Usage:  "debug tools",
			Hidden: true,
			BashComplete: func(c *cli.Context) {
				addTasks := []string{"decrypt-string"}
				if c.NArg() > 0 {
					return
				}
				for _, t := range addTasks {
					fmt.Println(t)
				}
			},
			Subcommands: []cli.Command{
				{
					Name:  "decrypt-string",
					Usage: "accepts a string in the format: <version>:<ciphertext>:<auth-data>, decrypts it using the session key (or one specified with --key) and returns the decrypted ciphertext",
					BashComplete: func(c *cli.Context) {
						hcKeysOpts := []string{"--key"}
						if c.NArg() > 0 {
							return
						}
						for _, ano := range hcKeysOpts {
							fmt.Println(ano)
						}
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "key",
							Usage: "override session's master key",
						},
					},
					Action: func(c *cli.Context) error {
						str := ""
						if c.Args().Present() {
							fmt.Printf("c.Args() %+v\n", c.Args())
							fmt.Printf("c.Args() %+v\n", c.Args().First())
							str = c.Args().First()
						}

						var opts configOptsOutput
						opts, err = getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						var sess session.Session

						sess, _, err = session.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
						if err != nil {
							return err
						}
						var res string
						res, err = sncli.DecryptString(sncli.DecryptStringInput{
							Session:   sess,
							UseStdOut: useStdOut,
							Key:       c.String("key"),
							In:        str,
						})
						if err != nil {
							return err
						}

						msg = fmt.Sprintf("plaintext: %s", res)

						return err
					},
				},
				{
					Name:  "output-session",
					Usage: "returns specified session items",
					BashComplete: func(c *cli.Context) {
						hcKeysOpts := []string{"--master-key"}
						if c.NArg() > 0 {
							return
						}
						for _, ano := range hcKeysOpts {
							fmt.Println(ano)
						}
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "master-key",
							Usage: "output master key",
						},
					},
					Action: func(c *cli.Context) error {
						var opts configOptsOutput
						opts, err = getOpts(c)
						if err != nil {
							return err
						}
						useStdOut = opts.useStdOut

						var sess session.Session

						sess, _, err = session.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)

						if err != nil {
							return err
						}
						err = sncli.OutputSession(sncli.OutputSessionInput{
							Session:         sess,
							UseStdOut:       useStdOut,
							OutputMasterKey: c.Bool("master-key"),
						})

						return err
					},
				},
				// {
				// 	Name:  "create-itemskey",
				// 	Usage: "creates and displays an items key without syncing",
				// 	BashComplete: func(c *cli.Context) {
				// 		hcKeysOpts := []string{"--master-key"}
				// 		if c.NArg() > 0 {
				// 			return
				// 		}
				// 		for _, ano := range hcKeysOpts {
				// 			fmt.Println(ano)
				// 		}
				// 	},
				// 	Flags: []cli.Flag{
				// 		cli.StringFlag{
				// 			Name:  "master-key",
				// 			Usage: "master key to encrypt the encrypted item key with",
				// 		},
				// 	},
				// 	Action: func(c *cli.Context) error {
				// 		var opts configOptsOutput
				// 		opts, err = getOpts(c)
				// 		if err != nil {
				// 			return err
				// 		}
				// 		useStdOut = opts.useStdOut
				//
				// 		return sncli.CreateItemsKey(sncli.CreateItemsKeyInput{
				// 			Debug:     opts.debug,
				// 			MasterKey: c.String("master-key"),
				// 		})
				// 	},
				// },
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	return msg, useStdOut, app.Run(args)
}

func getPassword() (res string, err error) {
	for {
		fmt.Print("password: ")
		var bytePassword []byte
		if bytePassword, err = term.ReadPassword(int(syscall.Stdin)); err != nil {
			return
		}

		if len(bytePassword) < sncli.MinPasswordLength {
			err = fmt.Errorf("\rpassword must be at least %d characters", sncli.MinPasswordLength)

			return
		}

		var bytePassword2 []byte
		fmt.Printf("\rconfirm password: ")
		if bytePassword2, err = term.ReadPassword(int(syscall.Stdin)); err != nil {
			return
		}

		if !bytes.Equal(bytePassword, bytePassword2) {
			fmt.Printf("\rpasswords do not match")
			fmt.Println()

			return
		}

		fmt.Println()
		if err == nil {
			res = string(bytePassword)

			return
		}
	}
}

func numTrue(in ...bool) (total int) {
	for _, i := range in {
		if i {
			total++
		}
	}

	return
}
