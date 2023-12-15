package main

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
	sncli "github.com/jonhadfield/sn-cli"

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

//
// func signIn(server, email, password string) {
// 	ts, err := auth.CliSignIn(email, password, server, true)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	gTtestSession = &session.Session{
// 		Debug:             true,
// 		HTTPClient:        common.NewHTTPClient(),
// 		SchemaValidation:  false,
// 		Server:            ts.Server,
// 		FilesServerUrl:    ts.FilesServerUrl,
// 		Token:             "",
// 		MasterKey:         ts.MasterKey,
// 		ItemsKeys:         nil,
// 		DefaultItemsKey:   session.SessionItemsKey{},
// 		KeyParams:         ts.KeyParams,
// 		AccessToken:       ts.AccessToken,
// 		RefreshToken:      ts.RefreshToken,
// 		AccessExpiration:  ts.AccessExpiration,
// 		RefreshExpiration: ts.RefreshExpiration,
// 		ReadOnlyAccess:    ts.ReadOnlyAccess,
// 		PasswordNonce:     "",
// 		Schemas:           nil,
// 	}
// }

func signIn(server, email, password string) {
	ts, err := auth.CliSignIn(email, password, server, true)
	if err != nil {
		log.Fatal(err)
	}

	if server == "" {
		server = session.SNServerURL
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

// func TestMain(m *testing.M) {
// 	if strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
// 		localTestMain()
// 	} else {
// 		signIn(os.Getenv("SN_SERVER"), os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"))
// 	}
//
// 	if _, err := items.Sync(items.SyncInput{Session: gTtestSession}); err != nil {
// 		log.Fatal(err)
// 	}
//
// 	if gTtestSession.DefaultItemsKey.ItemsKey == "" {
// 		panic("failed in TestMain due to empty default items key")
// 	}
//
// 	var err error
// 	testSession, err = cache.ImportSession(gTtestSession, "")
// 	if err != nil {
// 		return
// 	}
//
// 	testSession.CacheDBPath, err = cache.GenCacheDBPath(*testSession, "", gosn.LibName)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	os.Exit(m.Run())
// }

func TestMain(m *testing.M) {
	// if os.Getenv("SN_SERVER") == "" || strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
	if strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
		localTestMain()
	} else {
		signIn(session.SNServerURL, os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"))
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

func TestGetTagsByTitleAndUUID(t *testing.T) {
	addTagConfig := sncli.AddTagsInput{
		Session: testSession,
		Tags:    []string{"TestTagOne", "TestTagTwo"},
	}

	ato, err := addTagConfig.Run()
	require.NoError(t, err)
	require.Contains(t, ato.Added, "TestTagOne")
	require.Contains(t, ato.Added, "TestTagTwo")
	require.Empty(t, ato.Existing)

	var tags items.Tags
	tags, err = getTagsByTitle(*testSession, "TestTagOne")
	require.NoError(t, err)
	require.Len(t, tags, 1)
	require.Equal(t, "TestTagOne", tags[0].Content.Title)

	tagUUID := tags[0].UUID

	var tag items.Tag
	tag, err = getTagByUUID(testSession, tagUUID)
	require.NoError(t, err)
	require.Equal(t, "TestTagOne", tag.Content.Title)

	tags, err = getTagsByTitle(*testSession, "MissingTagOne")
	require.NoError(t, err)
	require.Empty(t, tags)

	_, err = getTagByUUID(testSession, "123")
	require.Error(t, err)
	require.Equal(t, "could not find tag with UUID 123", err.Error())
}
