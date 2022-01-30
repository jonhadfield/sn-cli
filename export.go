package sncli

import (
	"fmt"
	"github.com/jonhadfield/gosn-v2/cache"
)

type ExportConfig struct {
	Session   *cache.Session
	Decrypted bool
	File      string
	Debug     bool
}

type ImportConfig struct {
	Session *cache.Session
	File    string
	Format  string
	Debug   bool
}

// Run will retrieve all items from SN directly, re-encrypt them with a new ItemsKey and write them to a file.
func (i ExportConfig) Run() error {
	return i.Session.Export(i.File)
}

func (i *ImportConfig) Run() (imported int, err error) {
	// populate DB
	gii := cache.SyncInput{
		Session: i.Session,
	}
	gio, err := Sync(gii, true)
	if err != nil {
		return imported, err
	}

	var syncTokens []cache.SyncToken
	if err = gio.DB.All(&syncTokens); err != nil {
		return imported, err
	}
	syncToken := ""
	if len(syncTokens) > 0 {
		syncToken = syncTokens[0].SyncToken
	}
	if err = gio.DB.Close(); err != nil {
		return imported, err
	}

	iItems, iItemsKey, err := i.Session.Session.Import(i.File, syncToken, "")
	if err != nil {
		return
	}

	if iItemsKey.ItemsKey == "" {
		panic(fmt.Sprintf("iItemsKey.ItemsKey is empty for: '%s'", iItemsKey.UUID))
	}

	// push item and close db
	pii := cache.SyncInput{
		Session: i.Session,
		Close:   true,
	}

	_, err = Sync(pii, true)
	imported = len(iItems)

	return imported, err
}
