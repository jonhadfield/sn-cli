package main

import (
	"fmt"

	"github.com/jonhadfield/gosn-v2/session"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
)

func cmdHealthcheck() *cli.Command {
	return &cli.Command{
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
		Subcommands: []*cli.Command{
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
					&cli.BoolFlag{
						Hidden: true,
						Name:   "delete-invalid",
						Usage:  "delete items that cannot be decrypted",
					},
				},
				Action: func(c *cli.Context) error {
					opts := getOpts(c)
					// useStdOut = opts.useStdOut

					var sess session.Session

					sess, _, err := session.GetSession(nil, opts.useSession, opts.sessKey, opts.server, opts.debug)
					if err != nil {
						return err
					}
					err = sncli.ItemKeysHealthcheck(sncli.ItemsKeysHealthcheckInput{
						Session:       sess,
						UseStdOut:     opts.useStdOut,
						DeleteInvalid: c.Bool("delete-invalid"),
					})

					return err
				},
			},
		},
	}
}
