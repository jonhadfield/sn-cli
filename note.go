package sncli

import (
	"github.com/jonhadfield/gosn"
	//"log"
)

func (input *AddNoteConfig) Run() error {
	//gosn.SetErrorLogger(log.Println)
	//if input.Debug {
	//	gosn.SetDebugLogger(log.Println)
	//}

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
	tagUUIDs  []string
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

	pii := gosn.PutItemsInput{
		Session:   input.session,
		SyncToken: input.syncToken,
		Items:     []gosn.Item{*newNote},
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

	return
}

func (input *DeleteNoteConfig) Run() error {
	//gosn.SetErrorLogger(log.Println)
	//if input.Debug {
	//	gosn.SetDebugLogger(log.Println)
	//}
	var err error
	_, err = deleteNotes(input.Session, input.NoteTitles, input.NoteText, input.NoteUUIDs, input.Regex, "")
	return err
}

func (input *GetNoteConfig) Run() (output gosn.GetItemsOutput, err error) {
	//gosn.SetErrorLogger(log.Println)
	//if input.Debug {
	//	gosn.SetDebugLogger(log.Println)
	//}
	getItemsInput := gosn.GetItemsInput{
		PageSize:  input.PageSize,
		BatchSize: input.BatchSize,
		Session:   input.Session,
		Filters:   input.Filters,
	}
	output, err = gosn.GetItems(getItemsInput)
	output.DeDupe()
	return
}

// TODO: don't pass match criteria here, pass actual items to mark Deleted flag as true for and then PutItems them
func deleteNotes(session gosn.Session, noteTitles []string, noteText string, noteUUIDs []string, regex bool, syncToken string) (newSyncToken string, err error) {
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
		Filters:   itemFilter,
	}
	output, err := gosn.GetItems(getItemsInput)
	if err != nil {
		return
	}
	output.DeDupe()
	var notesToDelete []gosn.Item
	for _, item := range output.Items {
		if item.Content != nil && item.ContentType == "Note" {
			item.Deleted = true
			notesToDelete = append(notesToDelete, item)
		}
	}
	if len(notesToDelete) > 0 {
		pii := gosn.PutItemsInput{
			Session:   session,
			Items:     notesToDelete,
			SyncToken: syncToken,
		}
		var putItemsOutput gosn.PutItemsOutput
		putItemsOutput, err = gosn.PutItems(pii)
		if err != nil {
			return
		}
		newSyncToken = putItemsOutput.ResponseBody.SyncToken
	}
	return
}
