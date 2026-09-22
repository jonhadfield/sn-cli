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
			addTasks := []string{"tag", "note", "duplicates"}
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
						Name:  flagTitleName,
						Usage: "title of note to delete (separate multiple with commas)",
					},
					&cli.StringFlag{
						Name:  flagUUIDName,
						Usage: "unique id of note to delete (separate multiple with commas)",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					return processDeleteTags(c, opts)
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
						Name:  flagTitleName,
						Usage: "title of note to delete (separate multiple with commas)",
					},
					&cli.StringFlag{
						Name:  flagUUIDName,
						Usage: "unique id of note to delete (separate multiple with commas)",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					return processDeleteNote(c, opts)
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
						Name:  flagUUIDName,
						Usage: "unique id of item to delete (separate multiple with commas)",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					return processDeleteItems(c, opts)
				},
			},
			{
				Name:  "duplicates",
				Usage: "delete notes that Standard Notes marked as copies of another note",
				BashComplete: func(c *cli.Context) {
					delDupeOpts := []string{"--dry-run", "--identical-only", "--yes"}
					if c.NArg() > 0 {
						return
					}
					for _, t := range delDupeOpts {
						fmt.Println(t)
					}
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "list the duplicates without deleting them",
					},
					&cli.BoolFlag{
						Name:  "identical-only",
						Usage: "only delete notes whose title and text match the note being kept",
					},
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "delete without asking for confirmation",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)

					return processDeleteDuplicates(c, opts)
				},
			},
		},
	}
}
