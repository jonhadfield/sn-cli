package main

import (
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
)

func cmdResync() *cli.Command {
	return &cli.Command{
		Name:  "resync",
		Usage: "purge cache and resync content",
		Action: func(c *cli.Context) error {
			var opts configOptsOutput
			opts, err := getOpts(c)
			if err != nil {
				return err
			}
			var session cache.Session
			session, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				return err
			}

			return sncli.Resync(&session, opts.cacheDBDir, snAppName)
		},
	}
}