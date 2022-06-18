package sncli

import (
	"fmt"
	"strings"

	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

type TagItemsInput struct {
	Session        *cache.Session
	ItemType       string
	MatchTitle     string
	MatchText      string
	MatchTags      []string
	MatchNoteUUIDs []string
	NewTags        []string
	Referers       []string
	Replace        bool
}

// create tags if they don't exist
// get all notes and tags.
func TagItems(i TagItemsInput) (err error) {
	fmt.Printf("IN TAGITEMS WITH: %+v\n", i)
	// create tags, including referers, if they don't exist
	//fmt.Printf("TAG Titles Combined: %+v\n", tt)
	_, err = addTags(addTagsInput{
		session:   i.Session,
		tagTitles: i.NewTags,
		referers:  i.Referers,
		replace:   i.Replace,
	})
	if err != nil {
		return
	}

	_, err = addTags(addTagsInput{
		session:   i.Session,
		tagTitles: i.Referers,
		//referers:  i.Referers,
		replace: i.Replace,
	})
	if err != nil {
		return
	}

	// get notes and tags
	filters := []gosn.Filter{{
		Type: "Tag",
	}}
	if i.ItemType == "Note" {
		filters = append(filters, gosn.Filter{
			Type: "Note",
		})
	}

	itemFilter := gosn.ItemFilters{
		MatchAny: true,
		Filters:  filters,
	}

	// get all notes and tags from db
	so, err := Sync(cache.SyncInput{
		Session: i.Session,
	}, true)
	if err != nil {
		return
	}

	var allPersistedItems cache.Items
	if err = so.DB.All(&allPersistedItems); err != nil {
		return
	}

	items, err := allPersistedItems.ToItems(i.Session)
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

	noteTypeUUIDs := make(map[string][]string)
	// loop through all notes and create a list of those that
	// match the note title or exist in note text
	for _, note := range allNotes {
		switch {
		case StringInSlice(note.UUID, i.MatchNoteUUIDs, false):
			noteTypeUUIDs["Note"] = append(noteTypeUUIDs["Note"], note.UUID)
		case strings.TrimSpace(i.MatchTitle) != "" && strings.Contains(strings.ToLower(note.Content.GetTitle()), strings.ToLower(i.MatchTitle)):
			noteTypeUUIDs["Note"] = append(noteTypeUUIDs["Note"], note.UUID)
		case strings.TrimSpace(i.MatchText) != "" && strings.Contains(strings.ToLower(note.Content.GetText()), strings.ToLower(i.MatchText)):
			noteTypeUUIDs["Note"] = append(noteTypeUUIDs["Note"], note.UUID)
		}
	}

	tagTypeUUIDs := make(map[string][]string)
	// loop through all tags and create a list of those that
	// match the tag title
	for _, tag := range allTags {
		switch {
		case StringInSlice(tag.UUID, i.MatchNoteUUIDs, false):
			noteTypeUUIDs["Tag"] = append(noteTypeUUIDs["Tag"], tag.UUID)
		case strings.TrimSpace(i.MatchTitle) != "" && strings.Contains(strings.ToLower(tag.Content.GetTitle()), strings.ToLower(i.MatchTitle)):
			noteTypeUUIDs["Tag"] = append(noteTypeUUIDs["Tag"], tag.UUID)
		}
	}

	// update existing (and just created) tags to reference matching uuids
	// determine which tags need updating and create list to sync back to server
	var tagsToPush gosn.Tags

	for _, t := range allTags {
		fmt.Printf("t: %+v\n", t)
		// if tag title is in ones to add then update tag with new references
		if StringInSlice(t.Content.GetTitle(), i.NewTags, true) {
			fmt.Printf("t: %+v is in %+v\n", t, i.NewTags)

			var updatedTag gosn.Tag
			var changed bool
			if i.ItemType == "Note" {
				updatedTag, changed = upsertTagReferences(*t, noteTypeUUIDs)
			} else if i.ItemType == "Tag" {
				updatedTag, changed = upsertTagReferences(*t, tagTypeUUIDs)
			}
			if changed {
				tagsToPush = append(tagsToPush, updatedTag)
			}
		}
	}

	if len(tagsToPush) > 0 {
		fmt.Printf("tagsToPush: %+v\n", tagsToPush)
		if err = cache.SaveTags(so.DB, i.Session, tagsToPush, true); err != nil {
			return
		}

		so, err = Sync(cache.SyncInput{
			Session: i.Session,
		}, true)
		if err != nil {
			return
		}

		return so.DB.Close()
	}

	return nil
}

type tagNotesInput struct {
	session        *cache.Session
	matchTitle     string
	matchText      string
	matchTags      []string
	matchNoteUUIDs []string
	newTags        []string
	syncToken      string
	replace        bool
}

// create tags if they don't exist
// get all notes and tags.
func tagNotes(i tagNotesInput) (err error) {
	// create tags if they don't exist
	ati := addTagsInput{
		session:   i.session,
		tagTitles: i.newTags,
		replace:   i.replace,
	}

	_, err = addTags(ati)

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

	syncInput := cache.SyncInput{
		Session: i.session,
	}

	// get all notes and tags from db
	var so cache.SyncOutput

	so, err = Sync(syncInput, true)
	if err != nil {
		return
	}

	var allPersistedItems cache.Items
	if err = so.DB.All(&allPersistedItems); err != nil {
		return
	}

	var items gosn.Items

	items, err = allPersistedItems.ToItems(i.session)
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
		switch {
		case StringInSlice(note.UUID, i.matchNoteUUIDs, false):
			typeUUIDs["Note"] = append(typeUUIDs["Note"], note.UUID)
		case strings.TrimSpace(i.matchTitle) != "" && strings.Contains(strings.ToLower(note.Content.GetTitle()), strings.ToLower(i.matchTitle)):
			typeUUIDs["Note"] = append(typeUUIDs["Note"], note.UUID)
		case strings.TrimSpace(i.matchText) != "" && strings.Contains(strings.ToLower(note.Content.GetText()), strings.ToLower(i.matchText)):
			typeUUIDs["Note"] = append(typeUUIDs["Note"], note.UUID)
		}
	}

	// update existing (and just created) tags to reference matching uuids
	// determine which TAGS need updating and create list to sync back to server
	var tagsToPush gosn.Tags

	for _, t := range allTags {
		// if tag title is in ones to add then update tag with new references
		if StringInSlice(t.Content.GetTitle(), i.newTags, true) {
			// does it need updating
			updatedTag, changed := upsertTagReferences(*t, typeUUIDs)
			if changed {
				tagsToPush = append(tagsToPush, updatedTag)
			}
		}
	}

	if len(tagsToPush) > 0 {
		if err = cache.SaveTags(so.DB, i.session, tagsToPush, true); err != nil {
			return
		}

		pii := cache.SyncInput{
			Session: i.session,
		}

		so, err = Sync(pii, true)
		if err != nil {
			return
		}

		if err = so.DB.Close(); err != nil {
			return
		}

		return err
	}

	return nil
}

