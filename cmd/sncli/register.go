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

			for _, t := range []string{"--email"} {
				fmt.Println(t)
			}
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "email",
				Usage: "email address",
			},
		},
		Action: func(c *cli.Context) error {
			opts := getOpts(c)

			var err error
			if strings.TrimSpace(c.String("email")) == "" {
				if err = cli.ShowCommandHelp(c, "register"); err != nil {
					panic(err)
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
			if err = registerConfig.Run(); err != nil {
				return err
			}

			fmt.Println(msgRegisterSuccess)

			return nil
		},
	}
}
