package sncli

import (
	"fmt"
	"testing"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/stretchr/testify/require"
)

func TestAddDeleteNoteByUUID(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// create note
	addNoteConfig := AddNoteInput{
		Session: testSession,
		Title:   "TestNoteOne",
		Text:    "TestNoteOneText",
	}

	err := addNoteConfig.Run()
	require.NoError(t, err)

	// get new note
	filter := items.Filter{
		Type:       common.SNItemTypeNote,
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}

	iFilter := items.ItemFilters{
		Filters: []items.Filter{filter},
	}
	gnc := GetNoteConfig{
		Session: testSession,
		Filters: iFilter,
	}

	var preRes, postRes items.Items

	preRes, err = gnc.Run()

	require.NoError(t, err)

	newItemUUID := preRes[0].GetUUID()
	deleteNoteConfig := DeleteNoteConfig{
		Session:   testSession,
		NoteUUIDs: []string{newItemUUID},
	}

	var noDeleted int
	noDeleted, err = deleteNoteConfig.Run()
	require.Equal(t, 1, noDeleted)
	require.NoError(t, err)

	postRes, err = gnc.Run()
	require.NoError(t, err)
	require.EqualValues(t, len(postRes), 0, "note was not deleted")
}

func TestReplaceNote(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// create note
	addNoteConfig := AddNoteInput{
		Session: testSession,
		Title:   "TestNoteOne",
		Text:    "TestNoteOneText",
	}

	require.NoError(t, addNoteConfig.Run())

	// get new note
	gnc := GetNoteConfig{
		Session: testSession,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters: []items.Filter{{
				Type:       common.SNItemTypeNote,
				Key:        "Title",
				Comparison: "==",
				Value:      "TestNoteOne",
			}, {
				Type:       common.SNItemTypeNote,
				Key:        "Text",
				Comparison: "==",
				Value:      "TestNoteOneText",
			}},
		},
	}

	var preReplace, postReplace items.Items

	preReplace, err := gnc.Run()
	require.NoError(t, err)
	require.Len(t, preReplace, 1)
	// require.NoError(t, testSession.CacheDB.Close())

	// replace note
	replaceNoteConfig := AddNoteInput{
		Session: testSession,
		Title:   "TestNoteOne",
		Text:    "TestNoteOneReplacementText",
		Replace: true,
	}
	require.NoError(t, replaceNoteConfig.Run())

	// get updated note
	gnc = GetNoteConfig{
		Session: testSession,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters: []items.Filter{{
				Type:       common.SNItemTypeNote,
				Key:        "Title",
				Comparison: "==",
				Value:      "TestNoteOne",
			}, {
				Type:       common.SNItemTypeNote,
				Key:        "Text",
				Comparison: "==",
				Value:      "TestNoteOneReplacementText",
			}},
		},
	}

	postReplace, err = gnc.Run()
	require.NoError(t, err, err)
	require.Len(t, postReplace, 1)

	// check only one note with that title exists
	gnc = GetNoteConfig{
		Session: testSession,
		Filters: items.ItemFilters{
			MatchAny: false,
			Filters: []items.Filter{{
				Type:       common.SNItemTypeNote,
				Key:        "Title",
				Comparison: "==",
				Value:      "TestNoteOne",
			}},
		},
	}

	highlander, err := gnc.Run()
	require.NoError(t, err)
	require.Len(t, highlander, 1)
}

func TestWipeWith50(t *testing.T) {
	testDelay()

	// initial cleanup before first test
	cleanUp(*testSession)
	defer cleanUp(*testSession)

	numNotes := 50
	textParas := 3

	err := createNotes(testSession, numNotes, textParas)
	require.NoError(t, err)

	// check notes created
	noteFilter := items.Filter{
		Type: common.SNItemTypeNote,
	}
	filters := items.ItemFilters{
		Filters: []items.Filter{noteFilter},
	}
	gni := cache.SyncInput{
		Session: testSession,
	}

	var gno cache.SyncOutput
	gno, err = Sync(gni, false)

	require.NoError(t, err)
	require.NotNil(t, gno.DB)

	// get items from db
	var cItems cache.Items

	require.NoError(t, gno.DB.All(&cItems))
	require.NoError(t, gno.DB.Close())

	var nonotes int

	for _, i := range cItems {
		if i.ContentType == common.SNItemTypeNote {
			nonotes++
		}
	}

	var gItems items.Items
	gItems, err = cItems.ToItems(testSession)

	require.NoError(t, err)

	gItems.DeDupe()
	ei := gItems

	ei.Filter(filters)

	require.Equal(t, 50, len(ei))

	wipeConfig := WipeConfig{
		Session: testSession,
		Debug:   false,
	}

	var deleted int

	deleted, err = wipeConfig.Run()
	require.NoError(t, err)

	require.GreaterOrEqual(t, deleted, numNotes, fmt.Sprintf("notes created: %d items deleted: %d", numNotes, deleted))
}

