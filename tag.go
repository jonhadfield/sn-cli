package sncli

import (
	"strings"

	"github.com/jonhadfield/gosn-v2"
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

	var ato addTagsOutput

	ato, err = addTags(ati)
	if err != nil {
		return
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

	getItemsInput := gosn.SyncInput{
		Session:   input.session,
		SyncToken: ato.newSyncToken,
	}

	var output gosn.SyncOutput

	output, err = gosn.Sync(getItemsInput)
	if err != nil {
		return
	}

	output.Items.DeDupe()

	var items gosn.Items

	items, err = output.Items.DecryptAndParse(input.session.Mk, input.session.Ak, false)
	if err != nil {
		return
	}

	items.Filter(itemFilter)

	var allTags []*gosn.Tag

	var allNotes []*gosn.Note
	// create slices of notes and tags
	for _, item := range items {
		if item.IsDeleted() {
			continue
		}

		if item.GetContentType() == "Tag" {
			allTags = append(allTags, item.(*gosn.Tag))
		}

		if item.GetContentType() == "Note" {
			allNotes = append(allNotes, item.(*gosn.Note))
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
	var tagsToPush gosn.Tags

	for _, t := range allTags {
		// if tag title is in ones to add then update tag with new references
		if StringInSlice(t.Content.GetTitle(), input.newTags, true) {
			// does it need updating
			updatedTag, changed := upsertTagReferences(*t, typeUUIDs)
			if changed {
				tagsToPush = append(tagsToPush, updatedTag)
			}
		}
	}

	var eTagsToPush gosn.EncryptedItems

	eTagsToPush, err = tagsToPush.Encrypt(input.session.Mk, input.session.Ak, false)
	if err != nil {
		return
	}

	if len(tagsToPush) > 0 {
		pii := gosn.SyncInput{
			Items:     eTagsToPush,
			SyncToken: input.syncToken,
			Session:   input.session,
		}

		var putItemsOutput gosn.SyncOutput

		putItemsOutput, err = gosn.Sync(pii)
		if err != nil {
			return
		}

		newSyncToken = putItemsOutput.SyncToken

		return newSyncToken, err
	}

	return newSyncToken, nil
}

func (input *TagItemsConfig) Run() error {
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

func (input *AddTagsInput) Run() (output AddTagsOutput, err error) {
	ati := addTagsInput{
		tagTitles: input.Tags,
		session:   input.Session,
	}

	ato, err := addTags(ati)
	if err != nil {
		return
	}

	output.Added = ato.added
	output.Existing = ato.existing
	output.SyncToken = ato.newSyncToken

	return
}

func (input *GetTagConfig) Run() (tags gosn.Items, err error) {
	getItemsInput := gosn.SyncInput{
		Session: input.Session,
	}

	var output gosn.SyncOutput

	output, err = gosn.Sync(getItemsInput)
	if err != nil {
		return nil, err
	}

	output.Items.DeDupe()

	tags, err = output.Items.DecryptAndParse(input.Session.Mk, input.Session.Ak, input.Debug)

	if err != nil {
		return nil, err
	}

	tags.Filter(input.Filters)

	return
}

func (input *DeleteTagConfig) Run() (noDeleted int, err error) {
	noDeleted, _, err = deleteTags(input.Session, input.TagTitles, input.TagUUIDs, "")

	return noDeleted, err
}

func deleteTags(session gosn.Session, tagTitles []string, tagUUIDs []string, syncToken string) (noDeleted int, newSyncToken string, err error) {
	//deleteNotesFilter := gosn.Filter{
	//	Type: "Note",
	//}
	deleteTagsFilter := gosn.Filter{
		Type: "Tag",
	}
	filters := []gosn.Filter{deleteTagsFilter}
	//filters := []gosn.Filter{deleteNotesFilter, deleteTagsFilter}
	deleteFilter := gosn.ItemFilters{
		MatchAny: true,
		Filters:  filters,
	}

	getItemsInput := gosn.SyncInput{
		Session:   session,
		SyncToken: syncToken,
	}

	var output gosn.SyncOutput

	output, err = gosn.Sync(getItemsInput)
	if err != nil {
		return 0, "", err
	}

	output.Items.DeDupe()

	var tags gosn.Items

	tags, err = output.Items.DecryptAndParse(session.Mk, session.Ak, false)
	if err != nil {
		return 0, output.SyncToken, err
	}

	tags.Filter(deleteFilter)

	var tagsToDelete gosn.Items

	for _, item := range tags {
		if item.IsDeleted() {
			continue
		}

		var tag *gosn.Tag
		if item.GetContentType() == "Tag" {
			tag = item.(*gosn.Tag)
		} else {
			continue
		}

		if StringInSlice(tag.GetUUID(), tagUUIDs, true) ||
			StringInSlice(tag.Content.Title, tagTitles, true) {
			tag.Deleted = true
			tagsToDelete = append(tagsToDelete, tag)
		}
	}

	var eTagsToDelete gosn.EncryptedItems
	eTagsToDelete, err = tagsToDelete.Encrypt(session.Mk, session.Ak, false)

	if len(tagsToDelete) > 0 {
		pii := gosn.SyncInput{
			Items:     eTagsToDelete,
			SyncToken: syncToken,
			Session:   session,
		}

		var putItemsOutput gosn.SyncOutput

		putItemsOutput, err = gosn.Sync(pii)
		if err != nil {
			return
		}

		newSyncToken = putItemsOutput.SyncToken
	}

	noDeleted = len(tagsToDelete)

	return noDeleted, newSyncToken, err
}

type addTagsInput struct {
	session   gosn.Session
	tagTitles []string
	syncToken string
}

type addTagsOutput struct {
	newSyncToken string
	added        []string
	existing     []string
}

func addTags(ati addTagsInput) (ato addTagsOutput, err error) {
	addTagsFilter := gosn.Filter{
		Type: "Tag",
	}
	filters := []gosn.Filter{addTagsFilter}
	addFilter := gosn.ItemFilters{
		MatchAny: true,
		Filters:  filters,
	}

	getItemsInput := gosn.SyncInput{
		SyncToken: ati.syncToken,
		Session:   ati.session,
	}

	output, err := gosn.Sync(getItemsInput)
	if err != nil {
		return
	}

	output.Items.DeDupe()

	var items gosn.Items

	items, err = output.Items.DecryptAndParse(ati.session.Mk, ati.session.Ak, false)
	if err != nil {
		return
	}

	items.Filter(addFilter)

	var allTags gosn.Tags

	for _, item := range items {
		if item.IsDeleted() {
			continue
		}

		if item.GetContentType() == "Tag" {
			tag := item.(*gosn.Tag)
			allTags = append(allTags, *tag)
		}
	}

	var tagsToAdd gosn.Tags

	for _, tag := range ati.tagTitles {
		if tagExists(allTags, tag) {
			ato.existing = append(ato.existing, tag)
			continue
		}

		newTagContent := gosn.NewTagContent()
		newTag := gosn.NewTag()
		newTagContent.Title = tag
		newTag.Content = *newTagContent
		newTag.UUID = gosn.GenUUID()

		tagsToAdd = append(tagsToAdd, newTag)
		ato.added = append(ato.added, tag)
	}

	if len(tagsToAdd) > 0 {
		var eTagsToAdd gosn.EncryptedItems
		eTagsToAdd, err = tagsToAdd.Encrypt(ati.session.Mk, ati.session.Ak, false)

		if err != nil {
			return
		}

		putItemsInput := gosn.SyncInput{
			Session: ati.session,
			Items:   eTagsToAdd,
		}

		var putItemsOutput gosn.SyncOutput

		putItemsOutput, err = gosn.Sync(putItemsInput)
		if err != nil {
			return
		}

		ato.newSyncToken = putItemsOutput.SyncToken
	}

	return ato, err
}

func upsertTagReferences(tag gosn.Tag, typeUUIDs map[string][]string) (gosn.Tag, bool) {
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

func tagExists(existing []gosn.Tag, find string) bool {
	for _, tag := range existing {
		if tag.Content.GetTitle() == find {
			return true
		}
	}

	return false
}
