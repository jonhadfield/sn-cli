package main

import (
	"github.com/jonhadfield/gosn-v2"
	"github.com/urfave/cli"
	"os"
)

func processSession(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	sAdd := c.Bool("add")
	sRemove := c.Bool("remove")
	sStatus := c.Bool("status")
	sessKey := c.String("session-key")
	if sStatus || sRemove {
		if err = gosn.SessionExists(nil); err != nil {
			return "", err
		}
	}
	nTrue := numTrue(sAdd, sRemove, sStatus)
	if nTrue == 0 || nTrue > 1 {
		_ = cli.ShowCommandHelp(c, "session")
		os.Exit(1)
	}
	if sAdd {
		msg, err = gosn.AddSession(opts.server, sessKey, nil)
		return "", err
	}
	if sRemove {
		msg = gosn.RemoveSession(nil)
		return "", nil
	}
	if sStatus {
		msg, err = gosn.SessionStatus(sessKey, nil)
	}

	return msg, err
}
