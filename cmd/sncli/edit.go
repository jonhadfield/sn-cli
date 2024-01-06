package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func cmdEdit() *cli.Command {
	return &cli.Command{
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
		Subcommands: []*cli.Command{
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
					&cli.StringFlag{
						Name:  "title",
						Usage: "title of the tag",
					},
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "uuid of the tag",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					return processEditTag(c, opts)
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
					&cli.StringFlag{
						Name:  "title",
						Usage: "title of the note",
					},
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "uuid of the note",
					},
					&cli.StringFlag{
						Name:    "editor",
						Usage:   "path to editor",
						EnvVars: []string{"EDITOR"},
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)
					// useStdOut = opts.useStdOut
					return processEditNote(c, opts)
				},
			},
		},
	}
}
