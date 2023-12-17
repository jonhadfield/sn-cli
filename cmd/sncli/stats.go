package main

import (
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
)

func cmdStats() *cli.Command {
	return &cli.Command{
		Name:  "stats",
		Usage: "show statistics",
		Action: func(c *cli.Context) error {
			var opts configOptsOutput
			opts, err := getOpts(c)
			if err != nil {
				return err
			}
			var sess cache.Session
			sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
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
