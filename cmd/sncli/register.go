package main

import (
	"errors"
	"fmt"
	"strings"

	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
)

func cmdRegister() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "register a new user",
		BashComplete: func(c *cli.Context) {
			if c.NArg() > 0 {
				return
			}
			fmt.Println("--email")
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "email",
				Usage: "email address",
			},
		},
		Action: func(c *cli.Context) error {
			var opts configOptsOutput
			opts, err := getOpts(c)
			if err != nil {
				return err
			}

			// useStdOut = opts.useStdOut

			if strings.TrimSpace(c.String("email")) == "" {
				if cErr := cli.ShowCommandHelp(c, "register"); cErr != nil {
					panic(cErr)
				}

				return errors.New("email required")
			}

			var password string
			if password, err = getPassword(); err != nil {
				return err
			}

			registerConfig := sncli.RegisterConfig{
				Email:     c.String("email"),
				Password:  password,
				APIServer: opts.server,
				Debug:     opts.debug,
			}
			err = registerConfig.Run()
			if err != nil {
				return err
			}
			fmt.Println(msgRegisterSuccess)

			return nil
		},
	}
}
