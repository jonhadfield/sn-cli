package main

import (
	"os"
	"testing"

	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/stretchr/testify/assert"
)

var testSession *cache.Session

func sync(si cache.SyncInput) (so cache.SyncOutput, err error) {
	return cache.Sync(cache.SyncInput{
		Session: si.Session,
		Close:   si.Close,
	})
}

func TestMain(m *testing.M) {
	gs, err := gosn.CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"), true)
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

	path, err = cache.GenCacheDBPath(*testSession, "", sncli.SNAppName)
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
	so.DB.Close()

	if testSession.DefaultItemsKey.ItemsKey == "" {
		panic("failed in TestMain due to empty default items key")
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
