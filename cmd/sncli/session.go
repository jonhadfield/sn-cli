package main

import (
	"fmt"
	"os"

	"github.com/jonhadfield/gosn-v2/session"
	"github.com/urfave/cli/v2"
)

func cmdSession() *cli.Command {
	return &cli.Command{
		Name:  "session",
		Usage: "manage session credentials",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "add",
				Usage: "add session to keychain",
			},
			&cli.BoolFlag{
				Name:  "remove",
				Usage: "remove session from keychain",
			},
			&cli.BoolFlag{
				Name:  "status",
				Usage: "get session details",
			},
			&cli.StringFlag{
				Name:     "session-key",
				Usage:    "[optional] key to encrypt/decrypt session",
				Required: false,
			},
		},
		Hidden: false,
		Action: func(c *cli.Context) error {
			var opts configOptsOutput
			opts, err := getOpts(c)
			if err != nil {
				return err
			}
			// useStdOut = opts.useStdOut

			return processSession(c, opts)
		},
	}
}

func processSession(c *cli.Context, opts configOptsOutput) (err error) {
	sAdd := c.Bool("add")
	sRemove := c.Bool("remove")
	sStatus := c.Bool("status")
	sessKey := c.String("session-key")

	if sStatus || sRemove {
		if err = session.SessionExists(nil); err != nil {
			return err
		}
	}

	nTrue := numTrue(sAdd, sRemove, sStatus)
	if nTrue == 0 || nTrue > 1 {
		_ = cli.ShowCommandHelp(c, "session")

		os.Exit(1)
	}

	if sAdd {
		var msg string

		msg, err = session.AddSession(opts.server, sessKey, nil, opts.debug)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprint(c.App.Writer, msg)

		return nil
	}

	if sRemove {
		msg := session.RemoveSession(nil)
		_, _ = fmt.Fprint(c.App.Writer, msg)

		return nil
	}

	if sStatus {
		var msg string
		msg, err = session.SessionStatus(sessKey, nil)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprint(c.App.Writer, msg)
	}

	return err
}

// {
//			Name:  "output-session",
//			Usage: "returns specified session items",
//			BashComplete: func(c *cli.Context) {
//				hcKeysOpts := []string{"--master-key"}
//				if c.NArg() > 0 {
//					return
//				}
//				for _, ano := range hcKeysOpts {
//					fmt.Println(ano)
//				}
//			},
//			Flags: []cli.Flag{
//				&cli.BoolFlag{
//					Name:  "master-key",
//					Usage: "output master key",
//				},
//			},
//			Action: func(c *cli.Context) error {
//				var opts configOptsOutput
//				opts, err := getOpts(c)
//				if err != nil {
//					return err
//				}
//				// useStdOut = opts.useStdOut
//
//				var sess session.Session
//
//				sess, _, err = session.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
//
//				if err != nil {
//					return err
//				}
//				err = sncli.OutputSession(sncli.OutputSessionInput{
//					Session:         sess,
//					UseStdOut:       opts.useStdOut,
//					OutputMasterKey: c.Bool("master-key"),
//				})
//
//				return err
//			},
//		},
