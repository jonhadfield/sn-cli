package main

import (
	"fmt"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
)

func processItemKeysHealthcheck(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	msg = sncli.Green(fmt.Sprintf("%s", "TEST"))

	return msg, err
}
