package main

import (
	"fmt"
	"strings"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func cmdEditor() *cli.Command {
	return &cli.Command{
		Name:  "editor",
		Usage: "manage the editor used for new notes",
		Description: "Standard Notes records the editor used for new notes as an account preference.\n" +
			"   Note that a tag can override it: a note created while a tag with its own editor\n" +
			"   preference is selected uses the tag's editor instead of the account default.",
		Subcommands: []*cli.Command{
			cmdEditorList(),
			cmdEditorSetDefault(),
		},
	}
}

// editorSession assembles the session and cache path shared by the editor
// subcommands.
func editorSession(c *cli.Context) (sess cache.Session, err error) {
	opts := getOpts(c)

	sess, _, err = cache.GetSession(common.NewHTTPClient(), viper.GetBool("use_session"), opts.sessKey, viper.GetString("server"), opts.debug)
	if err != nil {
		return sess, err
	}

	sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)

	return sess, err
}

func cmdEditorList() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "list the editors available to this account",
		UsageText: "sn editor list",
		Action: func(c *cli.Context) error {
			sess, err := editorSession(c)
			if err != nil {
				return err
			}

			config := sncli.ListEditorsConfig{Session: &sess, Debug: getOpts(c).debug}

			editors, err := config.Run()
			if err != nil {
				return err
			}

			if len(editors) == 0 {
				return fmt.Errorf("no editors found")
			}

			// Width of the widest name, so identifiers line up.
			width := 0
			for _, e := range editors {
				if len(e.Name) > width {
					width = len(e.Name)
				}
			}

			for _, e := range editors {
				marker := "  "
				if e.Current {
					marker = "* "
				}

				var notes []string
				if !e.BuiltIn {
					notes = append(notes, "installed")
				}

				if e.Deprecated {
					notes = append(notes, "deprecated")
				}

				line := fmt.Sprintf("%s%-*s  %s", marker, width, e.Name, e.Identifier)
				if len(notes) > 0 {
					line += fmt.Sprintf("  (%s)", strings.Join(notes, ", "))
				}

				fmt.Println(line)
			}

			fmt.Println("\n* current default for new notes")

			return nil
		},
	}
}

func cmdEditorSetDefault() *cli.Command {
	return &cli.Command{
		Name:      "set-default",
		Usage:     "set the editor used for new notes",
		UsageText: "sn editor set-default <name or identifier>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Usage: "set the identifier without checking it against the available editors",
			},
		},
		Action: func(c *cli.Context) error {
			editor := strings.TrimSpace(strings.Join(c.Args().Slice(), " "))
			if editor == "" {
				_ = cli.ShowSubcommandHelp(c)

				return fmt.Errorf("an editor name or identifier is required")
			}

			sess, err := editorSession(c)
			if err != nil {
				return err
			}

			config := sncli.SetDefaultEditorConfig{
				Session: &sess,
				Editor:  editor,
				Force:   c.Bool("force"),
				Debug:   getOpts(c).debug,
			}

			set, err := config.Run()
			if err != nil {
				return err
			}

			name := set.Name
			if name == "" {
				name = set.Identifier
			}

			fmt.Printf("default editor for new notes set to %s (%s)\n", name, set.Identifier)

			if set.Deprecated {
				fmt.Println("note: Standard Notes has deprecated this editor")
			}

			return nil
		},
	}
}
