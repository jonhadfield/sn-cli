package main

import (
	"fmt"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
)

func processItemKeysHealthcheck(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	fmt.Printf("opts: %+v\n", opts)
	msg = sncli.Green(fmt.Sprintf("%s", "TEST"))

	return msg, err
}
