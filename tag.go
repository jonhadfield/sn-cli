package sncli

import (
	"log"
	"strings"

	"github.com/jonhadfield/gosn"
)

type tagNotesInput struct {
	session        gosn.Session
	matchTitle     string
	matchText      string
	matchTags      []string
	matchNoteUUIDs []string
	newTags        []string
	syncToken      string
}

func tagNotes(input tagNotesInput) (newSyncToken string, err error) {
	// create tags if they don't exist
	ati := addTagsInput{
		session:   input.session,
		tagTitles: input.newTags,
		syncToken: input.syncToken,
	}
	newSyncToken, err = addTags(ati)
	if err != nil {
		return newSyncToken, err
	}
	// get notes and tags
	getNotesFilter := gosn.Filter{
		Type: "Note",
	}
	getTagsFilter := gosn.Filter{
		Type: "Tag",
	}
	filters := []gosn.Filter{getNotesFilter, getTagsFilter}
	itemFilter := gosn.ItemFilters{
		MatchAny: true,
		Filters:  filters,
	}

	getItemsInput := gosn.GetItemsInput{
		Session:   input.session,
		SyncToken: input.syncToken,
		Filters:   itemFilter,
	}
	var output gosn.GetItemsOutput
	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return newSyncToken, err
	}
	output.DeDupe()
	var allTags []gosn.Item
	var allNotes []gosn.Item
	// create slices of notes and tags
	for _, item := range output.Items {
		if item.Deleted {
			continue
		}
		if item.ContentType == "Tag" {
			allTags = append(allTags, item)
		}
		if item.ContentType == "Note" {
			allNotes = append(allNotes, item)
		}
	}
	typeUUIDs := make(map[string][]string)
	// loop through all notes and create a list of those that
	// match the note title or exist in note text
	for _, note := range allNotes {
		if StringInSlice(note.UUID, input.matchNoteUUIDs, false) {
			typeUUIDs["Note"] = append(typeUUIDs["Note"], note.UUID)
		} else if strings.TrimSpace(input.matchTitle) != "" && strings.Contains(strings.ToLower(note.Content.GetTitle()), strings.ToLower(input.matchTitle)) {
			typeUUIDs["Note"] = append(typeUUIDs["Note"], note.UUID)
		} else if strings.TrimSpace(input.matchText) != "" && strings.Contains(strings.ToLower(note.Content.GetText()), strings.ToLower(input.matchText)) {
			typeUUIDs["Note"] = append(typeUUIDs["Note"], note.UUID)
		}
	}

	// update existing (and just created) tags to reference matching uuids
	// determine which TAGS need updating and create list to sync back to server
	var tagsToPush []gosn.Item
	for _, t := range allTags {
		// if tag title is in ones to add then update tag with new references
		if StringInSlice(t.Content.GetTitle(), input.newTags, true) {
			// does it need updating
			updatedTag, changed := upsertTagReferences(t, typeUUIDs)
			if changed {
				tagsToPush = append(tagsToPush, updatedTag)
			}
		}
	}

	if len(tagsToPush) > 0 {
		pii := gosn.PutItemsInput{
			Items:     tagsToPush,
			SyncToken: input.syncToken,
			Session:   input.session,
		}
		var putItemsOutput gosn.PutItemsOutput
		putItemsOutput, err = gosn.PutItems(pii)
		if err != nil {
			return
		}
		newSyncToken = putItemsOutput.ResponseBody.SyncToken
		return newSyncToken, err
	}
	return newSyncToken, nil

}

func (input *TagItemsConfig) Run() error {
	gosn.SetErrorLogger(log.Println)
	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}
	tni := tagNotesInput{
		matchTitle: input.FindTitle,
		matchText:  input.FindText,
		matchTags:  []string{input.FindTag},
		newTags:    input.NewTags,
		session:    input.Session,
	}
	_, err := tagNotes(tni)
	return err

}

func (input *AddTagConfig) Run() error {
	gosn.SetErrorLogger(log.Println)
	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}

	ati := addTagsInput{

		tagTitles: input.Tags,
		session:   input.Session,
	}
	var err error
	_, err = addTags(ati)
	return err
}

func (input *GetTagConfig) Run() (output gosn.GetItemsOutput, err error) {
	gosn.SetErrorLogger(log.Println)
	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}

	getItemsInput := gosn.GetItemsInput{
		Session: input.Session,
		Filters: input.Filters,
	}
	output, err = gosn.GetItems(getItemsInput)
	output.DeDupe()
	return
}

