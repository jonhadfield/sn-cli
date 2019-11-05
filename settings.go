package sncli

import (
	"log"

	"github.com/jonhadfield/gosn"
)

func (input *GetSettingsConfig) Run() (settings gosn.Items, err error) {
	gosn.SetErrorLogger(log.Println)

	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}

	getItemsInput := gosn.GetItemsInput{
		Session: input.Session,
	}

	var output gosn.GetItemsOutput

	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return nil, err
	}

	output.Items.DeDupe()

	settings, err = output.Items.DecryptAndParse(input.Session.Mk, input.Session.Ak)
	if err != nil {
		return nil, err
	}

	settings.Filter(input.Filters)

	return
}
