package sncli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/asdine/storm/v3/q"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
)

func (i *AddNoteInput) Run() (err error) {
	// get DB
	var syncToken, newNoteUUID string

	ani := addNoteInput{
		noteTitle: i.Title,
		noteText:  i.Text,
		tagTitles: i.Tags,
		filePath:  i.FilePath,
		session:   i.Session,
		replace:   i.Replace,
	}

	newNoteUUID, err = addNote(ani)
	if err != nil {
		return
	}

	if len(ani.tagTitles) > 0 {
		err = tagNotes(tagNotesInput{
			matchNoteUUIDs: []string{newNoteUUID},
			syncToken:      syncToken,
			session:        i.Session,
			newTags:        i.Tags,
			replace:        i.Replace,
		})
	}

	return
}

type addNoteInput struct {
	session   *cache.Session
	noteTitle string
	noteText  string
	filePath  string
	tagTitles []string
	replace   bool
}

func loadNoteContentFromFile(filePath string) (content string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		err = fmt.Errorf("%w failed to open: %s", err, filePath)
		return
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	return string(b), nil
}

func addNote(i addNoteInput) (noteUUID string, err error) {
	// if file path provided, try loading content as note text
	if i.filePath != "" {
		if i.noteText, err = loadNoteContentFromFile(i.filePath); err != nil {
			return
		}

		if i.noteTitle == "" {
			i.noteTitle = filepath.Base(i.filePath)
		}
	}

	var noteToAdd items.Note

	si := cache.SyncInput{
		Session: i.session,
		Close:   true,
	}

	var so cache.SyncOutput

	so, err = Sync(si, true)
	if err != nil {
		return
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
						Type:       "Note",
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
			return
		}
		switch len(gi) {
		case 0:
			err = fmt.Errorf("failed to find existing note to replace")

			return
		case 1:
			noteToAdd = gi.Notes()[0]
			noteToAdd.Content.SetText(i.noteText)
		default:
			err = fmt.Errorf("multiple notes found with that title")

			return
		}
	} else {
		noteToAdd, err = items.NewNote(i.noteTitle, i.noteText, nil)
		if err != nil {
			return
		}
		noteUUID = noteToAdd.UUID
	}

	si = cache.SyncInput{
		Session: i.session,
		Close:   false,
	}

	so, err = Sync(si, true)
	if err != nil {
		return
	}

	var allEncTags cache.Items

	query := so.DB.Select(q.And(q.Eq("ContentType", "Tag"), q.Eq("Deleted", false)))

	err = query.Find(&allEncTags)
	// it's ok if there are no tags, so only error if something else went wrong
	if err != nil && err.Error() != "not found" {
		return
	}

	if err = cache.SaveNotes(i.session, so.DB, items.Notes{noteToAdd}, false); err != nil {
		return
	}

	_ = so.DB.Close()

	pii := cache.SyncInput{
		Session: i.session,
	}

	so, err = Sync(pii, true)
	if err != nil {
		return
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
			return
		}
	}

	return noteUUID, err
}

func (i *DeleteItemConfig) Run() (noDeleted int, err error) {
	noDeleted, err = deleteItems(i.Session, []string{}, "", i.ItemsUUIDs, i.Regex)

	return noDeleted, err
}

func (i *DeleteNoteConfig) Run() (noDeleted int, err error) {
	noDeleted, err = deleteNotes(i.Session, i.NoteTitles, i.NoteText, i.NoteUUIDs, i.Regex)

	return noDeleted, err
}

func (i *GetNoteConfig) Run() (items items.Items, err error) {
	var so cache.SyncOutput
	so, err = Sync(cache.SyncInput{
		Session: i.Session,
	}, true)

	if err != nil {
		return
	}

	var allPersistedItems cache.Items

	err = so.DB.All(&allPersistedItems)
	if err != nil {
		err = fmt.Errorf("getting items from db: %w", err)
		return
	}

	defer func() {
		_ = so.DB.Close()
	}()

	items, err = allPersistedItems.ToItems(i.Session)
	if err != nil {
		return
	}

	items.Filter(i.Filters)

	return
}

func deleteNotes(session *cache.Session, noteTitles []string, noteText string, noteUUIDs []string, regex bool) (noDeleted int, err error) {
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
				Type:       "Note",
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
			Type:       "Note",
		})
	case len(noteUUIDs) > 0:
		for _, uuid := range noteUUIDs {
			getNotesFilters = append(getNotesFilters, items.Filter{
				Key:        "UUID",
				Value:      uuid,
				Comparison: "==",
				Type:       "Note",
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
		return
	}

	var allPersistedItems cache.Items

	err = gio.DB.All(&allPersistedItems)
	if err != nil {
		err = fmt.Errorf("getting items from db: %w", err)
		return
	}

	var notes items.Items

	notes, err = allPersistedItems.ToItems(session)
	if err != nil {
		return
	}

	notes.Filter(itemFilter)
	var notesToDelete items.Notes

	for _, item := range notes {
		if item.GetContentType() != "Note" {
			panic(fmt.Sprintf("Got a non-note item in the notes list: %s", item.GetContentType()))
		}
		note := item.(*items.Note)
		if note.GetContent() != nil {
			note.Content.SetText("")
			note.SetDeleted(true)
			notesToDelete = append(notesToDelete, *note)
		}
	}

	if notesToDelete == nil {
		return
	}

	if err = cache.SaveNotes(session, gio.DB, notesToDelete, true); err != nil {
		return 0, err
	}

	pii := cache.SyncInput{
		Session: session,
	}

	pii.Close = true
	_, err = Sync(pii, true)
	if err != nil {
		return
	}

	return len(notesToDelete), err
}

func deleteItems(session *cache.Session, noteTitles []string, noteText string, itemUUIDs []string, regex bool) (noDeleted int, err error) {
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
				Type:       "Note",
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
			Type:       "Note",
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
		return
	}

	var allPersistedItems cache.Items

	err = gio.DB.All(&allPersistedItems)
	if err != nil {
		err = fmt.Errorf("getting items from db: %w", err)
		return
	}

	var pItems items.Items

	pItems, err = allPersistedItems.ToItems(session)
	if err != nil {
		return
	}

	pItems.Filter(itemFilter)
	var itemsToDelete items.Items

	for _, pItem := range pItems {
		pItem.SetDeleted(true)
		itemsToDelete = append(itemsToDelete, pItem)
	}

	if itemsToDelete == nil {
		return
	}

	if err = cache.SaveItems(session, gio.DB, itemsToDelete, true); err != nil {
		return 0, err
	}

	pii := cache.SyncInput{
		Session: session,
	}

	pii.Close = true
	_, err = Sync(pii, true)
	if err != nil {
		return
	}

	return len(itemsToDelete), err
}
