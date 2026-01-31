package main

import (
	"fmt"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/urfave/cli/v2"
)

func cmdWipe() *cli.Command {
	return &cli.Command{
		Name:  "wipe",
		Usage: "deletes all supported content",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "yes",
				Usage: "accpet warning",
			},
			&cli.BoolFlag{
				Name:  "everything",
				Usage: "wipe settings also",
			},
		},
		BashComplete: func(c *cli.Context) {
			if c.NArg() > 0 {
				return
			}

			for _, t := range []string{"--yes", "--everything"} {
				fmt.Println(t)
			}
		},
		Action: func(c *cli.Context) error {
			opts := getOpts(c)

			cacheSession, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				return err
			}

			var cacheDBPath string

			cacheDBPath, err = cache.GenCacheDBPath(cacheSession, opts.cacheDBDir, snAppName)
			if err != nil {
				return err
			}

			cacheSession.CacheDBPath = cacheDBPath
			wipeConfig := sncli.WipeConfig{
				Session:    &cacheSession,
				UseStdOut:  opts.useStdOut,
				Everything: c.Bool("everything"),
				Debug:      opts.debug,
			}

			var proceed bool
			if c.Bool("yes") {
				proceed = true
			} else {
				fmt.Printf("wipe all items for account %s? ", cacheSession.Session.KeyParams.Identifier)
				var input string
				_, err = fmt.Scanln(&input)
				if err == nil && sncli.StringInSlice(input, []string{"y", "yes"}, false) {
					proceed = true
				}
			}
			if proceed {
				var numWiped int
				var wipeErr error
				numWiped, wipeErr = wipeConfig.Run()
				if wipeErr != nil {
					return wipeErr
				}

				_, _ = fmt.Fprintf(c.App.Writer, "%d %s", numWiped, msgItemsDeleted)

				return nil
			}

			return nil
		},
	}
}
