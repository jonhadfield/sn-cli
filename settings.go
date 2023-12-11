package sncli

import (
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
)

func (i *GetSettingsConfig) Run() (settings items.Items, err error) {
	getItemsInput := cache.SyncInput{
		Session: i.Session,
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

	var items items.Items

	items, err = allPersistedItems.ToItems(i.Session)
	if err != nil {
		return
	}

	items.Filter(i.Filters)

	return
}
