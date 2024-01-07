package main

import (
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
)

func cmdStats() *cli.Command {
	return &cli.Command{
		Name:  "stats",
		Usage: "show statistics",
		BashComplete: func(c *cli.Context) {
			if c.NArg() > 0 {
				return
			}
		},
		Action: func(c *cli.Context) error {
			opts := getOpts(c)

			sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				return err
			}

			sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
			if err != nil {
				return err
			}

			statsConfig := sncli.StatsConfig{
				Session: sess,
			}

			return statsConfig.Run()
		},
	}
}
