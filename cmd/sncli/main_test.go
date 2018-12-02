package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWipe(t *testing.T) {
	msg, err := startCLI([]string{"sncli", "wipe", "--yes"})
	assert.NoError(t, err)
	assert.Contains(t, msg, msgItemsDeleted)
}

func TestAddTag(t *testing.T) {
	msg, err := startCLI([]string{"sncli", "add", "tag", "--title", "testTag"})
	assert.NoError(t, err)
	assert.Equal(t, msg, msgAddSuccess)
}

func TestAddTagErrorMissingTitle(t *testing.T) {
	_, err := startCLI([]string{"sncli", "add", "tag"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestDeleteTag(t *testing.T) {
	msg, err := startCLI([]string{"sncli", "delete", "tag", "--title", "testTag"})
	assert.NoError(t, err)
	assert.Equal(t, msg, fmt.Sprintf("1 %s", msgDeleted))
}

func TestDeleteTagMissingUUID(t *testing.T) {
	msg, err := startCLI([]string{"sncli", "delete", "tag", "--uuid", "3a277f8d-f247-4236-a803-80795123135g"})
	assert.NoError(t, err)
	assert.Equal(t, msg, fmt.Sprintf("0 %s", msgDeleted))
}

func TestDeleteTagErrorMissingTitle(t *testing.T) {
	_, err := startCLI([]string{"sncli", "delete", "tag"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestAddNote(t *testing.T) {
	msg, err := startCLI([]string{"sncli", "add", "note", "--title", "testNote", "--text", "some example text"})
	assert.NoError(t, err)
	assert.Equal(t, msg, msgAddSuccess)
}

func TestAddNoteErrorMissingTitle(t *testing.T) {
	_, err := startCLI([]string{"sncli", "add", "note"})
	assert.Error(t, err)
}

func TestDeleteNote(t *testing.T) {
	msg, err := startCLI([]string{"sncli", "delete", "note", "--title", "testNote"})
	assert.NoError(t, err)
	assert.Equal(t, msg, fmt.Sprintf("1 %s", msgDeleted))
}

func TestDeleteNoteErrorMissingTitle(t *testing.T) {
	_, err := startCLI([]string{"sncli", "delete", "note"})
	assert.Error(t, err, "error should be returned if title is unspecified")
}

func TestTagNotesByTextWithNewTags(t *testing.T) {
	_, err := startCLI([]string{"sncli", "get", "notes"})
	assert.NoError(t, err)
	msg, err := startCLI([]string{"sncli", "add", "note", "--title", "TestNoteOne", "--text", "test note one"})
	assert.Equal(t, msg, msgAddSuccess)
	assert.NoError(t, err, err)
	msg, err = startCLI([]string{"sncli", "add", "note", "--title", "TestNoteTwo", "--text", "test note two"})
	assert.Equal(t, msg, msgAddSuccess)
	assert.NoError(t, err)
	msg, err = startCLI([]string{"sncli", "tag", "--find-text", "test note", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err)
	msg, err = startCLI([]string{"sncli", "delete", "note", "--title", "TestNoteOne,TestNoteTwo"})
	assert.NoError(t, err, err)
	_, err = startCLI([]string{"sncli", "get", "notes"})
	msg, err = startCLI([]string{"sncli", "delete", "tag", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err, err)
	_, err = startCLI([]string{"sncli", "get", "notes"})

}

//func TestAddTag1(t *testing.T) {
//	fmt.Println(randomdata.Paragraph())
//	//err := startCLI([]string{"sncli", "add", "tag", "--title", "testTag"})
//	//if err != nil {
//	//	t.Errorf("%+v", err)
//	//}
//}
