package sncli

import (
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

func (input *GetSettingsConfig) Run() (settings gosn.Items, err error) {
	getItemsInput := cache.SyncInput{
		Session: input.Session,
		Debug:   input.Debug,
	}

	var so cache.SyncOutput

	so, err = Sync(getItemsInput, true)
	if err != nil {
		return nil, err
	}

	var allPersistedItems cache.Items
	err = so.DB.All(&allPersistedItems)

	var items gosn.Items

	items, err = allPersistedItems.ToItems(input.Session.Mk, input.Session.Ak)
	if err != nil {
		return
	}

	items.Filter(input.Filters)

	return
}
