package sncli

import (
	"github.com/jonhadfield/gosn"
	"log"
)

type ExportConfig struct {
	Session gosn.Session
	Output  string
	Debug   bool
}

func (input *ExportConfig) Run() error {
	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}
	gii := gosn.GetItemsInput{
		Session: input.Session,
	}
	gio, err := gosn.GetItems(gii)
	err = writeGob(input.Output, gio.Items)
	return err
}