func (input *DeleteTagConfig) Run() (noDeleted int, err error) {
	gosn.SetErrorLogger(log.Println)
	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}
	noDeleted, _, err = deleteTags(input.Session, input.TagTitles, input.TagUUIDs, "")
	return noDeleted, err
}

func deleteTags(session gosn.Session, tagTitles []string, tagUUIDs []string, syncToken string) (noDeleted int, newSyncToken string, err error) {
	deleteNotesFilter := gosn.Filter{
		Type: "Note",
	}
	deleteTagsFilter := gosn.Filter{
		Type: "Tag",
	}
	filters := []gosn.Filter{deleteNotesFilter, deleteTagsFilter}
	deleteFilter := gosn.ItemFilters{
		MatchAny: true,
		Filters:  filters,
	}

	getItemsInput := gosn.GetItemsInput{
		Session:   session,
		SyncToken: syncToken,
		Filters:   deleteFilter,
	}
	output, err := gosn.GetItems(getItemsInput)
	output.DeDupe()
	var tagsToDelete []gosn.Item
	for _, item := range output.Items {
		if item.Deleted {
			continue
		}
		if item.ContentType == "Tag" &&
			(StringInSlice(item.UUID, tagUUIDs, true) ||
				StringInSlice(item.Content.GetTitle(), tagTitles, true)) {
			item.Deleted = true
			tagsToDelete = append(tagsToDelete, item)
		}
	}

	if len(tagsToDelete) > 0 {
		pii := gosn.PutItemsInput{
			Items:     tagsToDelete,
			SyncToken: syncToken,
			Session:   session,
		}
		var putItemsOutput gosn.PutItemsOutput
		putItemsOutput, err = gosn.PutItems(pii)
		if err != nil {
			return
		}
		newSyncToken = putItemsOutput.ResponseBody.SyncToken
	}
	noDeleted = len(tagsToDelete)
	return
}

type addTagsInput struct {
	session   gosn.Session
	tagTitles []string
	syncToken string
}

func addTags(input addTagsInput) (newSyncToken string, err error) {
	addNotesFilter := gosn.Filter{
		Type: "Note",
	}
	addTagsFilter := gosn.Filter{
		Type: "Tag",
	}
	filters := []gosn.Filter{addNotesFilter, addTagsFilter}
	addFilter := gosn.ItemFilters{
		MatchAny: true,
		Filters:  filters,
	}

	getItemsInput := gosn.GetItemsInput{
		SyncToken: input.syncToken,
		Session:   input.session,
		Filters:   addFilter,
	}
	output, err := gosn.GetItems(getItemsInput)
	if err != nil {
		return
	}
	output.DeDupe()
	var allTags []gosn.Item
	for _, item := range output.Items {
		if item.Deleted {
			continue
		}
		if item.ContentType == "Tag" {
			allTags = append(allTags, item)
		}
	}

	var tagsToAdd []gosn.Item
	for _, tag := range input.tagTitles {
		if !tagExists(allTags, tag) {
			newTagContent := gosn.NewTagContent()
			newTag := gosn.NewTag()
			newTagContent.Title = tag
			newTag.Content = newTagContent
			newTag.UUID = gosn.GenUUID()
			tagsToAdd = append(tagsToAdd, *newTag)
		}
	}
	if len(tagsToAdd) > 0 {
		putItemsInput := gosn.PutItemsInput{
			Session: input.session,
			Items:   tagsToAdd,
		}
		var putItemsOutput gosn.PutItemsOutput
		putItemsOutput, err = gosn.PutItems(putItemsInput)
		if err != nil {
			return
		}
		newSyncToken = putItemsOutput.ResponseBody.SyncToken
	}
	return
}

func upsertTagReferences(tag gosn.Item, typeUUIDs map[string][]string) (gosn.Item, bool) {
	// create item reference
	var newReferences []gosn.ItemReference
	var changed bool
	for k, v := range typeUUIDs {
		for _, ref := range v {
			if !referenceExists(tag, ref) {
				newReferences = append(newReferences, gosn.ItemReference{
					ContentType: k,
					UUID:        ref,
				})
			}
		}
	}
	if len(newReferences) > 0 {
		changed = true
		newContent := tag.Content
		newContent.UpsertReferences(newReferences)

		tag.Content = newContent
	}

	return tag, changed
}

func tagExists(existing []gosn.Item, find string) bool {
	for _, tag := range existing {
		if tag.Content.GetTitle() == find {
			return true
		}
	}
	return false
}
