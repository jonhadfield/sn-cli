package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/stretchr/testify/assert"
)

var (
	testSession      *cache.Session
	gTtestSession    *gosn.Session
	testUserEmail    string
	testUserPassword string
)

func localTestMain() {
	localServer := "http://ramea:3000"
	testUserEmail = fmt.Sprintf("ramea-%s", strconv.FormatInt(time.Now().UnixNano(), 16))
	testUserPassword = "secretsanta"

	rInput := gosn.RegisterInput{
		Password:   testUserPassword,
		Email:      testUserEmail,
		Identifier: testUserEmail,
		APIServer:  localServer,
		Version:    "004",
		Debug:      true,
	}

	_, err := rInput.Register()
	if err != nil {
		panic(fmt.Sprintf("failed to register with: %s", localServer))
	}

	signIn(localServer, testUserEmail, testUserPassword)
}

func signIn(server, email, password string) {
	ts, err := gosn.CliSignIn(email, password, server, true)
	if err != nil {
		log.Fatal(err)
	}

	gTtestSession = &ts
}

func TestMain(m *testing.M) {
	if os.Getenv("SN_SERVER") == "" || strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
		localTestMain()
	} else {
		signIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	}

	if _, err := gosn.Sync(gosn.SyncInput{Session: gTtestSession}); err != nil {
		log.Fatal(err)
	}

	if gTtestSession.DefaultItemsKey.ItemsKey == "" {
		panic("failed in TestMain due to empty default items key")
	}

	var err error
	testSession, err = cache.ImportSession(gTtestSession, "")
	if err != nil {
		return
	}

	testSession.CacheDBPath, err = cache.GenCacheDBPath(*testSession, "", gosn.LibName)
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
	assert.NoError(t, err)
	assert.Contains(t, ato.Added, "TestTagOne")
	assert.Contains(t, ato.Added, "TestTagTwo")
	assert.Empty(t, ato.Existing)

	var tags gosn.Tags
	tags, err = getTagsByTitle(*testSession, "TestTagOne")
	assert.NoError(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, "TestTagOne", tags[0].Content.Title)

	tagUUID := tags[0].UUID

	var tag gosn.Tag
	tag, err = getTagByUUID(testSession, tagUUID)
	assert.NoError(t, err)
	assert.Equal(t, "TestTagOne", tag.Content.Title)

	tags, err = getTagsByTitle(*testSession, "MissingTagOne")
	assert.NoError(t, err)
	assert.Empty(t, tags)

	_, err = getTagByUUID(testSession, "123")
	assert.Error(t, err)
	assert.Equal(t, "could not find tag with UUID 123", err.Error())
}
