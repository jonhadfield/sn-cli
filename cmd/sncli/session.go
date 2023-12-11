package main

import (
	"os"

	"github.com/jonhadfield/gosn-v2/session"
	"github.com/urfave/cli"
)

func processSession(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	sAdd := c.Bool("add")
	sRemove := c.Bool("remove")
	sStatus := c.Bool("status")
	sessKey := c.String("session-key")

	if sStatus || sRemove {
		if err = session.SessionExists(nil); err != nil {
			return "", err
		}
	}

	nTrue := numTrue(sAdd, sRemove, sStatus)
	if nTrue == 0 || nTrue > 1 {
		_ = cli.ShowCommandHelp(c, "session")

		os.Exit(1)
	}

	if sAdd {
		msg, err = session.AddSession(opts.server, sessKey, nil, opts.debug)

		return msg, err
	}

	if sRemove {
		msg = session.RemoveSession(nil)

		return msg, nil
	}

	if sStatus {
		msg, err = session.SessionStatus(sessKey, nil)
	}

	return msg, err
}
