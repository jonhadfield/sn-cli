package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddTag(t *testing.T) {
	err := startCLI([]string{"sncli", "add", "tag", "--title", "testTag"})
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestAddTagErrorMissingTitle(t *testing.T) {
	err := startCLI([]string{"sncli", "add", "tag"})
	if err == nil {
		t.Errorf("error should be returned if title is unspecified")
	}
}

func TestDeleteTag(t *testing.T) {
	err := startCLI([]string{"sncli", "delete", "tag", "--title", "testTag"})
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestDeleteTagErrorMissingTitle(t *testing.T) {
	err := startCLI([]string{"sncli", "delete", "tag"})
	if err == nil {
		t.Errorf("error should be returned if title is unspecified")
	}
}

func TestAddNote(t *testing.T) {
	err := startCLI([]string{"sncli", "add", "note", "--title", "testNote", "--text", "some example text"})
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestAddNoteErrorMissingTitle(t *testing.T) {
	err := startCLI([]string{"sncli", "add", "note"})
	if err == nil {
		t.Errorf("error should be returned if title is unspecified")
	}
}

func TestDeleteNote(t *testing.T) {
	err := startCLI([]string{"sncli", "delete", "note", "--title", "testNote"})
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestDeleteNoteErrorMissingTitle(t *testing.T) {
	err := startCLI([]string{"sncli", "delete", "note"})
	if err == nil {
		t.Errorf("error should be returned if title is unspecified")
	}
}

func TestTagNotesByTextWithNewTags(t *testing.T) {
	startCLI([]string{"sncli", "get", "notes"})
	err := startCLI([]string{"sncli", "add", "note", "--title", "TestNoteOne", "--text", "test note one"})
	assert.NoError(t, err, err)
	err = startCLI([]string{"sncli", "add", "note", "--title", "TestNoteTwo", "--text", "test note two"})
	assert.NoError(t, err, err)
	err = startCLI([]string{"sncli", "tag", "--find-text", "test note", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err, err)
	err = startCLI([]string{"sncli", "delete", "note", "--title", "TestNoteOne,TestNoteTwo"})
	assert.NoError(t, err, err)
	startCLI([]string{"sncli", "get", "notes"})
	err = startCLI([]string{"sncli", "delete", "tag", "--title", "testTagOne,testTagTwo"})
	assert.NoError(t, err, err)
	startCLI([]string{"sncli", "get", "notes"})

}

//func TestAddTag1(t *testing.T) {
//	fmt.Println(randomdata.Paragraph())
//	//err := startCLI([]string{"sncli", "add", "tag", "--title", "testTag"})
//	//if err != nil {
//	//	t.Errorf("%+v", err)
//	//}
//}
