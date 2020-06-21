package sncli

import (
	"fmt"
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

type ExportConfig struct {
	Session cache.Session
	File    string
	Debug   bool
}

type ImportConfig struct {
	Session   cache.Session
	Overwrite bool
	File      string
	Debug     bool
}

func (i *ExportConfig) Run() error {
	// populate DB
	gii := cache.SyncInput{
		Session: i.Session,
		Debug:   i.Debug,
	}

	gio, err := Sync(gii, true)
	if err != nil {
		return err
	}
	defer gio.DB.Close()

	// DB now populated and open with pointer in session
	// strip deleted items
	var out gosn.Items

	// load all items
	var allPersistedItems cache.Items

	// only export undeleted tags and notes
	err = gio.DB.Find("Deleted", false, &allPersistedItems)

	out, err = allPersistedItems.ToItems(i.Session.Mk, i.Session.Ak)
	if err != nil {
		return err
	}

	var e gosn.EncryptedItems
	e, err = out.Encrypt(i.Session.Mk, i.Session.Ak, i.Debug)

	return writeGob(i.File, e)
}

func (i *ImportConfig) Run() error {
	// populate DB
	gii := cache.SyncInput{
		Session: i.Session,
		Debug:   i.Debug,
	}

	gio, err := Sync(gii, true)
	if err != nil {
		return err
	}

	// DB now populated and open with pointer in session
	var existingItems cache.Items
	err = gio.DB.Find("Deleted", false, &existingItems)

	var encItemsToImport gosn.EncryptedItems

	err = readGob(i.File, &encItemsToImport)

	if err != nil {
		return err
	}

	var encFinalList gosn.EncryptedItems

	// get existing encItemsToImport
	var rawItems gosn.Items
	rawItems, err = existingItems.ToItems(i.Session.Mk, i.Session.Ak)

	if err != nil {
		return err
	}

	rawItems = filterItemsByTypes(rawItems, supportedContentTypes)

	// TODO: Handle import of an item with same UUID
	// TODO: Decrypt conflicting item (if note or tag), create a copy with new uuid, encrypt and import
	// for each (tag and note) item to import, check if uuid exists
	for _, itemToImport := range encItemsToImport {
		encFinalList = append(encFinalList, itemToImport)
	}

	if len(encFinalList) == 0 {
		return fmt.Errorf("no items to import were loaded")
	}

	if err = cache.SaveEncryptedItems(gio.DB, encFinalList, true); err != nil {
		return err
	}

	// push item and close db
	pii := cache.SyncInput{
		Session: i.Session,
		Close:   true,
	}
	_, err = Sync(pii, true)

	return err
}
