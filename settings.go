package sncli

import (
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
)

func (i *GetSettingsConfig) Run() (items.Items, error) {
	getItemsInput := cache.SyncInput{
		Session: i.Session,
	}

	var so cache.SyncOutput

	so, err := Sync(getItemsInput, true)
	if err != nil {
		return nil, err
	}

	var allPersistedItems cache.Items

	if err = so.DB.All(&allPersistedItems); err != nil {
		return nil, err
	}

	items, err := allPersistedItems.ToItems(i.Session)
	if err != nil {
		return nil, err
	}

	items.Filter(i.Filters)

	return items, nil
}
