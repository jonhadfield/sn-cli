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

	var noDeleted int
	noDeleted, err = deleteTagConfig.Run()
	assert.Equal(t, 2, noDeleted)
	assert.NoError(t, err, err)
}

func TestGetTag(t *testing.T) {
	defer cleanUp(&testSession)

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

	var output gosn.Items
	output, err = getTagConfig.Run()
	assert.NoError(t, err, err)
	assert.EqualValues(t, len(output), 2, "expected two items but got: %+v", output)
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

func _deleteNotesByTitle(session gosn.Session, input map[string]string) (noDeleted int, err error) {
	for k := range input {
		deleteNoteConfig := DeleteNoteConfig{
			Session:    session,
			NoteTitles: []string{k},
		}
		_, err = deleteNoteConfig.Run()

		if err != nil {
			return noDeleted, err
		}
		noDeleted++
	}

	return noDeleted, err
}

func _deleteTagsByTitle(session gosn.Session, input []string) (noDeleted int, err error) {
	deleteTagConfig := DeleteTagConfig{
		Session:   session,
		TagTitles: input,
	}

	return deleteTagConfig.Run()
}

func TestTaggingOfNotes(t *testing.T) {
	defer cleanUp(&testSession)

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

	var retNotes gosn.Items
	retNotes, err = gnc.Run()
	assert.NoError(t, err, err)

	if len(retNotes) != 2 {
		t.Errorf("expected two notes but got: %d", len(retNotes))
	}

	_, err = _deleteNotesByTitle(sOutput.Session, notes)
	assert.NoError(t, err, err)

	var deletedTags int
	deletedTags, err = _deleteTagsByTitle(sOutput.Session, tags)
	assert.NoError(t, err, err)
	assert.Equal(t, len(tags), deletedTags)
}
