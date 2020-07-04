package sncli

import (
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

func (i *GetSettingsConfig) Run() (settings gosn.Items, err error) {
	getItemsInput := cache.SyncInput{
		Session: i.Session,
		Debug:   i.Debug,
	}

	var so cache.SyncOutput

	so, err = Sync(getItemsInput, true)
	if err != nil {
		return nil, err
	}

	var allPersistedItems cache.Items

	err = so.DB.All(&allPersistedItems)
	if err != nil {
		return
	}

	var items gosn.Items

	items, err = allPersistedItems.ToItems(i.Session.Mk, i.Session.Ak)
	if err != nil {
		return
	}

	items.Filter(i.Filters)

	return
}
