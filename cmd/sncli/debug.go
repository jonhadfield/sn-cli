package main

import (
	"fmt"

	"github.com/jonhadfield/gosn-v2/session"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
)

func cmdDebug() *cli.Command {
	return &cli.Command{
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
		Subcommands: []*cli.Command{
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
					&cli.StringFlag{
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
					opts, err := getOpts(c)
					if err != nil {
						return err
					}
					// useStdOut = opts.useStdOut

					var sess session.Session

					sess, _, err = session.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
					if err != nil {
						return err
					}
					// var res string
					_, err = sncli.DecryptString(sncli.DecryptStringInput{
						Session:   sess,
						UseStdOut: opts.useStdOut,
						Key:       c.String("key"),
						In:        str,
					})
					if err != nil {
						return err
					}

					// msg = fmt.Sprintf("plaintext: %s", res)

					return err
				},
			},
		},
	}
}

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
// 		// useStdOut = opts.useStdOut
//
// 		return sncli.CreateItemsKey(sncli.CreateItemsKeyInput{
// 			Debug:     opts.debug,
// 			MasterKey: c.String("master-key"),
// 		})
// 	},
// },
