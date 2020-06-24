package sncli

import (
	"testing"

	gosn "github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/stretchr/testify/assert"
)

func TestAddDeleteTagByTitle(t *testing.T) {
	addTagConfig := AddTagsInput{
		Session: testSession,
		Tags:    []string{"TestTagOne", "TestTagTwo"},
	}

	ato, err := addTagConfig.Run()
	assert.NoError(t, err)
	assert.Contains(t, ato.Added, "TestTagOne")
	assert.Contains(t, ato.Added, "TestTagTwo")
	assert.Empty(t, ato.Existing)

	deleteTagConfig := DeleteTagConfig{
		Session:   testSession,
		TagTitles: []string{"TestTagOne", "TestTagTwo"},
	}

	var noDeleted int
	noDeleted, err = deleteTagConfig.Run()
	assert.Equal(t, 2, noDeleted)
	assert.NoError(t, err, err)
}

func TestGetTag(t *testing.T) {
	defer cleanUp(testSession)

	testTagTitles := []string{"TestTagOne", "TestTagTwo"}
	addTagInput := AddTagsInput{
		Session: testSession,
		Tags:    testTagTitles,
	}

	ato, err := addTagInput.Run()
	assert.NoError(t, err, err)
	assert.NoError(t, err)
	assert.Contains(t, ato.Added, "TestTagOne")
	assert.Contains(t, ato.Added, "TestTagTwo")
	assert.Empty(t, ato.Existing)

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
		Session: testSession,
		Filters: getTagFilters,
	}

	var output gosn.Items
	output, err = getTagConfig.Run()
	assert.NoError(t, err, err)
	assert.EqualValues(t, len(output), 2, "expected two items but got: %+v", output)
}

func _addNotes(session cache.Session, input map[string]string) error {
	for k, v := range input {
		addNoteConfig := AddNoteInput{
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

func _deleteNotesByTitle(session cache.Session, input map[string]string) (noDeleted int, err error) {
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

func _deleteTagsByTitle(session cache.Session, input []string) (noDeleted int, err error) {
	deleteTagConfig := DeleteTagConfig{
		Session:   session,
		TagTitles: input,
	}

	return deleteTagConfig.Run()
}

func TestTaggingOfNotes(t *testing.T) {
	defer cleanUp(testSession)

	// create four notes
	notes := map[string]string{
		"noteOneTitle":   "noteOneText example",
		"noteTwoTitle":   "noteTwoText",
		"noteThreeTitle": "noteThreeText",
		"noteFourTitle":  "noteFourText example",
	}

	err := _addNotes(testSession, notes)
	assert.NoError(t, err, err)
	// tag new notes with 'testTag'
	tags := []string{"testTag"}
	tni := TagItemsConfig{
		Session:  testSession,
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
		Session: testSession,
		Filters: itemFilters,
	}

	var retNotes gosn.Items
	retNotes, err = gnc.Run()
	assert.NoError(t, err, err)

	if len(retNotes) != 2 {
		t.Errorf("expected two notes but got: %d", len(retNotes))
	}

	_, err = _deleteNotesByTitle(testSession, notes)
	assert.NoError(t, err, err)

	var deletedTags int
	deletedTags, err = _deleteTagsByTitle(testSession, tags)
	assert.NoError(t, err, err)
	assert.Equal(t, len(tags), deletedTags)
}
