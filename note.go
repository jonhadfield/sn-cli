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
	_, err = deleteNotes(input.Session, input.NoteTitles, input.NoteUUIDs, "")
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
func deleteNotes(session gosn.Session, noteTitles []string, noteUUIDs []string, syncToken string) (newSyncToken string, err error) {
	getNotesFilter := gosn.Filter{
		Type: "Note",
	}
	itemFilter := gosn.ItemFilters{
		Filters: []gosn.Filter{getNotesFilter},
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
		var deleteNote bool
		if item.Content != nil && item.ContentType == "Note" {
			if StringInSlice(item.UUID, noteUUIDs, true) {
				item.Deleted = true
				deleteNote = true
			} else if StringInSlice(item.Content.GetTitle(), noteTitles, true) {
				item.Deleted = true
				deleteNote = true
			}
			if deleteNote {
				notesToDelete = append(notesToDelete, item)
			}
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
