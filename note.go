package sncli

import (
	"fmt"
	"github.com/asdine/storm/v3/q"
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

func (i *AddNoteInput) Run() (err error) {
	// get DB
	var syncToken, newNoteUUID string

	ani := addNoteInput{
		noteTitle: i.Title,
		noteText:  i.Text,
		tagTitles: i.Tags,
		session:   i.Session,
	}

	newNoteUUID, err = addNote(ani)
	if err != nil {
		return
	}

	if len(ani.tagTitles) > 0 {
		tni := tagNotesInput{
			matchNoteUUIDs: []string{newNoteUUID},
			syncToken:      syncToken,
			session:        i.Session,
			newTags:        i.Tags,
		}
		err = tagNotes(tni)
	}

	return
}

type addNoteInput struct {
	session   *cache.Session
	noteTitle string
	noteText  string
	tagTitles []string
}

func addNote(i addNoteInput) (noteUUID string, err error) {
	// check if note exists
	newNote := gosn.NewNote()
	newNoteContent := gosn.NewNoteContent()
	newNoteContent.Title = i.noteTitle
	newNoteContent.Text = i.noteText
	newNote.Content = *newNoteContent
	newNote.UUID = gosn.GenUUID()
	noteUUID = newNote.UUID
	newNoteItems := gosn.Notes{newNote}

	si := cache.SyncInput{
		Session: i.session,
	}

	var so cache.SyncOutput

	so, err = Sync(si, true)
	if err != nil {
		return
	}
	// get items key
	var allItemsKeys cache.Items

	kquery := so.DB.Select(q.And(q.Eq("ContentType", "SN|ItemsKey"), q.Eq("Deleted", false)))

	err = kquery.Find(&allItemsKeys)
	if err != nil {
		err = fmt.Errorf("no items keys were found")

		return
	}

	var allEncTags cache.Items

	query := so.DB.Select(q.And(q.Eq("ContentType", "Tag"), q.Eq("Deleted", false)))

	err = query.Find(&allEncTags)
	// it's ok if there are no tags, so only error if something else went wrong
	if err != nil && err.Error() != "not found" {
		return
	}

	if err = cache.SaveNotes(i.session, so.DB, newNoteItems, false); err != nil {
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
		tni := tagNotesInput{
			session:        i.session,
			matchNoteUUIDs: []string{newNote.UUID},
			newTags:        i.tagTitles,
		}

		err = tagNotes(tni)
		if err != nil {
			return
		}
	}

	return noteUUID, err
}

func (i *DeleteNoteConfig) Run() (noDeleted int, err error) {
	noDeleted, err = deleteNotes(i.Session, i.NoteTitles, i.NoteText, i.NoteUUIDs, i.Regex)

	return noDeleted, err
}

func (i *GetNoteConfig) Run() (items gosn.Items, err error) {
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

	notes, err = allPersistedItems.ToItems(session)
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

	if err = cache.SaveNotes(session, gio.DB, notesToDelete, true); err != nil {
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
