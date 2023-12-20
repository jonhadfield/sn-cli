package main

import (
	"fmt"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
)

func cmdChecklist() *cli.Command {
	return &cli.Command{
		Name:  "checklist",
		Usage: "manage checklists",
		BashComplete: func(c *cli.Context) {
			addTasks := []string{"list"}
			if c.NArg() > 0 {
				return
			}
			for _, t := range addTasks {
				fmt.Println(t)
			}
		},
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list checklists",
				Action: func(c *cli.Context) error {
					var opts configOptsOutput
					opts, err := getOpts(c)
					if err != nil {
						return err
					}
					var sess cache.Session
					sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
					if !sess.SchemaValidation {
						panic("schema validation is false")
					}
					if err != nil {
						return err
					}
					sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
					if err != nil {
						return err
					}
					listChecklistsConfig := sncli.ListChecklistsInput{
						Session: &sess,
					}

					return listChecklistsConfig.Run()
				},
			},
			// {
			// 	Name:  "show",
			// 	Usage: "show checklist",
			// 	BashComplete: func(c *cli.Context) {
			// 		if c.NArg() > 0 {
			// 			return
			// 		}
			// 		fmt.Println("--title")
			// 	},
			// 	Flags: []cli.Flag{
			// 		&cli.StringFlag{
			// 			Name:  "title",
			// 			Usage: "new tag title (separate multiple with commas)",
			// 		},
			// 	},
			// 	Action: func(c *cli.Context) error {
			// 		var opts configOptsOutput
			// 		opts, err := getOpts(c)
			// 		if err != nil {
			// 			return err
			// 		}
			// 		// useStdOut = opts.useStdOut
			// 		err = processAddTags(c, opts)
			//
			// 		return err
			// 	},
			// },
		},
	}
}
