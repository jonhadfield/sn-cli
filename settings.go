package sncli

import (
	"github.com/jonhadfield/gosn"
)

func (input *GetSettingsConfig) Run() (settings gosn.Items, err error) {
	getItemsInput := gosn.GetItemsInput{
		Session: input.Session,
		Debug: input.Debug,
	}

	var output gosn.GetItemsOutput

	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return nil, err
	}

	output.Items.DeDupe()

	settings, err = output.Items.DecryptAndParse(input.Session.Mk, input.Session.Ak, input.Debug)
	if err != nil {
		return nil, err
	}

	settings.Filter(input.Filters)

	return
}
