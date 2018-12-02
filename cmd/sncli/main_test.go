package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWipe(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "wipe", "--yes"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgItemsDeleted)
}

func TestAddDeleteTag(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "add", "tag", "--title", "testTag"})
	assert.NoError(t, err)
	assert.Equal(t, msg, msgAddSuccess)
	msg, _, err = startCLI([]string{"sncli", "get", "tag", "--title", "testTag", "--count"})
	assert.Equal(t, msg, "1")
	assert.NoError(t, err)
	//msg, _, err = startCLI([]string{"sncli", "delete", "tag", "--title", "testTag"})
	//assert.NoError(t, err)
	//assert.Equal(t, fmt.Sprintf("1 %s", msgDeleted), msg)
}

func TestAddTagErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "add", "tag", "--no-stdout"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestDeleteTagMissingUUID(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "delete", "tag", "--uuid", "3a277f8d-f247-4236-a803-80795123135g"})
	assert.NoError(t, err)
	assert.Equal(t, msg, fmt.Sprintf("0 %s", msgDeleted))
}

func TestDeleteTagErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "delete", "tag", "--no-stdout"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestAddDeleteNote(t *testing.T) {
	msg, _, err := startCLI([]string{"sncli", "add", "note", "--title", "testNote", "--text", "some example text"})
	assert.NoError(t, err)
	assert.Equal(t, msg, msgAddSuccess)
	//msg, _, err = startCLI([]string{"sncli", "delete", "note", "--title", "testNote"})
	//assert.NoError(t, err)
	//assert.Equal(t, msg, fmt.Sprintf("1 %s", msgDeleted))
}

func TestAddNoteErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "add", "note", "--no-stdout"})
	assert.Error(t, err)
}

func TestDeleteNoteErrorMissingTitle(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "delete", "note", "--no-stdout"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestTagNotesByTextWithNewTags(t *testing.T) {
	_, _, err := startCLI([]string{"sncli", "get", "notes"})
	assert.NoError(t, err)
	msg, _, err := startCLI([]string{"sncli", "add", "note", "--title", "TestNoteOne", "--text", "test note one"})
	assert.Equal(t, msg, msgAddSuccess)
	assert.NoError(t, err, err)
	msg, _, err = startCLI([]string{"sncli", "add", "note", "--title", "TestNoteTwo", "--text", "test note two"})
	assert.Equal(t, msg, msgAddSuccess)
	assert.NoError(t, err)
	msg, _, err = startCLI([]string{"sncli", "tag", "--find-text", "test note", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err)
	msg, _, err = startCLI([]string{"sncli", "delete", "note", "--title", "TestNoteOne,TestNoteTwo"})
	assert.NoError(t, err, err)
	msg, _, err = startCLI([]string{"sncli", "delete", "tag", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err, err)
	_, _, err = startCLI([]string{"sncli", "get", "notes"})

}

//func TestAddTag1(t *testing.T) {
//	fmt.Println(randomdata.Paragraph())
//	//err := startCLI([]string{"sncli", "add", "tag", "--title", "testTag"})
//	//if err != nil {
//	//	t.Errorf("%+v", err)
//	//}
//}
