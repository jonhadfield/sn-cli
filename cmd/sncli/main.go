package main

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/jonhadfield/gosn"

	"github.com/jonhadfield/sncli"

	"fmt"

	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

// overwritten at build time
var version, versionOutput, tag, sha, buildDate string

func main() {
	err := startCLI(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func processTags(input string) []string {
	tags := strings.Split(input, ",")
	if len(tags) == 1 && len(tags[0]) == 0 {
		return nil
	}
	return tags
}

func processNotes(input string) []string {
	notes := strings.Split(input, ",")
	if len(notes) == 1 && len(notes[0]) == 0 {
		return nil
	}
	return notes
}

func startCLI(args []string) error {
	viper.SetEnvPrefix("sn")
	err := viper.BindEnv("email")
	if err != nil {
		return err
	}
	err = viper.BindEnv("password")
	if err != nil {
		return err
	}
	err = viper.BindEnv("server")
	if err != nil {
		return err
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
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "server"},
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		_, _ = fmt.Fprintf(c.App.Writer, "\ninvalid command: \"%s\" \n\n", command)
		cli.ShowAppHelpAndExit(c, 1)
	}
	app.Commands = []cli.Command{
		{
			Name:  "add",
			Usage: "add items",
			Subcommands: []cli.Command{
				{
					Name:  "tag",
					Usage: "add tags",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "new tag title (separate multiple with commas)",
						},
					},
					Action: func(c *cli.Context) error {
						newTags := c.String("title")
						if strings.TrimSpace(newTags) == "" {
							fmt.Print("\nerror: tag title not defined\n\n")
							return cli.ShowSubcommandHelp(c)
						}
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}

						processedTags := processTags(newTags)

						appAddTagConfig := sncli.AddTagConfig{
							Session: session,
							Tags:    processedTags,
							Debug:   c.GlobalBool("debug"),
						}
						return appAddTagConfig.Run()
					},
				},
				{
					Name:  "note",
					Usage: "add a note",
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
						title := c.String("title")
						text := c.String("text")
						if strings.TrimSpace(title) == "" {
							fmt.Print("\nerror: note title not defined\n\n")
							return cli.ShowSubcommandHelp(c)
						}
						if strings.TrimSpace(text) == "" {
							fmt.Print("\nerror: note text	 not defined\n\n")
							return cli.ShowSubcommandHelp(c)
						}

						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						processedTags := processTags(c.String("tag"))

						AddNoteConfig := sncli.AddNoteConfig{
							Session: session,
							Title:   title,
							Text:    text,
							Tags:    processedTags,
							Replace: false,
							Debug:   c.GlobalBool("debug"),
						}
						return AddNoteConfig.Run()
					},
				},
			},
		},
		{
			Name:  "delete",
			Usage: "delete items",
			Subcommands: []cli.Command{
				{
					Name:  "tag",
					Usage: "delete tag",
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
						title := c.String("title")
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
						processedTags := processTags(title)

						DeleteTagConfig := sncli.DeleteTagConfig{
							Session:   session,
							TagTitles: processedTags,

							Debug: c.GlobalBool("debug"),
						}
						return DeleteTagConfig.Run()
					},
				},
				{
					Name:  "note",
					Usage: "delete note",
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
						title := c.String("title")
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
						processedNotes := processNotes(title)

						DeleteNoteConfig := sncli.DeleteNoteConfig{
							Session:    session,
							NoteTitles: processedNotes,
							Debug:      c.GlobalBool("debug"),
						}
						return DeleteNoteConfig.Run()
					},
				},
			},
		},
		{
			Name:  "tag",
			Usage: "tag items",

			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "find-text",
					Usage: "match text",
				},
				cli.StringFlag{
					Name:  "find-text-ignore-case",
					Usage: "match text - case insensitive",
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
				findText := c.String("find-text")
				findTag := c.String("find-tag")
				newTags := c.String("title")
				email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
				if errMsg != "" {
					fmt.Printf("\nerror: %s\n\n", errMsg)
					return cli.ShowSubcommandHelp(c)
				}
				var session gosn.Session
				session, err = sncli.CliSignIn(email, password, apiServer)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				if findTag == "" && findText == "" {
					fmt.Println("you must provide either text or tag to search for")
					return cli.ShowSubcommandHelp(c)
				}
				processedTags := processTags(newTags)

				appConfig := sncli.TagItemsConfig{
					Session:    session,
					FindText:   findText,
					FindTag:    findTag,
					NewTags:    processedTags,
					Replace:    c.Bool("replace"),
					IgnoreCase: c.Bool("ignore-case"),
					Debug:      c.GlobalBool("debug"),
				}
				return appConfig.Run()
			},
		},
		{
			Name:  "get",
			Usage: "get items",
			Subcommands: []cli.Command{
				{
					Name:  "tags",
					Usage: "get tags",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "title",
							Usage: "find by title",
						},
						cli.BoolFlag{
							Name:  "count",
							Usage: "display count matching query",
						},
						cli.StringFlag{
							Name:  "output",
							Value: "json",
							Usage: "output format",
						},
					},
					OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
						//fmt.Fprintf(c.App.Writer, "pants\n")
						return err
					},
					Action: func(c *cli.Context) error {
						uuid := c.String("uuid")
						title := c.String("title")
						count := c.Bool("count")
						tagFilter := gosn.Filter{
							Type: "Tag",
						}
						getTagsIF := gosn.ItemFilters{
							MatchAny: false,
							Filters:  []gosn.Filter{tagFilter},
						}

						if uuid != "" {
							titleFilter := gosn.Filter{
								Type:       "Tag",
								Key:        "uuid",
								Comparison: "==",
								Value:      uuid,
							}
							getTagsIF.Filters = append(getTagsIF.Filters, titleFilter)
						}
						if title != "" {
							titleFilter := gosn.Filter{
								Type:       "Tag",
								Key:        "Title",
								Comparison: "contains",
								Value:      title,
							}
							getTagsIF.Filters = append(getTagsIF.Filters, titleFilter)
						}

						newTags := c.String("title")
						if strings.TrimSpace(newTags) == "" {
							fmt.Print("\nerror: tag title not defined\n\n")
							return cli.ShowSubcommandHelp(c)
						}
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						processedTags := processTags(newTags)
						// TODO: validate output
						output := c.String("output")
						appGetTagConfig := sncli.GetTagConfig{
							Session:   session,
							TagTitles: processedTags,
							Output:    output,
							Debug:     c.GlobalBool("debug"),
						}
						var tags gosn.GetItemsOutput
						tags, err = appGetTagConfig.Run()
						if err != nil {
							return err
						}

						numResults := len(tags.Items)
						if numResults <= 0 {
							fmt.Println("no matches")
						} else if count {
							fmt.Printf("%d matches", numResults)
						} else {
							output = c.String("output")
							var bOutput []byte
							switch strings.ToLower(output) {
							case "json":
								bOutput, err = json.MarshalIndent(tags, "", "    ")
							case "yaml":
								bOutput, err = yaml.Marshal(tags)
							}
							if len(bOutput) > 0 {
								fmt.Println(string(bOutput))
							}
						}

						return err
					},
				},
				{
					Name:  "notes",
					Usage: "get notes",
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
							Usage: "display count matching query",
						},
						cli.StringFlag{
							Name:  "output",
							Value: "json",
							Usage: "output format",
						},
					},
					Action: func(c *cli.Context) error {
						uuid := c.String("uuid")
						title := c.String("title")
						text := c.String("text")
						count := c.Bool("count")
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
						processedTags := processTags(c.String("tag"))

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

						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						getNoteConfig := sncli.GetNoteConfig{
							Session: session,
							Filters: getNotesIF,
							Debug:   c.GlobalBool("debug"),
						}
						var notes gosn.GetItemsOutput
						notes, err = getNoteConfig.Run()
						if err != nil {
							return err
						}
						numResults := len(notes.Items)
						if numResults <= 0 {
							fmt.Println("no matches")
						} else if count {
							fmt.Printf("%d matches", numResults)
						} else {
							output := c.String("output")
							var bOutput []byte
							switch strings.ToLower(output) {
							case "json":
								bOutput, err = json.MarshalIndent(notes, "", "    ")
							case "yaml":
								bOutput, err = yaml.Marshal(notes)
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
			Name:  "register",
			Usage: "register a new user",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "email",
					Usage: "email address",
				},
			},
			Action: func(c *cli.Context) error {
				var apiServer string
				if viper.GetString("server") != "" {
					apiServer = viper.GetString("server")
				} else {
					apiServer = sncli.SNServerURL
				}
				if strings.TrimSpace(c.String("email")) == "" {
					_ = cli.ShowCommandHelp(c, "register")
					os.Exit(1)
				}
				var password string
				fmt.Print("password: ")
				var bytePassword []byte
				bytePassword, err = terminal.ReadPassword(syscall.Stdin)
				if err == nil {
					password = string(bytePassword)
				}
				if len(password) == 0 {
					fmt.Println("password cannot be empty")
					os.Exit(1)
				}
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				if strings.TrimSpace(password) == "" {
					fmt.Println("password not defined")
					os.Exit(1)
				}
				registerConfig := sncli.RegisterConfig{
					Email:     c.String("email"),
					Password:  password,
					APIServer: apiServer,
				}
				return registerConfig.Run()
			},
		},

		{
			Name:  "stats",
			Usage: "show statistics",
			Action: func(c *cli.Context) error {
				email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
				if errMsg != "" {
					fmt.Printf("\nerror: %s\n\n", errMsg)
					return cli.ShowSubcommandHelp(c)
				}
				var session gosn.Session
				session, err = sncli.CliSignIn(email, password, apiServer)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				statsConfig := sncli.StatsConfig{
					Session: session,
				}
				err = statsConfig.Run()

				return err

			},
		},
		{
			Name:  "wipe",
			Usage: "deletes all tags and notes",
			Action: func(c *cli.Context) error {
				email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
				if errMsg != "" {
					fmt.Printf("\nerror: %s\n\n", errMsg)
					return cli.ShowSubcommandHelp(c)
				}
				var session gosn.Session
				session, err = sncli.CliSignIn(email, password, apiServer)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				wipeConfig := sncli.WipeConfig{
					Session: session,
				}
				var numWiped int
				fmt.Printf("wipe all items for account %s? ", email)
				var input string
				_, err = fmt.Scanln(&input)
				if err == nil && sncli.StringInSlice(input, []string{"y", "yes"}, false) {
					numWiped, err = wipeConfig.Run()
					fmt.Printf("%d items deleted\n", numWiped)
				} else {
					return nil
				}
				return err
			},
		},
		{
			Name:  "fixup",
			Usage: "find and fix item issues",
			Action: func(c *cli.Context) error {
				email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
				if errMsg != "" {
					fmt.Printf("\nerror: %s\n\n", errMsg)
					return cli.ShowSubcommandHelp(c)
				}
				var session gosn.Session
				session, err = sncli.CliSignIn(email, password, apiServer)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fixupConfig := sncli.FixupConfig{
					Session: session,
				}
				return fixupConfig.Run()
			},
		},
		{
			Name:  "test-data",
			Usage: "create test data",
			Subcommands: []cli.Command{
				{
					Name:  "tags",
					Usage: "new tags to create",
					Flags: []cli.Flag{
						cli.Int64Flag{
							Name:  "number",
							Usage: "number of tags",
							Value: 0,
						},
					},
					Action: func(c *cli.Context) error {
						numTags := c.Int64("number")
						if numTags <= 0 {
							return cli.ShowSubcommandHelp(c)
						}
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						appTestDataCreateTagsConfig := sncli.TestDataCreateTagsConfig{
							Session: session,
							NumTags: numTags,
							Debug:   c.GlobalBool("debug"),
						}
						err = appTestDataCreateTagsConfig.Run()
						return err

					},
				},
				{
					Name:  "notes",
					Usage: "new notes to create",
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "number",
							Usage: "number of tags",
							Value: 0,
						},
						cli.IntFlag{
							Name:  "paras",
							Usage: "number of paragraphs per note (min: 1)",
							Value: 5,
						},
					},
					Action: func(c *cli.Context) error {
						numNotes := c.Int("number")
						if numNotes <= 0 {
							return cli.ShowSubcommandHelp(c)
						}
						numParas := c.Int("paras")
						if numParas <= 1 {
							return cli.ShowSubcommandHelp(c)
						}
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						appTestDataCreateNotesConfig := sncli.TestDataCreateNotesConfig{
							Session:  session,
							NumNotes: numNotes,
							NumParas: numParas,
							Debug:    c.GlobalBool("debug"),
						}
						err = appTestDataCreateNotesConfig.Run()
						return err

					},
				},
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	return app.Run(args)
}
