package sncli

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	gosn "github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

var testSession *cache.Session

func TestMain(m *testing.M) {
	gs, err := gosn.CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	if err != nil {
		panic(err)
	}

	testSession = &cache.Session{
		Session: &gosn.Session{
			Debug:             true,
			Server:            gs.Server,
			Token:             gs.Token,
			MasterKey:         gs.MasterKey,
			RefreshExpiration: gs.RefreshExpiration,
			RefreshToken:      gs.RefreshToken,
			AccessToken:       gs.AccessToken,
			AccessExpiration:  gs.AccessExpiration,
		},
		CacheDBPath: "",
	}

	var path string

	path, err = cache.GenCacheDBPath(*testSession, "", SNAppName)
	if err != nil {
		panic(err)
	}

	testSession.CacheDBPath = path

	var so cache.SyncOutput
	so, err = sync(cache.SyncInput{
		Session: testSession,
		Close:   false,
	})
	if err != nil {
		panic(err)
	}

	var allPersistedItems cache.Items

	if err = so.DB.All(&allPersistedItems); err != nil {
		return
	}
	if err = so.DB.Close() ; err != nil {
		panic(err)
	}

	if testSession.DefaultItemsKey.ItemsKey == "" {
		panic("failed in TestMain due to empty default items key")
	}
	os.Exit(m.Run())
}

func TestWipeWith50(t *testing.T) {
	testDelay()

	cleanUp(*testSession)
	defer cleanUp(*testSession)
	numNotes := 50
	textParas := 3

	err := createNotes(testSession, numNotes, textParas)
	assert.NoError(t, err)

	// check notes created
	noteFilter := gosn.Filter{
		Type: "Note",
	}
	filters := gosn.ItemFilters{
		Filters: []gosn.Filter{noteFilter},
	}
	gni := cache.SyncInput{
		Session: testSession,
	}

	var gno cache.SyncOutput
	gno, err = Sync(gni, false)
	assert.NoError(t, err)
	assert.NotNil(t, gno.DB)

	// get items from db
	var items cache.Items

	assert.NoError(t, gno.DB.All(&items))
	assert.NoError(t, gno.DB.Close())

	var nonotes int
	for _, i := range items {
		if i.ContentType == "Note" {
			nonotes++
		}
	}
	var gItems gosn.Items
	gItems, err = items.ToItems(testSession)

	assert.NoError(t, err)

	gItems.DeDupe()
	ei := gItems

	ei.Filter(filters)

	assert.Equal(t, 50, len(ei))

	wipeConfig := WipeConfig{
		Session: testSession,
	}

	var deleted int
	deleted, err = wipeConfig.Run()
	assert.NoError(t, err)
	assert.True(t, deleted >= numNotes, fmt.Sprintf("notes created: %d items deleted: %d", numNotes, deleted))
}


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
		Session: testSession,
		Filters: iFilter,
	}

	var preRes, postRes gosn.Items

	preRes, err = gnc.Run()

	assert.NoError(t, err, err)

	newItemUUID := preRes[0].GetUUID()
	deleteNoteConfig := DeleteNoteConfig{
		Session:   testSession,
		NoteUUIDs: []string{newItemUUID},
	}

	var noDeleted int
	noDeleted, err = deleteNoteConfig.Run()
	assert.Equal(t, noDeleted, 1)
	assert.NoError(t, err, err)

	postRes, err = gnc.Run()
	assert.NoError(t, err, err)
	assert.EqualValues(t, len(postRes), 0, "note was not deleted")
}

func TestAddDeleteNoteByTitle(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	addNoteConfig := AddNoteInput{
		Session: testSession,
		Title:   "TestNoteOne",
	}
	err := addNoteConfig.Run()
	assert.NoError(t, err, err)

	deleteNoteConfig := DeleteNoteConfig{
		Session:    testSession,
		NoteTitles: []string{"TestNoteOne"},
	}

	var noDeleted int
	noDeleted, err = deleteNoteConfig.Run()
	assert.Equal(t, noDeleted, 1)
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
		Session: testSession,
		Filters: iFilter,
	}

	var postRes gosn.Items
	postRes, err = gnc.Run()
	assert.NoError(t, err, err)
	assert.EqualValues(t, len(postRes), 0, "note was not deleted")
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
	assert.NoError(t, err, err)

	// delete note
	deleteNoteConfig := DeleteNoteConfig{
		Session:    testSession,
		NoteTitles: []string{"^T.*ote..[def]"},
		Regex:      true,
	}

	var noDeleted int
	noDeleted, err = deleteNoteConfig.Run()
	assert.Equal(t, noDeleted, 1)
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
		Session: testSession,
		Filters: iFilter,
	}

	var postRes gosn.Items
	postRes, err = gnc.Run()

	assert.NoError(t, err, err)
	assert.EqualValues(t, len(postRes), 0, "note was not deleted")
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
		Session: testSession,
		Filters: itemFilters,
	}

	var output gosn.Items
	output, err = getNoteConfig.Run()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(output))
}

func TestCreateOneHundredNotes(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	numNotes := 100
	textParas := 10

	cleanUp(*testSession)

	err := createNotes(testSession, numNotes, textParas)
	assert.NoError(t, err)

	noteFilter := gosn.Filter{
		Type: "Note",
	}
	filter := gosn.ItemFilters{
		Filters: []gosn.Filter{noteFilter},
	}

	gnc := GetNoteConfig{
		Session: testSession,
		Filters: filter,
	}

	var res gosn.Items
	res, err = gnc.Run()
	assert.NoError(t, err)

	assert.True(t, len(res) >= numNotes)

	wipeConfig := WipeConfig{
		Session: testSession,
	}

	var deleted int
	deleted, err = wipeConfig.Run()
	assert.NoError(t, err)
	assert.True(t, deleted >= numNotes)
}

func cleanUp(session cache.Session) {
	removeDB(session.CacheDBPath)
	err := gosn.DeleteContent(&gosn.Session{
		Token:     testSession.Token,
		MasterKey: testSession.MasterKey,
		Server:    testSession.Server,
		AccessToken: testSession.AccessToken,
		AccessExpiration: testSession.AccessExpiration,
		RefreshExpiration: testSession.RefreshExpiration,
		RefreshToken: testSession.RefreshToken,
	})
	if err != nil {
		panic(err)
	}
}