func (i *TagItemsConfig) Run() error {
	tni := tagNotesInput{
		matchTitle: i.FindTitle,
		matchText:  i.FindText,
		matchTags:  []string{i.FindTag},
		newTags:    i.NewTags,
		session:    i.Session,
	}

	return tagNotes(tni)
}

func (i *AddTagsInput) Run() (output AddTagsOutput, err error) {
	// Sync DB
	si := cache.SyncInput{
		Session: i.Session,
		Close:   false,
	}

	var so cache.SyncOutput
	so, err = Sync(si, true)

	if err != nil {
		return
	}

	err = so.DB.Close()
	if err != nil {
		return
	}

	defer func() {
		_ = so.DB.Close()
	}()

	ati := addTagsInput{
		tagTitles: i.Tags,
		referers:  i.ReferringTags,
		session:   i.Session,
		replace:   i.Replace,
	}

	var ato addTagsOutput
	ato, err = addTags(ati)
	if err != nil {
		return
	}

	output.Added = ato.added
	output.Existing = ato.existing

	// Sync DB with SN
	err = so.DB.Close()
	if err != nil {
		return
	}

	so, err = Sync(cache.SyncInput{
		Session: i.Session,
	}, true)
	if err != nil {
		return
	}

	return output, err
}

func (i *GetTagConfig) Run() (items gosn.Items, err error) {
	var so cache.SyncOutput

	si := cache.SyncInput{
		Session: i.Session,
	}

	so, err = Sync(si, true)
	if err != nil {
		return
	}

	var allPersistedItems cache.Items

	err = so.DB.All(&allPersistedItems)
	if err != nil {
		return
	}

	err = so.DB.Close()
	if err != nil {
		return
	}

	items, err = allPersistedItems.ToItems(i.Session)
	if err != nil {
		return
	}

	items.Filter(i.Filters)

	return items, err
}

