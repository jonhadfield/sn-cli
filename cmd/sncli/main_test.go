package main

import (
	"testing"
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
		t.Errorf("%+v", err)
	}
}

func TestDeleteTag(t *testing.T) {
	err := startCLI([]string{"sncli", "delete", "tag", "--title", "testTag"})
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestAddNote(t *testing.T) {
	err := startCLI([]string{"sncli", "add", "note", "--title", "testNote", "--text", "some example text"})
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestDeleteNote(t *testing.T) {
	err := startCLI([]string{"sncli", "delete", "note", "--title", "testNote"})
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestTagNotesByTextWithNewTags(t *testing.T) {
	err := startCLI([]string{"sncli", "add", "note", "--title", "TestNoteOne", "--text", "test note one"})
	if err != nil {
		t.Errorf("%+v", err)
	}
	err = startCLI([]string{"sncli", "add", "note", "--title", "TestNoteTwo", "--text", "test note two"})
	if err != nil {
		t.Errorf("%+v", err)
	}
	err = startCLI([]string{"sncli", "tag", "--find-text", "test note", "--title", "testTagOne,testTagTwo"})
	if err != nil {
		t.Errorf("%+v", err)
	}
	// clean up
	err = startCLI([]string{"sncli", "delete", "note", "--title", "TestNoteOne,TestNoteTwo"})
	if err != nil {
		t.Errorf("%+v", err)
	}
	err = startCLI([]string{"sncli", "delete", "tag", "--title", "testTagOne,testTagTwo"})
	if err != nil {
		t.Errorf("%+v", err)
	}
}

//func TestAddTag1(t *testing.T) {
//	fmt.Println(randomdata.Paragraph())
//	//err := startCLI([]string{"sncli", "add", "tag", "--title", "testTag"})
//	//if err != nil {
//	//	t.Errorf("%+v", err)
//	//}
//}
