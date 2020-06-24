package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWipe(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "wipe", "--yes"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgItemsDeleted)
}

func TestAddDeleteTag(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "wipe", "--yes"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgItemsDeleted)
	msg, _, err = startCLI([]string{"sncli", "add", "tag", "--title", "testTag"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "get", "tag", "--title", "testTag", "--count"})
	assert.Equal(t, "1", msg)
	assert.NoError(t, err)
	msg, _, err = startCLI([]string{"sncli", "delete", "tag", "--title", "testTag"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgDeleted)
}

func TestAddTagErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--no-stdout", "add", "tag"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestDeleteTagErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--no-stdout", "delete", "tag"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestAddDeleteNote(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "wipe", "--yes"})
	assert.NoError(t, err)

	msg, _, err := startCLI([]string{"sncli", "add", "note", "--title", "testNote", "--text", "some example text"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "1", msg)
	msg, _, err = startCLI([]string{"sncli", "delete", "note", "--title", "testNote"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgDeleted)
	msg, _, err = startCLI([]string{"sncli", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
}

func TestAddNoteErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--no-stdout", "add", "note"})
	assert.Error(t, err)
}

func TestDeleteNoteErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--no-stdout", "delete", "note"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestTagNotesByTextWithNewTags(t *testing.T) {
	var msg string

	var err error

	msg, _, err = startCLI([]string{"sncli", "--no-stdout", "add", "note", "--title", "TestNoteOne", "--text", "test note one"})
	assert.Contains(t, msg, msgAddSuccess)
	assert.NoError(t, err, err)
	msg, _, err = startCLI([]string{"sncli", "--no-stdout", "add", "note", "--title", "TestNoteTwo", "--text", "test note two"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgAddSuccess)

	_, _, err = startCLI([]string{"sncli", "--no-stdout", "tag", "--find-text", "test note", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err)
	_, _, err = startCLI([]string{"sncli", "--no-stdout", "delete", "note", "--title", "TestNoteOne,TestNoteTwo"})
	assert.NoError(t, err, err)

	msg, _, err = startCLI([]string{"sncli", "get", "note"})
	assert.NotEmpty(t, msg)

	msg, _, err = startCLI([]string{"sncli", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
	msg, _, err = startCLI([]string{"sncli", "get", "note"})
	assert.NoError(t, err)
	assert.NotEmpty(t, msg)

	_, _, err = startCLI([]string{"sncli", "--no-stdout", "delete", "tag", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err, err)
}

func TestAddOneNoteGetCount(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "add", "note", "--title", "testAddOneNoteGetCount Title",
		"--text", "testAddOneNoteGetCount Text"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "get", "note"})
	assert.NoError(t, err)
	msg, _, err = startCLI([]string{"sncli", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "1", msg)

	_, _, err = startCLI([]string{"sncli", "delete", "note", "--title", "testAddOneNoteGetCount Title"})
	assert.NoError(t, err, err)
}

func TestAddOneTagGetCount(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "add", "tag", "--title", "testAddOneTagGetCount Title"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "get", "tag", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "1", msg)

	_, _, err = startCLI([]string{"sncli", "delete", "tag", "--title", "testAddOneTagGetCount Title"})
	assert.NoError(t, err, err)
}

func TestGetNoteCountWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
}

func TestGetTagCountWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "get", "tag", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
}

func TestGetNotesWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "get", "note"})
	assert.NoError(t, err)
	assert.Equal(t, msgNoMatches, msg)
}

func TestGetTagsWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "get", "tag"})
	assert.NoError(t, err)
	assert.Equal(t, msgNoMatches, msg)
}

func TestFinalWipeAndCountZero(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "wipe", "--yes"})
	assert.NoError(t, err)

	var msg string

	msg, _, err = startCLI([]string{"sncli", "get", "note", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)

	msg, _, err = startCLI([]string{"sncli", "get", "tag", "--count"})
	assert.NoError(t, err)
	assert.Equal(t, "0", msg)
}
