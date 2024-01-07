package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func cmdAdd() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "add items",
		BashComplete: func(c *cli.Context) {
			if c.NArg() > 0 {
				return
			}
			for _, t := range []string{"tag", "note"} {
				fmt.Println(t)
			}
		},
		Subcommands: []*cli.Command{
			{
				Name:  "tag",
				Usage: "add tags",
				BashComplete: func(c *cli.Context) {
					if c.NArg() > 0 {
						return
					}
					for _, t := range []string{"--title", "--parent", "--parent-uuid"} {
						fmt.Println(t)
					}
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "title",
						Usage: "new tag title (separate multiple with commas)",
					},
					&cli.StringFlag{
						Name:  "parent",
						Usage: "parent tag title to make a sub-tag of",
					},
					&cli.StringFlag{
						Name:  "parent-uuid",
						Usage: "parent tag uuid to make a sub-tag of",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)
					// useStdOut = opts.useStdOut
					return processAddTags(c, opts)
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
					&cli.StringFlag{
						Name:  "title",
						Usage: "new note title",
					},
					&cli.StringFlag{
						Name:  "text",
						Usage: "new note text",
					},
					&cli.StringFlag{
						Name:  "file",
						Usage: "path to file with note content (specify --title or leave blank to use filename)",
					},
					&cli.StringFlag{
						Name:  "tag",
						Usage: "associate with tag (separate multiple with commas)",
					},
					&cli.BoolFlag{
						Name:  "replace",
						Usage: "replace note with same title",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					return processAddNotes(c, opts)
				},
			},
		},
	}
}
