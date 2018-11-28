package sncli

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jonhadfield/gosn"
)

func TestAddDeleteTagByTitle(t *testing.T) {
	sOutput, err := gosn.SignIn(sInput)
	assert.NoError(t, err, err)
	addTagConfig := AddTagConfig{
		Session: sOutput.Session,
		Tags:    []string{"TestTagOne", "TestTagTwo"},
	}
	err = addTagConfig.Run()
	assert.NoError(t, err, err)
	deleteTagConfig := DeleteTagConfig{
		Session:   sOutput.Session,
		TagTitles: []string{"TestTagOne", "TestTagTwo"},
	}
	err = deleteTagConfig.Run()
	assert.NoError(t, err, err)
}

func TestGetTag(t *testing.T) {
	sOutput, err := gosn.SignIn(sInput)
	assert.NoError(t, err, err)
	testTagTitles := []string{"TestTagOne", "TestTagTwo"}
	addTagConfig := AddTagConfig{
		Session: sOutput.Session,
		Tags:    testTagTitles,
	}
	err = addTagConfig.Run()
	assert.NoError(t, err, err)

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
	assert.NoError(t, err, err)
	assert.EqualValues(t, len(output.Items), 2, "expected two items but got: %+v", output.Items)

	// clean up
	deleteTagConfig := DeleteTagConfig{
		Session:   sOutput.Session,
		TagTitles: []string{"TestTagOne", "TestTagTwo"},
	}
	err = deleteTagConfig.Run()
	assert.NoError(t, err, err)

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
	assert.NoError(t, err, err)

	// tag new notes with 'testTag'
	tags := []string{"testTag"}
	tni := TagItemsConfig{
		Session:  sOutput.Session,
		FindText: "example",
		NewTags:  tags,
	}
	err = tni.Run()
	assert.NoError(t, err, err)

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
	assert.NoError(t, err, err)

	if len(retNotes.Items) != 2 {
		t.Errorf("expected two notes but got: %d", len(retNotes.Items))
	}

	err = _deleteNotesByTitle(sOutput.Session, notes)
	assert.NoError(t, err, err)

	err = _deleteTagsByTitle(sOutput.Session, tags)
	assert.NoError(t, err, err)

}
