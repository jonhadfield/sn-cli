package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWipe(t *testing.T) {
	time.Sleep(250 * time.Millisecond)
	var outputBuffer bytes.Buffer
	app := appSetup()
	app.Writer = &outputBuffer

	osArgs := []string{"sncli", "wipe", "--yes"}
	err := app.Run(osArgs)
	stdout := outputBuffer.String()
	// fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgItemsDeleted)
}

// func TestAddDeleteNote(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
// 	require.NoError(t, err, "failed to wipe")
// 	err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "testNote", "--text", "some example text"})
// 	require.NoError(t, err, "failed to add note")
// 	require.Contains(t, msg, msgAddSuccess)
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
// 	require.NoError(t, err, "failed to get note count")
// 	require.Equal(t, "1", msg)
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "testNote"})
// 	require.NoError(t, err, "failed to delete note")
// 	require.Contains(t, msg, msgDeleted)
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
// 	require.NoError(t, err, "failed to get note count")
// 	require.Equal(t, "0", msg)
// 	time.Sleep(250 * time.Millisecond)
// }
//
// func TestAddNoteErrorMissingTitle(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note"})
// 	require.Error(t, err)
// }
//
// func TestDeleteNoteErrorMissingTitle(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note"})
// 	require.Error(t, err, "error should be returned if title is unspecified")
// }
//
// func TestTagNotesByTextWithNewTags(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "TestNoteOne", "--text", "test note one"})
// 	require.Contains(t, msg, msgAddSuccess)
// 	require.NoError(t, err)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "note", "--title", "TestNoteTwo", "--text", "test note two"})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgAddSuccess)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "tag", "--find-text", "test note", "--title", "testTagOne,testTagTwo"})
// 	require.NoError(t, err)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "TestNoteOne,TestNoteTwo"})
// 	require.NoError(t, err)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
// 	require.NoError(t, err)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "0", msg)
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
// 	require.NoError(t, err)
// 	require.NotEmpty(t, msg)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testTagOne,testTagTwo"})
// 	require.NoError(t, err)
// 	time.Sleep(250 * time.Millisecond)
// }
//
// func TestAddOneNoteGetCount(t *testing.T) {
// 	err := startCLI([]string{
// 		"sncli", "--debug", "--no-stdout", "add", "note", "--title", "testAddOneNoteGetCount Title",
// 		"--text", "testAddOneNoteGetCount Text",
// 	})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgAddSuccess)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note"})
// 	require.NoError(t, err)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "1", msg)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "note", "--title", "testAddOneNoteGetCount Title"})
// 	require.NoError(t, err)
// 	time.Sleep(250 * time.Millisecond)
// }
//
// func TestAddOneTagGetCount(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "add", "tag", "--title", "testAddOneTagGetCount Title"})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgAddSuccess)
// 	err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "1", msg)
//
// 	err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testAddOneTagGetCount Title"})
// 	require.NoError(t, err)
//
// 	time.Sleep(250 * time.Millisecond)
// }
//
// func TestGetNoteCountWithNoResults(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "note", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "0", msg)
// 	time.Sleep(250 * time.Millisecond)
// }
//
// func TestGetTagCountWithNoResults(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "0", msg)
// 	time.Sleep(250 * time.Millisecond)
// }
//
// func TestGetNotesWithNoResults(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "get", "note"})
// 	require.NoError(t, err)
// 	require.Equal(t, msgNoMatches, msg)
// 	time.Sleep(250 * time.Millisecond)
// }
//
// func TestGetTagsWithNoResults(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "get", "tag"})
// 	require.NoError(t, err)
// 	require.Equal(t, msgNoMatches, msg)
// 	time.Sleep(250 * time.Millisecond)
// }
//
// func TestFinalWipeAndCountZero(t *testing.T) {
// 	err := startCLI([]string{"sncli", "--debug", "wipe", "--yes"})
// 	require.NoError(t, err)
//
// 	var msg string
//
// 	err = startCLI([]string{"sncli", "--debug", "get", "note", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "0", msg)
//
// 	err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "0", msg)
// }
