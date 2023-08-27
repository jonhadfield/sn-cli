package main

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWipe(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	require.NoError(t, err)
	require.Contains(t, msg, msgItemsDeleted)
	time.Sleep(1 * time.Second)
}

func TestAddTag(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	require.NoError(t, err)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testAddOneTagGetCount"})
	require.NoError(t, err)
	require.Contains(t, msg, msgAddSuccess)
}

func TestAddGetTag(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	require.NoError(t, err)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testAddOneTagGetCount"})
	require.NoError(t, err)
	require.Contains(t, msg, msgAddSuccess)
	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "tag", "--title", "testAddOneTagGetCount"})
	require.NoError(t, err)
}

// func TestAddGetTagExport(t *testing.T) {
// 	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
// 	require.NoError(t, err)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testAddOneTagGetCount"})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgAddSuccess)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "tag", "--title", "testAddOneTagGetCount"})
// 	require.NoError(t, err)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "export"})
// 	require.NoError(t, err)
// }

func TestAddDeleteTag(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	require.NoError(t, err, "'wipe --yes' failed")
	require.Contains(t, msg, msgItemsDeleted)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testTag"})
	require.NoError(t, err, "'add tag --title testTag' failed")
	require.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--title", "testTag", "--count"})
	require.Equal(t, "1", msg)
	require.NoError(t, err)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testTag"})
	require.NoError(t, err)
	require.Contains(t, msg, msgDeleted)
}

// func TestAddTagExportDeleteTagReImport(t *testing.T) {
// 	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
// 	require.NoError(t, err)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testAddOneTagGetCount"})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgAddSuccess)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "1", msg)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "export"})
// 	require.NoError(t, err)
// 	require.True(t, strings.HasPrefix(msg, "encrypted export written to:"))
// 	path := strings.TrimPrefix(msg, "encrypted export written to:")
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testAddOneTagGetCount"})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgDeleted)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "0", msg)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "import", "--experiment", "--file", path})
// 	require.NoError(t, err)
// 	require.True(t, strings.HasPrefix(msg, "imported"))
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "1", msg)
// }

func TestAddTagErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag"})
	require.Error(t, err, "error should be returned if title is unspecified")
}

func TestDeleteTagErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag"})
	require.Error(t, err, "error should be returned if title is unspecified")
}

func TestAddDeleteNote(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
	require.NoError(t, err, "failed to wipe")
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "testNote", "--text", "some example text"})
	require.NoError(t, err, "failed to add note")
	require.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	require.NoError(t, err, "failed to get note count")
	require.Equal(t, "1", msg)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "testNote"})
	require.NoError(t, err, "failed to delete note")
	require.Contains(t, msg, msgDeleted)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	require.NoError(t, err, "failed to get note count")
	require.Equal(t, "0", msg)
	time.Sleep(1 * time.Second)
}

func TestAddNoteErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note"})
	require.Error(t, err)
}

func TestDeleteNoteErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note"})
	require.Error(t, err, "error should be returned if title is unspecified")
}

func TestTagNotesByTextWithNewTags(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "TestNoteOne", "--text", "test note one"})
	require.Contains(t, msg, msgAddSuccess)
	require.NoError(t, err)

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "TestNoteTwo", "--text", "test note two"})
	require.NoError(t, err)
	require.Contains(t, msg, msgAddSuccess)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "tag", "--find-text", "test note", "--title", "testTagOne,testTagTwo"})
	require.NoError(t, err)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "TestNoteOne,TestNoteTwo"})
	require.NoError(t, err)

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
	require.NoError(t, err)

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	require.NoError(t, err)
	require.Equal(t, "0", msg)
	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
	require.NoError(t, err)
	require.NotEmpty(t, msg)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testTagOne,testTagTwo"})
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
}

func TestAddOneNoteGetCount(t *testing.T) {
	msg, _, err := startCLI([]string{
		"sncli", "--debug", "--no-stdout", "add", "note", "--title", "testAddOneNoteGetCount Title",
		"--text", "testAddOneNoteGetCount Text",
	})
	require.NoError(t, err)
	require.Contains(t, msg, msgAddSuccess)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
	require.NoError(t, err)

	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	require.NoError(t, err)
	require.Equal(t, "1", msg)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "testAddOneNoteGetCount Title"})
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
}

func TestAddOneTagGetCount(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "add", "tag", "--title", "testAddOneTagGetCount Title"})
	require.NoError(t, err)
	require.Contains(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
	require.NoError(t, err)
	require.Equal(t, "1", msg)

	_, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testAddOneTagGetCount Title"})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
}

func TestGetNoteCountWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
	require.NoError(t, err)
	require.Equal(t, "0", msg)
	time.Sleep(1 * time.Second)
}

func TestGetTagCountWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
	require.NoError(t, err)
	require.Equal(t, "0", msg)
	time.Sleep(1 * time.Second)
}

func TestGetNotesWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "get", "note"})
	require.NoError(t, err)
	require.Equal(t, msgNoMatches, msg)
	time.Sleep(1 * time.Second)
}

func TestGetTagsWithNoResults(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "--debug", "get", "tag"})
	require.NoError(t, err)
	require.Equal(t, msgNoMatches, msg)
	time.Sleep(1 * time.Second)
}

func TestFinalWipeAndCountZero(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "--debug", "wipe", "--yes"})
	require.NoError(t, err)

	var msg string

	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "note", "--count"})
	require.NoError(t, err)
	require.Equal(t, "0", msg)

	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
	require.NoError(t, err)
	require.Equal(t, "0", msg)
}