func (i *DeleteTagConfig) Run() (noDeleted int, err error) {
	noDeleted, err = deleteTags(i.Session, i.TagTitles, i.TagUUIDs)
	return noDeleted, err
}

func deleteTags(session *cache.Session, tagTitles []string, tagUUIDs []string) (noDeleted int, err error) {
	deleteTagsFilter := gosn.Filter{
		Type: "Tag",
	}
	filters := []gosn.Filter{deleteTagsFilter}
	deleteFilter := gosn.ItemFilters{
		MatchAny: true,
		Filters:  filters,
	}

	syncInput := cache.SyncInput{
		Session: session,
	}

	// load db
	var so cache.SyncOutput

	so, err = Sync(syncInput, true)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = so.DB.Close()
	}()

	var tags gosn.Items

	// get items from db
	var allPersistedItems cache.Items

	err = so.DB.All(&allPersistedItems)
	if err != nil {
		return
	}

	var items gosn.Items

	items, err = allPersistedItems.ToItems(session)
	if err != nil {
		return
	}

	tags = items
	tags.Filter(deleteFilter)

	var tagsToDelete gosn.Items

	for _, tag := range tags {
		if tag.IsDeleted() {
			continue
		}

		var gTag *gosn.Tag
		if tag.GetContentType() == "Tag" {
			gTag = tag.(*gosn.Tag)
		} else {
			continue
		}

		if StringInSlice(gTag.GetUUID(), tagUUIDs, true) ||
			StringInSlice(gTag.Content.Title, tagTitles, true) {
			gTag.Deleted = true
			tagsToDelete = append(tagsToDelete, gTag)
		}
	}

	var eTagsToDelete gosn.EncryptedItems

	eTagsToDelete, err = tagsToDelete.Encrypt(session.Session, session.DefaultItemsKey)
	if err != nil {
		return 0, err
	}

	if err = cache.SaveEncryptedItems(so.DB, eTagsToDelete, true); err != nil {
		return
	}

	if len(tagsToDelete) > 0 {
		pii := cache.SyncInput{
			Session: session,
			Close:   true,
		}

		_, err = Sync(pii, true)
		if err != nil {
			return
		}
	}

	noDeleted = len(tagsToDelete)

	return noDeleted, err
}

type addTagsInput struct {
	session   *cache.Session
	tagTitles []string
	referers  []string
	replace   bool
}

type addTagsOutput struct {
	added    []string
	existing []string
	tags     gosn.Tags
}

func addTags(ati addTagsInput) (ato addTagsOutput, err error) {
	fmt.Printf("in addTags with: %+v\n", ati.tagTitles)

	// get all tags
	addTagsFilter := gosn.Filter{
		Type: "Tag",
	}

	filters := []gosn.Filter{addTagsFilter}

	addFilter := gosn.ItemFilters{
		MatchAny: true,
		Filters:  filters,
	}

	putItemsInput := cache.SyncInput{
		Session: ati.session,
	}

	var so cache.SyncOutput
	so, err = Sync(putItemsInput, true)
	if err != nil {
		return
	}

	var allPersistedItems cache.Items

	err = so.DB.All(&allPersistedItems)
	if err != nil {
		return
	}

	var items gosn.Items
	items, err = allPersistedItems.ToItems(ati.session)
	if err != nil {
		return
	}

	if err = so.DB.Close(); err != nil {
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

		newTag, _ := gosn.NewTag(tag, nil)

		tagsToAdd = append(tagsToAdd, newTag)
		ato.added = append(ato.added, tag)
	}

	if len(tagsToAdd) > 0 {
		fmt.Printf("GOT TAGS TO ADD: %+v\n", tagsToAdd)
		so, err = Sync(putItemsInput, true)
		if err != nil {
			return
		}

		var eTagsToAdd gosn.EncryptedItems

		eTagsToAdd, err = tagsToAdd.Encrypt(ati.session.Gosn())
		if err != nil {
			return
		}

		err = cache.SaveEncryptedItems(so.DB, eTagsToAdd, true)
		if err != nil {
			return
		}

		so, err = Sync(putItemsInput, true)
		if err != nil {
			return
		}

		err = so.DB.Close()
		if err != nil {
			return
		}
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
