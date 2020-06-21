package sncli

import (
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

func (input *AddNoteInput) Run() (err error) {
	// get DB
	var syncToken, newNoteUUID string

	ani := addNoteInput{
		noteTitle: input.Title,
		noteText:  input.Text,
		tagTitles: input.Tags,
		session:   input.Session,
	}

	newNoteUUID, err = addNote(ani)
	if err != nil {
		return
	}

	if len(ani.tagTitles) > 0 {
		tni := tagNotesInput{
			matchNoteUUIDs: []string{newNoteUUID},
			syncToken:      syncToken,
			session:        input.Session,
			newTags:        input.Tags,
		}
		err = tagNotes(tni)
	}

	return
}

type addNoteInput struct {
	session   cache.Session
	noteTitle string
	noteText  string
	tagTitles []string
}

func addNote(input addNoteInput) (noteUUID string, err error) {
	// check if note exists
	newNote := gosn.NewNote()
	newNoteContent := gosn.NewNoteContent()
	newNoteContent.Title = input.noteTitle
	newNoteContent.Text = input.noteText
	newNote.Content = *newNoteContent
	newNote.UUID = gosn.GenUUID()
	noteUUID = newNote.UUID
	newNoteItems := gosn.Notes{newNote}

	var eNewNoteItems gosn.EncryptedItems

	eNewNoteItems, err = newNoteItems.Encrypt(input.session.Mk, input.session.Ak, false)
	if err != nil {
		return
	}

	si := cache.SyncInput{
		Session: input.session,
	}

	var so cache.SyncOutput

	so, err = Sync(si, true)
	if err != nil {
		return
	}

	if err = cache.SaveEncryptedItems(so.DB, eNewNoteItems, true); err != nil {
		return
	}

	pii := cache.SyncInput{
		Session: input.session,
	}

	so, err = Sync(pii, true)
	if err != nil {
		return
	}
	defer func() {
		_ = so.DB.Close()
	}()

	if len(input.tagTitles) > 0 {
		tni := tagNotesInput{
			session:        input.session,
			matchNoteUUIDs: []string{newNote.UUID},
			newTags:        input.tagTitles,
		}

		err = tagNotes(tni)
		if err != nil {
			return
		}
	}

	return noteUUID, err
}

func (input *DeleteNoteConfig) Run() (noDeleted int, err error) {
	noDeleted, err = deleteNotes(input.Session, input.NoteTitles, input.NoteText, input.NoteUUIDs, input.Regex, "", input.Debug)

	return noDeleted, err
}

func (input *GetNoteConfig) Run() (items gosn.Items, err error) {
	var so cache.SyncOutput
	so, err = Sync(cache.SyncInput{
		Session: input.Session,
		Debug:   input.Debug,
	}, true)

	if err != nil {
		return
	}

	var allPersistedItems cache.Items

	err = so.DB.All(&allPersistedItems)
	if err != nil {
		return
	}
	defer so.DB.Close()

	items, err = allPersistedItems.ToItems(input.Session.Mk, input.Session.Ak)
	if err != nil {
		return
	}

	items.Filter(input.Filters)

	return
}

func deleteNotes(session cache.Session, noteTitles []string, noteText string, noteUUIDs []string, regex bool, syncToken string, debug bool) (noDeleted int, err error) {
	var getNotesFilters []gosn.Filter

	switch {
	case len(noteTitles) > 0:
		for _, title := range noteTitles {
			comparison := "=="
			if regex {
				comparison = "~"
			}

			getNotesFilters = append(getNotesFilters, gosn.Filter{
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

		getNotesFilters = append(getNotesFilters, gosn.Filter{
			Key:        "Text",
			Value:      noteText,
			Comparison: comparison,
			Type:       "Note",
		})
	case len(noteUUIDs) > 0:
		for _, uuid := range noteUUIDs {
			getNotesFilters = append(getNotesFilters, gosn.Filter{
				Key:        "UUID",
				Value:      uuid,
				Comparison: "==",
				Type:       "Note",
			})
		}
	}

	itemFilter := gosn.ItemFilters{
		Filters:  getNotesFilters,
		MatchAny: true,
	}

	getItemsInput := cache.SyncInput{
		Session: session,
		Debug:   debug,
	}

	var gio cache.SyncOutput

	gio, err = Sync(getItemsInput, true)
	if err != nil {
		return
	}

	var allPersistedItems cache.Items

	err = gio.DB.All(&allPersistedItems)
	if err != nil {
		return
	}

	var notes gosn.Items

	notes, err = allPersistedItems.ToItems(session.Mk, session.Ak)
	if err != nil {
		return
	}

	notes.Filter(itemFilter)

	var notesToDelete gosn.Notes

	for _, item := range notes {
		note := item.(*gosn.Note)

		if note.GetContent() != nil {
			note.Content.SetText("")
			note.SetDeleted(true)
			notesToDelete = append(notesToDelete, *note)
		}
	}

	if notesToDelete == nil {
		return
	}

	var eNotesToDelete gosn.EncryptedItems

	eNotesToDelete, err = notesToDelete.Encrypt(session.Mk, session.Ak, false)
	if err != nil {
		return
	}

	err = cache.SaveEncryptedItems(gio.DB, eNotesToDelete, true)
	if err != nil {
		return 0, err
	}

	pii := cache.SyncInput{
		Session: session,
	}

	gio, err = Sync(pii, true)
	if err != nil {
		return
	}

	_ = gio.DB.Close()

	return len(notesToDelete), err
}
