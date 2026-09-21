package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonhadfield/gosn-v2/session"
	"github.com/urfave/cli/v2"
	"github.com/zalando/go-keyring"
)

func cmdSession() *cli.Command {
	return &cli.Command{
		Name:  "session",
		Usage: "manage session credentials",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "add",
				Usage: "add session to keychain, or to the file given by --session-file",
			},
			&cli.BoolFlag{
				Name:  "remove",
				Usage: "remove session from keychain, or from the file given by --session-file",
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
		BashComplete: func(c *cli.Context) {
			if c.NArg() > 0 {
				return
			}

			for _, t := range []string{"--add", "--remove", "--status", "--session-key"} {
				fmt.Println(t)
			}
		},
		Action: func(c *cli.Context) error {
			opts := getOpts(c)

			return processSession(c, opts)
		},
	}
}

// useSessionFile makes the session commands, and every command run with
// --use-session, store the session in path instead of the system keyring. An
// empty path leaves the system keyring in use.
func useSessionFile(path string) error {
	if path == "" {
		session.SetDefaultKeyring(nil)

		return nil
	}

	// expand ~ ourselves, as the shell does not when the path comes from
	// SN_SESSION_FILE or the config file
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand session file path: %w", err)
		}

		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}

	session.SetDefaultKeyring(session.NewFileKeyring(path))

	return nil
}

func processSession(c *cli.Context, opts configOptsOutput) (err error) {
	sAdd := c.Bool("add")
	sRemove := c.Bool("remove")
	sStatus := c.Bool("status")
	sessKey := c.String("session-key")

	if sStatus || sRemove {
		if err = session.SessionExists(nil); err != nil {
			if sessionFile := c.String("session-file"); sessionFile != "" && errors.Is(err, keyring.ErrNotFound) {
				return fmt.Errorf("no session found in %s, add one with: sn --session-file %s session --add", sessionFile, sessionFile)
			}

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

		msg, err = session.AddSession(nil, opts.server, sessKey, nil, opts.debug)
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

		_, _ = fmt.Fprint(c.App.Writer, msg+"\n")
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
//				opts := getOpts(c)
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
