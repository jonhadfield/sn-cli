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
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "visual",
				Aliases: []string{"v"},
				Usage:   "display stats with visual charts and graphs",
			},
		},
		BashComplete: func(c *cli.Context) {
			if c.NArg() > 0 {
				return
			}
		},
		Action: func(c *cli.Context) error {
			opts := getOpts(c)

			// Show progress spinner while loading
			spinner, _ := ShowProgress("📊 Loading statistics...")

			sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
			if err != nil {
				if spinner != nil {
					spinner.Fail("Failed to get session")
				}
				return err
			}

			sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
			if err != nil {
				if spinner != nil {
					spinner.Fail("Failed to generate cache path")
				}
				return err
			}

			statsConfig := sncli.StatsConfig{
				Session: sess,
			}

			// Get the data
			data, err := statsConfig.GetData()
			if spinner != nil {
				if err != nil {
					spinner.Fail("Failed to load statistics")
					return err
				}
				spinner.Success("Statistics loaded")
			}

			// Display stats
			if c.Bool("visual") {
				return ShowVisualStats(data)
			}

			// Use original display
			return statsConfig.Run()
		},
	}
}
