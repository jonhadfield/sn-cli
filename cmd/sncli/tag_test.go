package main

import (
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var testSession cache.Session

func TestMain(m *testing.M) {
	gs, err := gosn.CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	if err != nil {
		panic(err)
	}

	testSession.Server = gs.Server
	testSession.Mk = gs.Mk
	testSession.Ak = gs.Ak
	testSession.Token = gs.Token

	var path string

	path, err = cache.GenCacheDBPath(testSession, "", sncli.SNAppName)
	if err != nil {
		panic(err)
	}

	testSession.CacheDBPath = path

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
	tags, err = getTagsByTitle(testSession, "TestTagOne", true)
	assert.NoError(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, "TestTagOne", tags[0].Content.Title)


	tagUUID := tags[0].UUID

	var tag gosn.Tag
	tag, err = getTagByUUID(testSession, tagUUID, true)
	assert.NoError(t, err)
	assert.Equal(t, "TestTagOne", tag.Content.Title)

	tags, err = getTagsByTitle(testSession, "MissingTagOne", true)
	assert.NoError(t, err)
	assert.Empty(t, tags)

	tag, err = getTagByUUID(testSession, "123", true)
	assert.Error(t, err)
	assert.Equal(t, "could not find tag with UUID 123", err.Error())
}
