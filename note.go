package sncli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/asdine/storm/v3/q"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

func (i *AddNoteInput) Run() error {
	// get DB
	var syncToken string

	ani := addNoteInput{
		noteTitle: i.Title,
		noteText:  i.Text,
		tagTitles: i.Tags,
		filePath:  i.FilePath,
		session:   i.Session,
		replace:   i.Replace,
	}

	newNoteUUID, err := addNote(ani)
	if err != nil {
		return err
	}

	if len(ani.tagTitles) > 0 {
		if err = tagNotes(tagNotesInput{
			matchNoteUUIDs: []string{newNoteUUID},
			syncToken:      syncToken,
			session:        i.Session,
			newTags:        i.Tags,
			replace:        i.Replace,
		}); err != nil {
			return err
		}
	}

	return nil
}

type addNoteInput struct {
	session   *cache.Session
	noteTitle string
	noteText  string
	filePath  string
	tagTitles []string
	replace   bool
}

func loadNoteContentFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("%w failed to open: %s", err, filePath)
	}

	b, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("%w failed to read: %s", err, filePath)
	}

	return string(b), nil
}

func addNote(i addNoteInput) (string, error) {
	// if file path provided, try loading content as note text
	if i.filePath != "" {
		if i.noteText, err = loadNoteContentFromFile(i.filePath); err != nil {
			return "", err
		}

		if i.noteTitle == "" {
			i.noteTitle = filepath.Base(i.filePath)
		}
	}

	var noteToAdd items.Note
	var noteUUID string

	si := cache.SyncInput{
		Session: i.session,
		Close:   true,
	}

	var so cache.SyncOutput

	so, err = Sync(si, true)
	if err != nil {
		return "", err
	}

	// if we're replacing, then retrieve note to update
	if i.replace {
		gnc := GetNoteConfig{
			Session: i.session,
			// Filters: items.ItemFilters{},
			Filters: items.ItemFilters{
				MatchAny: false,
				Filters: []items.Filter{
					{
						Type:       common.SNItemTypeNote,
						Key:        "Title",
						Comparison: "==",
						Value:      i.noteTitle,
					},
				},
			},
		}

		var gi items.Items

		gi, err = gnc.Run()
		if err != nil {
			return "", err
		}

		switch len(gi) {
		case 0:
			return "", errors.New("failed to find existing note to replace")
		case 1:
			noteToAdd = gi.Notes()[0]
			noteToAdd.Content.SetText(i.noteText)
		default:
			return "", errors.New("multiple notes found with that title")
		}
	} else {
		noteToAdd, err = items.NewNote(i.noteTitle, i.noteText, nil)
		if err != nil {
			return "", err
		}
		noteUUID = noteToAdd.UUID
	}

	si = cache.SyncInput{
		Session: i.session,
		Close:   false,
	}

	so, err = Sync(si, true)
	if err != nil {
		return "", err
	}

	var allEncTags cache.Items

	query := so.DB.Select(q.And(q.Eq("ContentType", common.SNItemTypeTag), q.Eq("Deleted", false)))

	err = query.Find(&allEncTags)
	// it's ok if there are no tags, so only error if something else went wrong
	if err != nil && err.Error() != "not found" {
		return "", err
	}

	if err = cache.SaveNotes(i.session, so.DB, items.Notes{noteToAdd}, false); err != nil {
		return "", err
	}

	_ = so.DB.Close()

	pii := cache.SyncInput{
		Session: i.session,
		Close:   false,
	}

	so, err = Sync(pii, true)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = so.DB.Close()
	}()

	if len(i.tagTitles) > 0 {
		_ = so.DB.Close()
		tni := tagNotesInput{
			session:        i.session,
			matchNoteUUIDs: []string{noteToAdd.UUID},
			newTags:        i.tagTitles,
		}

		err = tagNotes(tni)
		if err != nil {
			return "", err
		}
	}

	return noteUUID, err
}

func (i *DeleteItemConfig) Run() (int, error) {
	noDeleted, err := deleteItems(i.Session, []string{}, "", i.ItemsUUIDs, i.Regex)

	return noDeleted, err
}

func (i *DeleteNoteConfig) Run() (int, error) {
	return deleteNotes(i.Session, i.NoteTitles, i.NoteText, i.NoteUUIDs, i.Regex)
}

func (i *GetNoteConfig) Run() (items.Items, error) {
	var so cache.SyncOutput
	var err error
	so, err = Sync(cache.SyncInput{
		Session: i.Session,
	}, true)
	if err != nil {
		return nil, err
	}

	var allPersistedItems cache.Items

	err = so.DB.All(&allPersistedItems)
	if err != nil {
		return nil, fmt.Errorf("getting items from db: %w", err)
	}

	defer func() {
		_ = so.DB.Close()
	}()

	items, err := allPersistedItems.ToItems(i.Session)
	if err != nil {
		return nil, err
	}

	items.Filter(i.Filters)

	return items, nil
}

