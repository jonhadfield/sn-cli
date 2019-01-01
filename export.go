package sncli

import (
	"log"

	"github.com/jonhadfield/gosn"
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
	if i.Debug {
		gosn.SetDebugLogger(log.Println)
	}
	gii := gosn.GetItemsInput{
		Session: i.Session,
	}
	gio, err := gosn.GetItems(gii)
	err = writeGob(i.File, gio.Items)
	return err
}

func (i *ImportConfig) Run() error {
	if i.Debug {
		gosn.SetDebugLogger(log.Println)
	}
	var encItemsToImport gosn.EncryptedItems
	err := readGob(i.File, &encItemsToImport)
	if err != nil {
		return err
	}
	var decItemsToImport gosn.DecryptedItems
	decItemsToImport, err = encItemsToImport.Decrypt(i.Session.Mk, i.Session.Ak)
	if err != nil {
		return err
	}
	var itemsToImport gosn.Items
	itemsToImport, err = decItemsToImport.Parse()
	if err != nil {
		return err
	}

	// get existing encItemsToImport
	var existingDecryptedItems gosn.DecryptedItems
	var existingItems gosn.Items
	gii := gosn.GetItemsInput{
		Session: i.Session,
	}
	var gio gosn.GetItemsOutput
	gio, err = gosn.GetItems(gii)
	if err != nil {
		return err
	}
	existingDecryptedItems, err = gio.Items.Decrypt(i.Session.Mk, i.Session.Ak)
	existingItems, err = existingDecryptedItems.Parse()
	if err != nil {
		return err
	}
	var finalList gosn.Items
	// for each (tag and note) item to import, check if uuid exists
	for _, itemToImport := range itemsToImport {
		var done, found bool
		for _, existingItem := range existingItems {
			// if uuid exists
			if itemToImport.UUID == existingItem.UUID {
				// if item deleted, push with new uuid
				found = true
				if existingItem.Deleted {
					itemToImport.UUID = gosn.GenUUID()
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
	encFinalList, err = finalList.Encrypt(i.Session.Mk, i.Session.Ak)
	// push item
	pii := gosn.PutItemsInput{
		Session: i.Session,
		Items:   encFinalList,
	}

	_, err = gosn.PutItems(pii)
	return err
}
