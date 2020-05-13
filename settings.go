package sncli

import (
	"github.com/jonhadfield/gosn-v2"
)

func (input *GetSettingsConfig) Run() (settings gosn.Items, err error) {
	getItemsInput := gosn.SyncInput{
		Session: input.Session,
		Debug: input.Debug,
	}

	var output gosn.SyncOutput

	output, err = gosn.Sync(getItemsInput)
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
