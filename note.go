package sncli

import (
	"github.com/jonhadfield/gosn"
)

func (input *AddNoteInput) Run() error {
	var syncToken, newNoteUUID string

	ani := addNoteInput{
		noteTitle: input.Title,
		noteText:  input.Text,
		tagTitles: input.Tags,
		syncToken: syncToken,
		session:   input.Session,
	}

	syncToken, newNoteUUID, err := addNote(ani)
	if err != nil {
		return err
	}

	if len(ani.tagTitles) > 0 {
		tni := tagNotesInput{
			matchNoteUUIDs: []string{newNoteUUID},
			syncToken:      syncToken,
			session:        input.Session,
			newTags:        input.Tags,
		}
		_, err = tagNotes(tni)
	}

	return err
}

type addNoteInput struct {
	session   gosn.Session
	noteTitle string
	noteText  string
	tagTitles []string
	syncToken string
}

func addNote(input addNoteInput) (newSyncToken, noteUUID string, err error) {
	// check if note exists
	newNote := gosn.NewNote()
	newNoteContent := gosn.NewNoteContent()
	newNoteContent.Title = input.noteTitle
	newNoteContent.Text = input.noteText
	newNote.Content = newNoteContent
	newNote.UUID = gosn.GenUUID()
	newNoteItems := gosn.Items{*newNote}

	var eNewNoteItems gosn.EncryptedItems

	eNewNoteItems, err = newNoteItems.Encrypt(input.session.Mk, input.session.Ak, false)
	if err != nil {
		return
	}

	pii := gosn.PutItemsInput{
		Session:   input.session,
		SyncToken: input.syncToken,
		Items:     eNewNoteItems,
	}

	var putItemsOutput gosn.PutItemsOutput

	putItemsOutput, err = gosn.PutItems(pii)
	if err != nil {
		return
	}

	newSyncToken = putItemsOutput.ResponseBody.SyncToken

	if len(input.tagTitles) > 0 {
		tni := tagNotesInput{
			session:        input.session,
			matchNoteUUIDs: []string{newNote.UUID},
			newTags:        input.tagTitles,
			syncToken:      newSyncToken,
		}

		_, err = tagNotes(tni)
		if err != nil {
			return
		}
	}

	return newSyncToken, noteUUID, err
}

func (input *DeleteNoteConfig) Run() (noDeleted int, err error) {
	noDeleted, _, err = deleteNotes(input.Session, input.NoteTitles, input.NoteText, input.NoteUUIDs, input.Regex, "")

	return noDeleted, err
}

func (input *GetNoteConfig) Run() (output gosn.Items, err error) {
	getItemsInput := gosn.GetItemsInput{
		PageSize:  input.PageSize,
		BatchSize: input.BatchSize,
		Session:   input.Session,
		Debug:     input.Debug,
	}

	var gio gosn.GetItemsOutput

	gio, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return
	}

	gio.Items.DeDupe()

	output, err = gio.Items.DecryptAndParse(input.Session.Mk, input.Session.Ak, input.Debug)
	if err != nil {
		return
	}

	output.Filter(input.Filters)

	return
}

func deleteNotes(session gosn.Session, noteTitles []string, noteText string, noteUUIDs []string, regex bool, syncToken string) (noDeleted int, newSyncToken string, err error) {
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

	getItemsInput := gosn.GetItemsInput{
		Session:   session,
		SyncToken: syncToken,
	}

	gio, err := gosn.GetItems(getItemsInput)
	if err != nil {
		return
	}

	gio.Items.DeDupe()
	ei := gio.Items

	var notes gosn.Items

	notes, err = ei.DecryptAndParse(session.Mk, session.Ak, false)
	if err != nil {
		return
	}

	notes.Filter(itemFilter)

	var notesToDelete gosn.Items

	for _, item := range notes {
		if item.Content != nil && item.ContentType == "Note" {
			item.Content.SetText("")
			item.Deleted = true
			notesToDelete = append(notesToDelete, item)
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

	pii := gosn.PutItemsInput{
		Session:   session,
		Items:     eNotesToDelete,
		SyncToken: syncToken,
	}

	var putItemsOutput gosn.PutItemsOutput

	putItemsOutput, err = gosn.PutItems(pii)
	if err != nil {
		return
	}

	noDeleted = len(notesToDelete)

	newSyncToken = putItemsOutput.ResponseBody.SyncToken

	return noDeleted, newSyncToken, err
}