func deleteNotes(session *cache.Session, noteTitles []string, noteText string, noteUUIDs []string, regex bool) (int, error) {
	var getNotesFilters []items.Filter

	switch {
	case len(noteTitles) > 0:
		for _, title := range noteTitles {
			comparison := "=="
			if regex {
				comparison = "~"
			}

			getNotesFilters = append(getNotesFilters, items.Filter{
				Key:        "Title",
				Value:      title,
				Comparison: comparison,
				Type:       common.SNItemTypeNote,
			})
		}
	case noteText != "":
		comparison := "=="
		if regex {
			comparison = "~"
		}

		getNotesFilters = append(getNotesFilters, items.Filter{
			Key:        "Text",
			Value:      noteText,
			Comparison: comparison,
			Type:       common.SNItemTypeNote,
		})
	case len(noteUUIDs) > 0:
		for _, uuid := range noteUUIDs {
			getNotesFilters = append(getNotesFilters, items.Filter{
				Key:        "UUID",
				Value:      uuid,
				Comparison: "==",
				Type:       common.SNItemTypeNote,
			})
		}
	}

	itemFilter := items.ItemFilters{
		Filters:  getNotesFilters,
		MatchAny: true,
	}

	getItemsInput := cache.SyncInput{
		Session: session,
	}

	var gio cache.SyncOutput

	gio, err = Sync(getItemsInput, true)
	if err != nil {
		return 0, err
	}

	var allPersistedItems cache.Items

	err = gio.DB.All(&allPersistedItems)
	if err != nil {
		return 0, fmt.Errorf("getting items from db: %w", err)
	}

	var notes items.Items

	notes, err = allPersistedItems.ToItems(session)
	if err != nil {
		return 0, err
	}

	notes.Filter(itemFilter)

	var notesToDelete items.Notes

	for _, item := range notes {
		if item.GetContentType() != common.SNItemTypeNote {
			panic(fmt.Sprintf("got a non-note item in the notes list: %s", item.GetContentType()))
		}
		note := item.(*items.Note)
		if note.GetContent() != nil {
			note.Content.SetText("")
			note.SetDeleted(true)
			notesToDelete = append(notesToDelete, *note)
		}
	}

	if notesToDelete == nil || len(notesToDelete) == 0 {
		// close db as we're not going to save anything
		_ = session.CacheDB.Close()

		return 0, nil
	}

	if err = cache.SaveNotes(session, gio.DB, notesToDelete, true); err != nil {
		return 0, err
	}

	pii := cache.SyncInput{
		Session: session,
		Close:   true,
	}

	_, err = Sync(pii, true)
	if err != nil {
		return 0, err
	}

	return len(notesToDelete), err
}

func deleteItems(session *cache.Session, noteTitles []string, noteText string, itemUUIDs []string, regex bool) (int, error) {
	var getItemsFilters []items.Filter

	switch {
	case len(noteTitles) > 0:
		for _, title := range noteTitles {
			comparison := "=="
			if regex {
				comparison = "~"
			}

			getItemsFilters = append(getItemsFilters, items.Filter{
				Key:        "Title",
				Value:      title,
				Comparison: comparison,
				Type:       common.SNItemTypeNote,
			})
		}
	case noteText != "":
		comparison := "=="
		if regex {
			comparison = "~"
		}

		getItemsFilters = append(getItemsFilters, items.Filter{
			Key:        "Text",
			Value:      noteText,
			Comparison: comparison,
			Type:       common.SNItemTypeNote,
		})
	case len(itemUUIDs) > 0:
		for _, uuid := range itemUUIDs {
			getItemsFilters = append(getItemsFilters, items.Filter{
				Key:        "UUID",
				Value:      uuid,
				Comparison: "==",
				Type:       "Anything",
			})
		}
	}

	itemFilter := items.ItemFilters{
		Filters:  getItemsFilters,
		MatchAny: true,
	}

	getItemsInput := cache.SyncInput{
		Session: session,
	}

	var gio cache.SyncOutput

	gio, err = Sync(getItemsInput, true)
	if err != nil {
		return 0, err
	}

	var allPersistedItems cache.Items

	err = gio.DB.All(&allPersistedItems)
	if err != nil {
		return 0, fmt.Errorf("getting items from db: %w", err)
	}

	var pItems items.Items

	pItems, err = allPersistedItems.ToItems(session)
	if err != nil {
		return 0, err
	}

	pItems.Filter(itemFilter)
	var itemsToDelete items.Items

	for _, pItem := range pItems {
		pItem.SetDeleted(true)
		itemsToDelete = append(itemsToDelete, pItem)
	}

	if itemsToDelete == nil {
		return 0, nil
	}

	if err = cache.SaveItems(session, gio.DB, itemsToDelete, true); err != nil {
		return 0, err
	}

	pii := cache.SyncInput{
		Session: session,
		Close:   true,
	}

	_, err = Sync(pii, true)
	if err != nil {
		return 0, err
	}

	return len(itemsToDelete), err
}
