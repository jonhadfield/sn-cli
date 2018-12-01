package main

import (
	"encoding/json"
	"errors"
	"gopkg.in/urfave/cli.v1"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/yaml.v2"

	"github.com/jonhadfield/gosn"

	"github.com/jonhadfield/sncli"

	"fmt"

	"github.com/spf13/viper"
)

const (
	msgAddSuccess      = "added."
	msgDeleteSuccess   = "deleted."
	msgCreateSuccess   = "created."
	msgRegisterSuccess = "registered."
)

var yamlAbbrevs = []string{"yml", "yaml"}

// overwritten at build time
var version, versionOutput, tag, sha, buildDate string

func main() {
	msg, err := startCLI(os.Args)
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		os.Exit(1)
	}
	fmt.Println(msg)
	os.Exit(0)
}

func commaSplit(input string) []string {
	o := strings.Split(input, ",")
	if len(o) == 1 && len(o[0]) == 0 {
		return nil
	}
	return o
}

func startCLI(args []string) (msg string, err error) {
	viper.SetEnvPrefix("sn")
	err = viper.BindEnv("email")
	if err != nil {
		return "", err
	}
	err = viper.BindEnv("password")
	if err != nil {
		return "", err
	}
	err = viper.BindEnv("server")
	if err != nil {
		return "", err
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
						tagInput := c.String("title")
						if strings.TrimSpace(tagInput) == "" {
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New("tag title not defined")
						}
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							return fmt.Errorf("failed to authenticate. %+v", err)
						}
						if errMsg != "" {
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New(errMsg)
						}

						tags := commaSplit(tagInput)
						appAddTagConfig := sncli.AddTagConfig{
							Session: session,
							Tags:    tags,
							Debug:   c.GlobalBool("debug"),
						}
						if err = appAddTagConfig.Run(); err != nil {
							return fmt.Errorf("failed to add tag: %+v", err)
						}
						msg = msgAddSuccess
						return nil
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
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New("note title not defined")
						}
						if strings.TrimSpace(text) == "" {
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New("note text not defined")

						}

						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New(errMsg)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							return err
						}

						processedTags := commaSplit(c.String("tag"))

						AddNoteConfig := sncli.AddNoteConfig{
							Session: session,
							Title:   title,
							Text:    text,
							Tags:    processedTags,
							Replace: false,
							Debug:   c.GlobalBool("debug"),
						}
						if err = AddNoteConfig.Run(); err != nil {
							return fmt.Errorf("failed to add note. %+v", err)
						}
						msg = msgAddSuccess
						return nil
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
						titleIn := strings.TrimSpace(c.String("title"))
						uuidIn := strings.Replace(c.String("uuid"), " ", "", -1)
						if titleIn == "" && uuidIn == "" {
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New("title or uuid required")
						}
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New(errMsg)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							return err
						}
						tags := commaSplit(titleIn)
						uuids := commaSplit(uuidIn)

						DeleteTagConfig := sncli.DeleteTagConfig{
							Session:   session,
							TagTitles: tags,
							TagUUIDs:  uuids,
							Debug:     c.GlobalBool("debug"),
						}
						if err = DeleteTagConfig.Run(); err != nil {
							return fmt.Errorf("failed to delete tag. %+v", err)
						}
						msg = msgDeleteSuccess
						return nil
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
						title := strings.TrimSpace(c.String("title"))
						uuid := strings.TrimSpace(c.String("uuid"))
						if title == "" && uuid == "" {
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New("")
						}
						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							return err
						}
						processedNotes := commaSplit(title)

						DeleteNoteConfig := sncli.DeleteNoteConfig{
							Session:    session,
							NoteTitles: processedNotes,
							Debug:      c.GlobalBool("debug"),
						}
						if err = DeleteNoteConfig.Run(); err != nil {
							return fmt.Errorf("failed to delete note. %+v", err)
						}
						msg = msgDeleteSuccess
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
				findTitle := c.String("find-title")
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
					return err
				}
				if findText == "" && findTitle == "" && findTag == "" {
					fmt.Println("you must provide either text, title, or tag to search for")
					return cli.ShowSubcommandHelp(c)
				}
				processedTags := commaSplit(newTags)

				appConfig := sncli.TagItemsConfig{
					Session:    session,
					FindText:   findText,
					FindTitle:  findTitle,
					FindTag:    findTag,
					NewTags:    processedTags,
					Replace:    c.Bool("replace"),
					IgnoreCase: c.Bool("ignore-case"),
					Debug:      c.GlobalBool("debug"),
				}
				err = appConfig.Run()
				if err != nil {
					return err
				}
				msg = msgDeleteSuccess
				return nil
			},
		},
		{
			Name:  "get",
			Usage: "get items",
			//Description: "get all items or limit results by search criteria",
			Subcommands: []cli.Command{
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
							Usage: "display count only",
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
							for _, uuid := range commaSplit(inUUID) {
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
							for _, title := range commaSplit(inTitle) {
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

						email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
						if errMsg != "" {
							fmt.Printf("\nerror: %s\n\n", errMsg)
							return cli.ShowSubcommandHelp(c)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							return err
						}

						//processedTags := commaSplit(newTags)
						// TODO: validate output
						output := c.String("output")
						appGetTagConfig := sncli.GetTagConfig{
							Session: session,
							Filters: getTagsIF,
							Output:  output,
							Debug:   c.GlobalBool("debug"),
						}
						var rawTags gosn.GetItemsOutput
						rawTags, err = appGetTagConfig.Run()
						if err != nil {
							return err
						}

						var tagsYAML []sncli.TagYAML
						var tagsJSON []sncli.TagJSON
						var numResults int
						for _, rt := range rawTags.Items {
							numResults++
							if !count && sncli.StringInSlice(output, yamlAbbrevs, false) {
								tagContentOrgStandardNotesSNDetailYAML := sncli.OrgStandardNotesSNDetailYAML{
									ClientUpdatedAt: rt.Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								tagContentAppDataContent := sncli.AppDataContentYAML{
									OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailYAML,
								}
								tagContentYAML := sncli.TagContentYAML{
									Title:   rt.Content.GetTitle(),
									AppData: tagContentAppDataContent,
								}

								tagsYAML = append(tagsYAML, sncli.TagYAML{
									UUID:        rt.UUID,
									ContentType: rt.ContentType,
									Content:     tagContentYAML,
									UpdatedAt:   rt.UpdatedAt,
									CreatedAt:   rt.CreatedAt,
								})
							}
							if !count && strings.ToLower(output) == "json" {
								tagContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
									ClientUpdatedAt: rt.Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								tagContentAppDataContent := sncli.AppDataContentJSON{
									OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailJSON,
								}
								tagContentJSON := sncli.TagContentJSON{
									Title:   rt.Content.GetTitle(),
									AppData: tagContentAppDataContent,
								}

								tagsJSON = append(tagsJSON, sncli.TagJSON{
									UUID:        rt.UUID,
									ContentType: rt.ContentType,
									Content:     tagContentJSON,
									UpdatedAt:   rt.UpdatedAt,
									CreatedAt:   rt.CreatedAt,
								})
							}
						}
						if numResults <= 0 {
							fmt.Println("no matches.")
						} else if count {
							fmt.Println(numResults)
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
							Usage: "display countonly",
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
						processedTags := commaSplit(c.String("tag"))

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
							return err
						}

						getNoteConfig := sncli.GetNoteConfig{
							Session: session,
							Filters: getNotesIF,
							Debug:   c.GlobalBool("debug"),
						}
						var rawNotes gosn.GetItemsOutput
						rawNotes, err = getNoteConfig.Run()
						if err != nil {
							return err
						}
						numResults := len(rawNotes.Items)
						var notesYAML []sncli.NoteYAML
						var notesJSON []sncli.NoteJSON

						for _, rt := range rawNotes.Items {
							numResults++
							if !count && sncli.StringInSlice(output, yamlAbbrevs, false) {
								noteContentOrgStandardNotesSNDetailYAML := sncli.OrgStandardNotesSNDetailYAML{
									ClientUpdatedAt: rt.Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								noteContentAppDataContent := sncli.AppDataContentYAML{
									OrgStandardNotesSN: noteContentOrgStandardNotesSNDetailYAML,
								}
								noteContentYAML := sncli.NoteContentYAML{
									Title:   rt.Content.GetTitle(),
									AppData: noteContentAppDataContent,
								}

								notesYAML = append(notesYAML, sncli.NoteYAML{
									UUID:        rt.UUID,
									ContentType: rt.ContentType,
									Content:     noteContentYAML,
									UpdatedAt:   rt.UpdatedAt,
									CreatedAt:   rt.CreatedAt,
								})
							}
							if !count && strings.ToLower(output) == "json" {
								noteContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
									ClientUpdatedAt: rt.Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
								}
								noteContentAppDataContent := sncli.AppDataContentJSON{
									OrgStandardNotesSN: noteContentOrgStandardNotesSNDetailJSON,
								}
								noteContentJSON := sncli.NoteContentJSON{
									Title:   rt.Content.GetTitle(),
									Text:   rt.Content.GetText(),
									AppData: noteContentAppDataContent,
								}

								notesJSON = append(notesJSON, sncli.NoteJSON{
									UUID:        rt.UUID,
									ContentType: rt.ContentType,
									Content:     noteContentJSON,
									UpdatedAt:   rt.UpdatedAt,
									CreatedAt:   rt.CreatedAt,
								})
							}
						}



						if numResults <= 0 {
							fmt.Println("no matches.")
						} else if count {
							fmt.Println(numResults)
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
					if cErr := cli.ShowCommandHelp(c, "register"); err != nil {
						panic(cErr)
					}
					return errors.New("email required")
				}
				var password string
				fmt.Print("password: ")
				var bytePassword []byte
				bytePassword, err = terminal.ReadPassword(0)
				if err == nil {
					password = string(bytePassword)
				}
				if len(password) == 0 {
					return errors.New("password cannot be empty")
				}
				if err != nil {
					return err
				}
				if strings.TrimSpace(password) == "" {
					return errors.New("password not defined")
				}
				registerConfig := sncli.RegisterConfig{
					Email:     c.String("email"),
					Password:  password,
					APIServer: apiServer,
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
				email, password, apiServer, errMsg := sncli.GetCredentials(c.GlobalString("server"))
				if errMsg != "" {
					fmt.Printf("\nerror: %s\n\n", errMsg)
					return cli.ShowSubcommandHelp(c)
				}
				var session gosn.Session
				session, err = sncli.CliSignIn(email, password, apiServer)
				if err != nil {
					return err
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
					return err
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
					fmt.Printf("%d items deleted\n\n", numWiped)
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
					return err
				}
				fixupConfig := sncli.FixupConfig{
					Session: session,
				}
				err = fixupConfig.Run()
				if err != nil {
					return fmt.Errorf("fixup failed. %+v", err)
				}
				return nil
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
							return err
						}

						appTestDataCreateTagsConfig := sncli.TestDataCreateTagsConfig{
							Session: session,
							NumTags: numTags,
							Debug:   c.GlobalBool("debug"),
						}
						if appTestDataCreateTagsConfig.Run() != nil {
							return err
						}
						fmt.Println(msgCreateSuccess)
						return nil

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
							if cErr := cli.ShowSubcommandHelp(c); err != nil {
								panic(cErr)
							}
							return errors.New(errMsg)
						}
						var session gosn.Session
						session, err = sncli.CliSignIn(email, password, apiServer)
						if err != nil {
							return err
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
	return msg, app.Run(args)
}
