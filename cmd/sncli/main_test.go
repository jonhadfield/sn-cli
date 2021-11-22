package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWipe(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgItemsDeleted)
	time.Sleep(1 * time.Second)
}

func TestAddDeleteTag(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	assert.NoError(t, err, "'wipe --yes' failed")
	assert.Contains(t, msg, msgItemsDeleted)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testTag"})
	assert.NoError(t, err, "'add tag --title testTag' failed")
	assert.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "tag", "--title", "testTag", "--count"})
	assert.Equal(t, "1", msg)
	assert.NoError(t, err)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testTag"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgDeleted)
	time.Sleep(1 * time.Second)

}

func TestAddTagErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestDeleteTagErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestAddDeleteNote(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	assert.NoError(t, err, "failed to wipe")
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "testNote", "--text", "some example text"})
	assert.NoError(t, err, "failed to add note")
	assert.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	assert.NoError(t, err, "failed to get note count")
	assert.Equal(t, "1", msg)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "testNote"})
	assert.NoError(t, err, "failed to delete note")
	assert.Contains(t, msg, msgDeleted)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	assert.NoError(t, err, "failed to get note count")
	assert.Equal(t, "0", msg)
	time.Sleep(1 * time.Second)

}

func TestAddNoteErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note"})
	assert.Error(t, err)
}

func TestDeleteNoteErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestTagNotesByTextWithNewTags(t *testing.T) {
	var msg string

	var err error

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "TestNoteOne", "--text", "test note one"})
	assert.Contains(t, msg, msgAddSuccess)
	assert.NoError(t, err, err)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "TestNoteTwo", "--text", "test note two"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgAddSuccess)
	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "tag", "--find-text", "test note", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err)
	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "TestNoteOne,TestNoteTwo"})
	assert.NoError(t, err)

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
	assert.NoError(t, err)

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
	assert.NoError(t, err)
	assert.NotEmpty(t, msg)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err, err)
	time.Sleep(1 * time.Second)

}

func TestAddOneNoteGetCount(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "testAddOneNoteGetCount Title",
		"--text", "testAddOneNoteGetCount Text"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgAddSuccess)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
	assert.NoError(t, err)

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "1", msg)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "testAddOneNoteGetCount Title"})
	assert.NoError(t, err, err)
	time.Sleep(1 * time.Second)
}

func TestAddOneTagGetCount(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testAddOneTagGetCount Title"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "tag", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "1", msg)

	_, _, err = startCLI([]string{"sncli", "--no-stdout", "delete", "tag", "--title", "testAddOneTagGetCount Title"})
	assert.NoError(t, err, err)

	time.Sleep(1 * time.Second)
}

func TestGetNoteCountWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
	time.Sleep(1 * time.Second)

}

func TestGetTagCountWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "tag", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
	time.Sleep(1 * time.Second)
}

func TestGetNotesWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
	assert.NoError(t, err)
	assert.Equal(t, msgNoMatches, msg)
	time.Sleep(1 * time.Second)
}

func TestGetTagsWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "tag"})
	assert.NoError(t, err)
	assert.Equal(t, msgNoMatches, msg)
	time.Sleep(1 * time.Second)

}

func TestFinalWipeAndCountZero(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	assert.NoError(t, err)

	var msg string

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "tag", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
}
