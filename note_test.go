package sncli

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jonhadfield/gosn-v2/auth"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
	"github.com/stretchr/testify/require"
)

var (
	testSession      *cache.Session
	gTtestSession    *session.Session
	testUserEmail    string
	testUserPassword string
)

func localTestMain() {
	localServer := "http://ramea:3000"
	testUserEmail = fmt.Sprintf("ramea-%s", strconv.FormatInt(time.Now().UnixNano(), 16))
	testUserPassword = "secretsanta"

	rInput := auth.RegisterInput{
		Password:  testUserPassword,
		Email:     testUserEmail,
		APIServer: localServer,
		Version:   "004",
		Debug:     true,
	}

	_, err := rInput.Register()
	if err != nil {
		panic(fmt.Sprintf("failed to register with: %s", localServer))
	}

	signIn(localServer, testUserEmail, testUserPassword)
}

func signIn(server, email, password string) {
	ts, err := auth.CliSignIn(email, password, server, true)
	if err != nil {
		log.Fatal(err)
	}

	if server == "" {
		server = SNServerURL
	}

	gTtestSession = &session.Session{
		Debug:             true,
		HTTPClient:        common.NewHTTPClient(),
		SchemaValidation:  false,
		Server:            server,
		FilesServerUrl:    ts.FilesServerUrl,
		Token:             "",
		MasterKey:         ts.MasterKey,
		ItemsKeys:         nil,
		DefaultItemsKey:   session.SessionItemsKey{},
		KeyParams:         auth.KeyParams{},
		AccessToken:       ts.AccessToken,
		RefreshToken:      ts.RefreshToken,
		AccessExpiration:  ts.AccessExpiration,
		RefreshExpiration: ts.RefreshExpiration,
		ReadOnlyAccess:    ts.ReadOnlyAccess,
		PasswordNonce:     ts.PasswordNonce,
		Schemas:           nil,
	}

	testSession = &cache.Session{
		Session:     gTtestSession,
		CacheDB:     nil,
		CacheDBPath: "",
	}
}

func TestMain(m *testing.M) {
	// if os.Getenv("SN_SERVER") == "" || strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
	if strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
		localTestMain()
	} else {
		signIn(SNServerURL, os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"))
	}

	if _, err := items.Sync(items.SyncInput{Session: gTtestSession}); err != nil {
		log.Fatal(err)
	}

	if gTtestSession.DefaultItemsKey.ItemsKey == "" {
		panic("failed in TestMain due to empty default items key")
	}
	if strings.TrimSpace(gTtestSession.Server) == "" {
		panic("failed in TestMain due to empty server")
	}

	var err error
	testSession, err = cache.ImportSession(&auth.SignInResponseDataSession{
		Debug:             gTtestSession.Debug,
		HTTPClient:        gTtestSession.HTTPClient,
		SchemaValidation:  false,
		Server:            gTtestSession.Server,
		FilesServerUrl:    gTtestSession.FilesServerUrl,
		Token:             "",
		MasterKey:         gTtestSession.MasterKey,
		KeyParams:         gTtestSession.KeyParams,
		AccessToken:       gTtestSession.AccessToken,
		RefreshToken:      gTtestSession.RefreshToken,
		AccessExpiration:  gTtestSession.AccessExpiration,
		RefreshExpiration: gTtestSession.RefreshExpiration,
		ReadOnlyAccess:    gTtestSession.ReadOnlyAccess,
		PasswordNonce:     gTtestSession.PasswordNonce,
	}, "")
	if err != nil {
		return
	}

	testSession.CacheDBPath, err = cache.GenCacheDBPath(*testSession, "", common.LibName)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
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
	require.NoError(t, err)

	// get new note
	filter := items.Filter{
		Type:       "Note",
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
				Type:       "Note",
				Key:        "Title",
				Comparison: "==",
				Value:      "TestNoteOne",
			}, {
				Type:       "Note",
				Key:        "Text",
				Comparison: "==",
				Value:      "TestNoteOneText",
			}},
		},
	}

	var preReplace, postReplace items.Items

	preReplace, err := gnc.Run()
	require.NoError(t, err, err)
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
				Type:       "Note",
				Key:        "Title",
				Comparison: "==",
				Value:      "TestNoteOne",
			}, {
				Type:       "Note",
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
				Type:       "Note",
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
		Type: "Note",
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
		if i.ContentType == "Note" {
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
		Debug:   true,
	}

	var deleted int
	deleted, err = wipeConfig.Run()
	require.NoError(t, err)
	require.True(t, deleted >= numNotes, fmt.Sprintf("notes created: %d items deleted: %d", numNotes, deleted))
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
		Type:       "Note",
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
	require.Equal(t, noDeleted, 1)
	require.NoError(t, err)

	// get same note again
	filter := items.Filter{
		Type:       "Note",
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
		Type:       "Note",
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
		Type: "Note",
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
