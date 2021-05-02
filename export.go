package sncli

import (
	"fmt"

	"github.com/asdine/storm/v3/q"
	gosn "github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

type ExportConfig struct {
	Session *cache.Session
	File    string
	Debug   bool
}

type ImportConfig struct {
	Session *cache.Session
	File    string
	Debug   bool
}

func (i ExportConfig) Run() error {
	// populate DB
	gii := cache.SyncInput{
		Session: i.Session,
		Debug:   i.Debug,
	}

	gio, err := Sync(gii, true)
	if err != nil {
		return err
	}

	defer func() {
		_ = gio.DB.Close()
	}()

	// DB now populated and open with pointer in session
	// strip deleted items
	var out gosn.Items

	// load all items
	var allPersistedItems cache.Items

	// only export undeleted tags and notes
	query := gio.DB.Select(q.Eq("Deleted", false),
		q.Or(
			q.Eq("ContentType", "Note"),
			q.Eq("ContentType", "Tag")),
	)

	if err = query.Find(&allPersistedItems); err != nil {
		return err
	}

	out, err = allPersistedItems.ToItems(i.Session)
	if err != nil {
		return err
	}

	var e gosn.EncryptedItems

	e, err = out.Encrypt(i.Session.Gosn())
	if err != nil {
		return err
	}

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

	var encItemsToImport gosn.EncryptedItems

	err = readGob(i.File, &encItemsToImport)
	if err != nil {
		return err
	}

	var encFinalList gosn.EncryptedItems
	encFinalList = append(encFinalList, encItemsToImport...)

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
