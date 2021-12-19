package sncli

import (
	"fmt"

	"github.com/asdine/storm/v3/q"
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

type ExportConfig struct {
	Session   *cache.Session
	Decrypted bool
	Format    string
	File      string
	Debug     bool
}

type ImportConfig struct {
	Session *cache.Session
	File    string
	Format  string
	Debug   bool
}

func (i ExportConfig) Run() error {
	// populate DB
	gii := cache.SyncInput{
		Session: i.Session,
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

	var w interface{}

	if i.Decrypted {
		w = out
	} else {
		w, err = out.Encrypt(i.Session.Gosn())
		if err != nil {
			return err
		}
	}

	switch i.Format {
	case "gob":
		if err = writeGob(i.File, w) ; err == nil {
			return err
		}
	case "json":
		if err = writeJSON(i.File, w) ; err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid format specified: '%s'", i.Format)
	}

	fmt.Printf("export written to: %s\n", i.File)

	return nil
}

func (i *ImportConfig) Run() error {
	// populate DB
	gii := cache.SyncInput{
		Session: i.Session,
	}

	gio, err := Sync(gii, true)
	if err != nil {
		return err
	}

	var encItemsToImport gosn.EncryptedItems

	switch i.Format {
	case "gob":
		if err = readGob(i.File, encItemsToImport) ; err == nil {
			return err
		}
	case "json":
		if err = readJSON(i.File, &encItemsToImport) ; err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid format specified: '%s'", i.Format)
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
