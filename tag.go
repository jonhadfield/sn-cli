package sncli

import (
	"strings"

	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

type tagNotesInput struct {
	session        *cache.Session
	matchTitle     string
	matchText      string
	matchTags      []string
	matchNoteUUIDs []string
	newTags        []string
	syncToken      string
}

// create tags if they don't exist
// get all notes and tags
func tagNotes(i tagNotesInput) (err error) {
	// create tags if they don't exist
	ati := addTagsInput{
		session:   i.session,
		tagTitles: i.newTags,
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
		if StringInSlice(note.UUID, i.matchNoteUUIDs, false) {
			typeUUIDs["Note"] = append(typeUUIDs["Note"], note.UUID)
		} else if strings.TrimSpace(i.matchTitle) != "" && strings.Contains(strings.ToLower(note.Content.GetTitle()), strings.ToLower(i.matchTitle)) {
			typeUUIDs["Note"] = append(typeUUIDs["Note"], note.UUID)
		} else if strings.TrimSpace(i.matchText) != "" && strings.Contains(strings.ToLower(note.Content.GetText()), strings.ToLower(i.matchText)) {
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

	if err = cache.SaveTags(so.DB, i.session, tagsToPush, true); err != nil {
		return
	}

	if len(tagsToPush) > 0 {
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
		Debug:   i.Debug,
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
		session:   i.Session,
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
		Debug:   i.Debug,
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
		Debug:   i.Debug,
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
	eTagsToDelete, err = tagsToDelete.Encrypt(session.Gosn())

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
}

type addTagsOutput struct {
	added    []string
	existing []string
}

func addTags(ati addTagsInput) (ato addTagsOutput, err error) {
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

		newTagContent := gosn.NewTagContent()
		newTag := gosn.NewTag()
		newTagContent.Title = tag
		newTag.Content = *newTagContent
		newTag.UUID = gosn.GenUUID()

		tagsToAdd = append(tagsToAdd, newTag)
		ato.added = append(ato.added, tag)
	}

	if len(tagsToAdd) > 0 {
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
