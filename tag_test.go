package sncli

import (
	"testing"

	"github.com/jonhadfield/gosn"
)

func TestAddDeleteTagByTitle(t *testing.T) {
	sOutput, signInErr := gosn.SignIn(sInput)
	if signInErr != nil {
		t.Errorf("CliSignIn error:: %+v", signInErr)
	}
	addTagConfig := AddTagConfig{
		Session: sOutput.Session,
		Tags:    []string{"TestTagOne", "TestTagTwo"},
	}
	err := addTagConfig.Run()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
	deleteTagConfig := DeleteTagConfig{
		Session:   sOutput.Session,
		TagTitles: []string{"TestTagOne", "TestTagTwo"},
	}
	err = deleteTagConfig.Run()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestGetTag(t *testing.T) {
	sOutput, signInErr := gosn.SignIn(sInput)
	if signInErr != nil {
		t.Errorf("CliSignIn error:: %+v", signInErr)
	}
	testTagTitles := []string{"TestTagOne", "TestTagTwo"}
	addTagConfig := AddTagConfig{
		Session: sOutput.Session,
		Tags:    testTagTitles,
	}
	err := addTagConfig.Run()
	if err != nil {
		t.Errorf("%+v", err)
	}

	// create filters
	getTagFilters := gosn.ItemFilters{
		MatchAny: true,
	}

	for _, testTagTitle := range testTagTitles {
		getTagFilters.Filters = append(getTagFilters.Filters, gosn.Filter{
			Key:        "Title",
			Value:      testTagTitle,
			Type:       "Tag",
			Comparison: "==",
		})
	}
	getTagConfig := GetTagConfig{
		Session: sOutput.Session,
		Filters: getTagFilters,
	}
	var output gosn.GetItemsOutput
	output, err = getTagConfig.Run()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
	if len(output.Items) != 2 {
		t.Errorf("expected two items but got: %+v", output.Items)
	}

	// clean up
	deleteTagConfig := DeleteTagConfig{
		Session:   sOutput.Session,
		TagTitles: []string{"TestTagOne", "TestTagTwo"},
	}
	err = deleteTagConfig.Run()
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func _addNotes(session gosn.Session, input map[string]string) error {
	for k, v := range input {
		addNoteConfig := AddNoteConfig{
			Session: session,
			Title:   k,
			Text:    v,
		}
		err := addNoteConfig.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func _deleteNotesByTitle(session gosn.Session, input map[string]string) error {
	for k := range input {
		deleteNoteConfig := DeleteNoteConfig{
			Session:    session,
			NoteTitles: []string{k},
		}
		err := deleteNoteConfig.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func _deleteTagsByTitle(session gosn.Session, input []string) error {

	deleteTagConfig := DeleteTagConfig{
		Session:   session,
		TagTitles: input,
	}
	err := deleteTagConfig.Run()
	return err
}

func TestTaggingOfNotes(t *testing.T) {
	sOutput, signInErr := gosn.SignIn(sInput)
	if signInErr != nil {
		t.Errorf("CliSignIn error:: %+v", signInErr)
	}
	// create four notes
	notes := map[string]string{
		"noteOneTitle":   "noteOneText example",
		"noteTwoTitle":   "noteTwoText",
		"noteThreeTitle": "noteThreeText",
		"noteFourTitle":  "noteFourText example",
	}

	err := _addNotes(sOutput.Session, notes)
	if err != nil {
		t.Errorf("failed to add notes")
	}
	// tag new notes with 'testTag'
	tags := []string{"testTag"}
	tni := TagItemsConfig{
		Session:  sOutput.Session,
		FindText: "example",
		NewTags:  tags,
	}
	err = tni.Run()
	if err != nil {
		t.Errorf("%+v", err)
	}
	// get newly tagged notes

	filterNotesByTagName := gosn.Filter{
		Type:       "Note",
		Key:        "TagTitle",
		Comparison: "==",
		Value:      "testTag",
	}
	itemFilters := gosn.ItemFilters{
		Filters:  []gosn.Filter{filterNotesByTagName},
		MatchAny: true,
	}
	gnc := GetNoteConfig{
		Session: sOutput.Session,
		Filters: itemFilters,
	}

	var retNotes gosn.GetItemsOutput
	retNotes, err = gnc.Run()
	if err != nil {
		return
	}
	if len(retNotes.Items) != 2 {
		t.Errorf("expected two notes but got: %d", len(retNotes.Items))
	}

	err = _deleteNotesByTitle(sOutput.Session, notes)
	if err != nil {
		t.Errorf("failed to clean up notes")
	}

	err = _deleteTagsByTitle(sOutput.Session, tags)
	if err != nil {
		t.Errorf("failed to clean up notes")
	}
}
