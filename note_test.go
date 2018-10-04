package sncli

import (
	"fmt"
	"os"
	"testing"

	"github.com/jonhadfield/gosn"
)

func TestAddDeleteNoteByUUID(t *testing.T) {
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	if err != nil {
		t.Error(err)
	}
	addNoteConfig := AddNoteConfig{
		Session: session,
		Title:   "TestNoteOne",
	}
	err = addNoteConfig.Run()
	if err != nil {
		t.Errorf("error: %+v", err)
	}

	filter := gosn.Filter{
		Type:       "Note",
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}

	iFilter := gosn.ItemFilters{
		Filters: []gosn.Filter{filter},
	}
	gnc := GetNoteConfig{
		Session: session,
		Filters: iFilter,
	}
	var preRes, postRes gosn.GetItemsOutput
	preRes, err = gnc.Run()
	if err != nil {
		t.Errorf("error: failed to retrieve note: %+v\n", err)
	}

	newItemUUID := preRes.Items[0].UUID
	deleteNoteConfig := DeleteNoteConfig{
		Session:   session,
		NoteUUIDs: []string{newItemUUID},
	}
	err = deleteNoteConfig.Run()
	if err != nil {
		t.Errorf("error: %+v", err)
	}

	postRes, err = gnc.Run()
	if err != nil {
		t.Errorf("error: %+v", err)
	}
	if len(postRes.Items) != 0 {
		t.Error("note was not deleted")
	}
}

func TestAddDeleteNoteByTitle(t *testing.T) {
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	if err != nil {
		t.Error(err)
	}
	addNoteConfig := AddNoteConfig{
		Session: session,
		Title:   "TestNoteOne",
	}
	err = addNoteConfig.Run()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
	deleteNoteConfig := DeleteNoteConfig{
		Session:    session,
		NoteTitles: []string{"TestNoteOne"},
	}
	err = deleteNoteConfig.Run()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestGetNote(t *testing.T) {
	//session, _, signInErr := gosn.CliSignIn(os.Getenv("SN_EMAIL"), "", "", os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	//if signInErr != nil {
	//	t.Errorf("CliSignIn error:: %+v", signInErr)
	//}
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// create one note
	addNoteConfig := AddNoteConfig{
		Session: session,
		Title:   "TestNoteOne",
	}
	err = addNoteConfig.Run()
	if err != nil {
		t.Errorf("%+v", err)
	}

	noteFilter := gosn.Filter{
		Type:       "Note",
		Key:        "Title",
		Comparison: "==",
		Value:      "TestNoteOne",
	}
	// retrieve one note
	itemFilters := gosn.ItemFilters{
		MatchAny: false,
		Filters:  []gosn.Filter{noteFilter},
	}
	getNoteConfig := GetNoteConfig{
		Session: session,
		Filters: itemFilters,
	}
	var output gosn.GetItemsOutput
	output, err = getNoteConfig.Run()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
	if len(output.Items) != 1 {
		t.Errorf("expected one item but got: %+v", output.Items)
	}

	// clean up
	deleteNoteConfig := DeleteNoteConfig{
		Session:    session,
		NoteTitles: []string{"TestNoteOne"},
	}
	err = deleteNoteConfig.Run()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestWipe(t *testing.T) {
	numNotes := 50
	textParas := 10
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	if err != nil {
		t.Errorf("sign-in failed: %+v", err)
	}

	err = createNotes(session, numNotes, textParas)
	if err != nil {
		t.Errorf("error: %+v", err)
	}

	wipeConfig := WipeConfig{
		Session: session,
	}
	var deleted int
	deleted, err = wipeConfig.Run()
	if err != nil {
		t.Errorf("error: %+v\n", err)
	}
	if deleted != numNotes {
		t.Errorf("error: created %d notes but deleted %d\n", numNotes, deleted)
	}

}

func TestCreateOneHundredNotes(t *testing.T) {
	numNotes := 100
	textParas := 10
	session, err := CliSignIn(os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"), os.Getenv("SN_SERVER"))
	if err != nil {
		t.Errorf("sign-in failed: %+v", err)
	}

	err = createNotes(session, numNotes, textParas)
	if err != nil {
		t.Errorf("error: %+v", err)
	}

	noteFilter := gosn.Filter{
		Type: "Note",
	}
	filter := gosn.ItemFilters{
		Filters: []gosn.Filter{noteFilter},
	}

	gnc := GetNoteConfig{
		Session: session,
		Filters: filter,
	}
	var res gosn.GetItemsOutput
	res, err = gnc.Run()
	if err != nil {
		t.Errorf("error: failed to retrieve items: %+v\n", err)
	}
	if len(res.Items) != numNotes {
		t.Errorf("error: expected %d notes but got %d items\n", numNotes, len(res.Items))
	}
	wipeConfig := WipeConfig{
		Session: session,
	}
	var deleted int
	deleted, err = wipeConfig.Run()
	if err != nil {
		t.Errorf("error: %+v\n", err)
	}
	if deleted != numNotes {
		t.Errorf("error: created %d notes but deleted %d\n", numNotes, deleted)
	}
}
