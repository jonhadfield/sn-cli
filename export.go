package sncli

import (
	"github.com/jonhadfield/gosn-v2"
)

type ExportConfig struct {
	Session gosn.Session
	File    string
	Debug   bool
}

type ImportConfig struct {
	Session gosn.Session
	File    string
	Debug   bool
}

func (i *ExportConfig) Run() error {
	gii := gosn.SyncInput{
		Session: i.Session,
		Debug:   i.Debug,
	}

	gio, err := gosn.Sync(gii)
	if err != nil {
		return err
	}
	// strip deleted items
	var out gosn.EncryptedItems

	for _, i := range gio.Items {
		if !i.Deleted {
			out = append(out, i)
		}
	}

	return writeGob(i.File, out)
}

func (i *ImportConfig) Run() error {
	var encItemsToImport gosn.EncryptedItems

	err := readGob(i.File, &encItemsToImport)

	if err != nil {
		return err
	}

	var itemsToImport gosn.Items

	itemsToImport, err = encItemsToImport.DecryptAndParse(i.Session.Mk, i.Session.Ak, i.Debug)
	if err != nil {
		return err
	}

	// get existing encItemsToImport
	var existingItems gosn.Items

	gii := gosn.SyncInput{
		Session: i.Session,
	}

	var gio gosn.SyncOutput

	gio, err = gosn.Sync(gii)

	if err != nil {
		return err
	}
	gio.Items = filterByTypes(gio.Items, supportedContentTypes)

	existingItems, err = gio.Items.DecryptAndParse(i.Session.Mk, i.Session.Ak, i.Debug)

	if err != nil {
		return err
	}

	var finalList gosn.Items
	// for each (tag and note) item to import, check if uuid exists
	for _, itemToImport := range itemsToImport {
		var done, found bool

		for _, existingItem := range existingItems {
			// if uuid exists
			if existingItem != nil && itemToImport != nil && itemToImport.GetUUID() == existingItem.GetUUID() {
				// if item deleted, push with new uuid
				found = true

				if existingItem.IsDeleted() {
					itemToImport.SetUUID(gosn.GenUUID())
					finalList = append(finalList, itemToImport)
					done = true
				} else {
					// just push so it replaces existing
					finalList = append(finalList, itemToImport)
					done = true
				}
			}

			if done {
				break
			}
		}
		// if uuid does not match then just add
		if !found {
			finalList = append(finalList, itemToImport)
		}
	}

	var encFinalList gosn.EncryptedItems
	
	encFinalList, err = finalList.Encrypt(i.Session.Mk, i.Session.Ak, i.Debug)
	if err != nil {
		return err
	}
	// push item
	pii := gosn.SyncInput{
		Session: i.Session,
		Items:   encFinalList,
	}

	_, err = gosn.Sync(pii)

	return err
}
