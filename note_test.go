package sncli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jonhadfield/gosn"
)

func TestAddDeleteNoteByUUID(t *testing.T) {
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	if err != nil {
		t.Error(err)
	}
	// create note
	addNoteConfig := AddNoteConfig{
		Session: session,
		Title:   "TestNoteOne",
		Text:    "TestNoteOneText",
	}
	err = addNoteConfig.Run()
	assert.NoError(t, err, err)

	// get new note
	filter := gosn.Filter{
		Type:       "Note",
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}

	iFilter := gosn.ItemFilters{
		Filters: []gosn.Filter{filter},
	}
	gnc := GetNoteConfig{
		Session: session,
		Filters: iFilter,
	}
	var preRes, postRes gosn.GetItemsOutput
	preRes, err = gnc.Run()
	assert.NoError(t, err, err)

	newItemUUID := preRes.Items[0].UUID
	deleteNoteConfig := DeleteNoteConfig{
		Session:   session,
		NoteUUIDs: []string{newItemUUID},
	}
	err = deleteNoteConfig.Run()
	assert.NoError(t, err, err)

	postRes, err = gnc.Run()
	assert.NoError(t, err, err)
	assert.EqualValues(t, len(postRes.Items), 0, "note was not deleted")

}

func TestAddDeleteNoteByTitle(t *testing.T) {
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	assert.NoError(t, err, err)

	addNoteConfig := AddNoteConfig{
		Session: session,
		Title:   "TestNoteOne",
	}
	err = addNoteConfig.Run()
	assert.NoError(t, err, err)

	deleteNoteConfig := DeleteNoteConfig{
		Session:    session,
		NoteTitles: []string{"TestNoteOne"},
	}
	err = deleteNoteConfig.Run()
	assert.NoError(t, err, err)

	filter := gosn.Filter{
		Type:       "Note",
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}

	iFilter := gosn.ItemFilters{
		Filters: []gosn.Filter{filter},
	}
	gnc := GetNoteConfig{
		Session: session,
		Filters: iFilter,
	}
	var postRes gosn.GetItemsOutput
	postRes, err = gnc.Run()
	assert.NoError(t, err, err)
	assert.EqualValues(t, len(postRes.Items), 0, "note was not deleted")

}

func TestAddDeleteNoteByTitleRegex(t *testing.T) {
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	assert.NoError(t, err, err)
	// add note
	addNoteConfig := AddNoteConfig{
		Session: session,
		Title:   "TestNoteOne",
	}
	err = addNoteConfig.Run()
	assert.NoError(t, err, err)

	// delete note
	deleteNoteConfig := DeleteNoteConfig{
		Session:    session,
		NoteTitles: []string{"^T.*ote..[def]"},
		Regex:      true,
	}
	err = deleteNoteConfig.Run()
	assert.NoError(t, err, err)

	// get same note again
	filter := gosn.Filter{
		Type:       "Note",
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}
	iFilter := gosn.ItemFilters{
		Filters: []gosn.Filter{filter},
	}
	gnc := GetNoteConfig{
		Session: session,
		Filters: iFilter,
	}
	var postRes gosn.GetItemsOutput
	postRes, err = gnc.Run()

	assert.NoError(t, err, err)
	assert.EqualValues(t, len(postRes.Items), 0, "note was not deleted")

}

func TestGetNote(t *testing.T) {
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	assert.NoError(t, err)
	// create one note
	addNoteConfig := AddNoteConfig{
		Session: session,
		Title:   "TestNoteOne",
	}
	err = addNoteConfig.Run()
	assert.NoError(t, err)

	noteFilter := gosn.Filter{
		Type:       "Note",
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}
	// retrieve one note
	itemFilters := gosn.ItemFilters{
		MatchAny: false,
		Filters:  []gosn.Filter{noteFilter},
	}
	getNoteConfig := GetNoteConfig{
		Session: session,
		Filters: itemFilters,
	}
	var output gosn.GetItemsOutput
	output, err = getNoteConfig.Run()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(output.Items))

	// clean up
	deleteNoteConfig := DeleteNoteConfig{
		Session:    session,
		NoteTitles: []string{"TestNoteOne"},
	}
	assert.NoError(t, deleteNoteConfig.Run(), "clean up failed")
}

func TestWipe(t *testing.T) {
	numNotes := 50
	textParas := 10
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	assert.NoError(t, err)

	err = createNotes(session, numNotes, textParas)
	assert.NoError(t, err)

	wipeConfig := WipeConfig{
		Session: session,
	}
	var deleted int
	deleted, err = wipeConfig.Run()
	assert.NoError(t, err)
	assert.EqualValues(t, deleted, numNotes, "wipe failed")

}

func TestCreateOneHundredNotes(t *testing.T) {
	numNotes := 100
	textParas := 10
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	assert.NoError(t, err)

	err = createNotes(session, numNotes, textParas)
	assert.NoError(t, err)

	noteFilter := gosn.Filter{
		Type: "Note",
	}
	filter := gosn.ItemFilters{
		Filters: []gosn.Filter{noteFilter},
	}

	gnc := GetNoteConfig{
		Session: session,
		Filters: filter,
	}
	var res gosn.GetItemsOutput
	res, err = gnc.Run()
	assert.NoError(t, err)

	assert.EqualValues(t, len(res.Items), numNotes)
	wipeConfig := WipeConfig{
		Session: session,
	}
	var deleted int
	deleted, err = wipeConfig.Run()
	assert.NoError(t, err)
	assert.EqualValues(t, deleted, numNotes)

}