func TestAddDeleteNoteByTitle(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	addNoteConfig := AddNoteInput{
		Session: testSession,
		Title:   "TestNoteOne",
	}
	err := addNoteConfig.Run()
	require.NoError(t, err)

	deleteNoteConfig := DeleteNoteConfig{
		Session:    testSession,
		NoteTitles: []string{"TestNoteOne"},
	}

	var noDeleted int
	noDeleted, err = deleteNoteConfig.Run()
	require.Equal(t, 1, noDeleted)
	require.NoError(t, err)

	filter := items.Filter{
		Type:       common.SNItemTypeNote,
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}

	iFilter := items.ItemFilters{
		Filters: []items.Filter{filter},
	}
	gnc := GetNoteConfig{
		Session: testSession,
		Filters: iFilter,
	}

	var postRes items.Items
	postRes, err = gnc.Run()
	require.NoError(t, err)
	require.EqualValues(t, len(postRes), 0, "note was not deleted")
}

func TestAddDeleteNoteByTitleRegex(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)
	// add note
	addNoteConfig := AddNoteInput{
		Session: testSession,
		Title:   "TestNoteOne",
	}
	err := addNoteConfig.Run()
	require.NoError(t, err)

	// delete note
	deleteNoteConfig := DeleteNoteConfig{
		Session:    testSession,
		NoteTitles: []string{"^T.*ote..[def]"},
		Regex:      true,
	}

	var noDeleted int
	noDeleted, err = deleteNoteConfig.Run()
	require.Equal(t, 1, noDeleted)
	require.NoError(t, err)

	// get same note again
	filter := items.Filter{
		Type:       common.SNItemTypeNote,
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}
	iFilter := items.ItemFilters{
		Filters: []items.Filter{filter},
	}
	gnc := GetNoteConfig{
		Session: testSession,
		Filters: iFilter,
	}

	var postRes items.Items
	postRes, err = gnc.Run()

	require.NoError(t, err)
	require.EqualValues(t, len(postRes), 0, "note was not deleted")
}

func TestGetNote(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// create one note
	addNoteConfig := AddNoteInput{
		Session: testSession,
		Title:   "TestNoteOne",
	}
	err := addNoteConfig.Run()
	require.NoError(t, err)

	noteFilter := items.Filter{
		Type:       common.SNItemTypeNote,
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}
	// retrieve one note
	itemFilters := items.ItemFilters{
		MatchAny: false,
		Filters:  []items.Filter{noteFilter},
	}
	getNoteConfig := GetNoteConfig{
		Session: testSession,
		Filters: itemFilters,
	}

	var output items.Items
	output, err = getNoteConfig.Run()
	require.NoError(t, err)
	require.EqualValues(t, 1, len(output))
}

func TestCreateOneHundredNotes(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	numNotes := 100
	textParas := 10

	err := createNotes(testSession, numNotes, textParas)
	require.NoError(t, err)

	noteFilter := items.Filter{
		Type: common.SNItemTypeNote,
	}
	filter := items.ItemFilters{
		Filters: []items.Filter{noteFilter},
	}

	gnc := GetNoteConfig{
		Session: testSession,
		Filters: filter,
	}

	var res items.Items
	res, err = gnc.Run()
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(res), numNotes)

	wipeConfig := WipeConfig{
		Session: testSession,
	}

	var deleted int
	deleted, err = wipeConfig.Run()
	require.NoError(t, err)
	require.GreaterOrEqual(t, deleted, numNotes)
}

func cleanUp(session cache.Session) {
	session.RemoveDB()
	_, err := items.DeleteContent(session.Session, true)
	if err != nil {
		panic(err)
	}
}
