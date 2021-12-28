package sncli

import (
	"fmt"
	"github.com/jonhadfield/gosn-v2"
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

func (i ExportConfig) Run() error {
	// populate DB
	gii := gosn.SyncInput{
		Session: i.Session.Session,
	}

	gio, err := gosn.Sync(gii)
	if err != nil {
		return err
	}

	// DB now populated and open with pointer in session
	// strip deleted items and re-encrypt
	var out gosn.EncryptedItems
	for _, item := range gio.Items {
		if item.Deleted == false {
			// set new uuid
			out = append(out, item)
		}
	}

	// decrypt
	di, err := out.DecryptAndParse(i.Session.Session)
	if err != nil {
		return err
	}
	// encrypt
	nei, err := di.Encrypt(*i.Session.Session)
	if err != nil {
		return err
	}

	if err = writeJSON(i, nei); err != nil {
		return err
	}

	fmt.Printf("export written to: %s\n", i.File)

	return nil
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

	var encItemsToImport gosn.EncryptedItems

	switch i.Format {
	case "gob":
		if err = readGob(i.File, &encItemsToImport); err != nil {
			return imported, fmt.Errorf("%w", err)
		}
	case "json":
		encItemsToImport, err = readJSON(i.File)
		if err != nil {
			return imported, fmt.Errorf("%w", err)
		}
	default:
		return imported, fmt.Errorf("invalid format specified: '%s'", i.Format)
	}

	var encFinalList gosn.EncryptedItems
	if encItemsToImport != nil {
		for _, item := range encItemsToImport {
			if item.DuplicateOf != nil {
				err = fmt.Errorf("duplicate of item found: %s", *item.DuplicateOf)
			}
		}
		encFinalList = append(encFinalList, encItemsToImport...)
	}

	if len(encFinalList) == 0 {
		return imported, fmt.Errorf("no items to import were loaded")
	}

	if err = cache.SaveEncryptedItems(gio.DB, encFinalList, true); err != nil {
		return imported, err
	}

	// push item and close db
	pii := cache.SyncInput{
		Session: i.Session,
		Close:   true,
	}

	_, err = Sync(pii, true)
	imported = len(encFinalList)

	return imported, err
}
