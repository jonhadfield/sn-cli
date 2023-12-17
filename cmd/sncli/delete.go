package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func cmdDelete() *cli.Command {
	return &cli.Command{
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
		Subcommands: []*cli.Command{
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
					&cli.StringFlag{
						Name:  "title",
						Usage: "title of note to delete (separate multiple with commas)",
					},
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "unique id of note to delete (separate multiple with commas)",
					},
				},
				Action: func(c *cli.Context) error {
					opts, err := getOpts(c)
					if err != nil {
						return err
					}

					// useStdOut = opts.useStdOut

					err = processDeleteTags(c, opts)

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
					&cli.StringFlag{
						Name:  "title",
						Usage: "title of note to delete (separate multiple with commas)",
					},
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "unique id of note to delete (separate multiple with commas)",
					},
				},
				Action: func(c *cli.Context) error {
					opts, err := getOpts(c)
					if err != nil {
						return err
					}

					err = processDeleteNote(c, opts)

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
					&cli.StringFlag{
						Name:  "uuid",
						Usage: "unique id of item to delete (separate multiple with commas)",
					},
				},
				Action: func(c *cli.Context) error {
					opts, err := getOpts(c)
					if err != nil {
						return err
					}

					// useStdOut = opts.useStdOut

					err = processDeleteItems(c, opts)

					return err
				},
			},
		},
	}
}
