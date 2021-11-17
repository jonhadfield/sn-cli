package main

import (
	"os"

	gosn "github.com/jonhadfield/gosn-v2"
	"github.com/urfave/cli"
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
		msg, err = gosn.AddSession(opts.server, sessKey, nil, opts.debug)
		return msg, err
	}

	if sRemove {
		msg = gosn.RemoveSession(nil)
		return msg, nil
	}

	if sStatus {
		msg, err = gosn.SessionStatus(sessKey, nil)
	}

	return msg, err
}
