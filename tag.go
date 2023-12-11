package sncli

import (
	"strings"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
)

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
	getNotesFilter := items.Filter{
		Type: "Note",
	}
	getTagsFilter := items.Filter{
		Type: "Tag",
	}
	filters := []items.Filter{getNotesFilter, getTagsFilter}
	itemFilter := items.ItemFilters{
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

	var gItems items.Items

	gItems, err = allPersistedItems.ToItems(i.session)
	if err != nil {
		return
	}

	gItems.Filter(itemFilter)

	var allTags []*items.Tag

	var allNotes []*items.Note
	// create slices of notes and tags

	for _, item := range gItems {
		if item.IsDeleted() {
			continue
		}

		if item.GetContentType() == "Tag" {
			allTags = append(allTags, item.(*items.Tag))
		}

		if item.GetContentType() == "Note" {
			allNotes = append(allNotes, item.(*items.Note))
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
	var tagsToPush items.Tags

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
		session:   i.Session,
		replace:   i.Replace,
	}

	var ato addTagsOutput
	ato, err = addTags(ati)
	if err != nil {
		return AddTagsOutput{}, err
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

func (i *GetTagConfig) Run() (items items.Items, err error) {
	var so cache.SyncOutput

	si := cache.SyncInput{
		Session: i.Session,
	}

	so, err = Sync(si, true)
	if err != nil {
		return items, err
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
	deleteTagsFilter := items.Filter{
		Type: "Tag",
	}
	filters := []items.Filter{deleteTagsFilter}
	deleteFilter := items.ItemFilters{
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

	var tags items.Items

	// get items from db
	var allPersistedItems cache.Items

	err = so.DB.All(&allPersistedItems)
	if err != nil {
		return
	}

	var gItems items.Items

	gItems, err = allPersistedItems.ToItems(session)
	if err != nil {
		return
	}

	tags = gItems
	tags.Filter(deleteFilter)

	var tagsToDelete items.Items

	for _, tag := range tags {
		if tag.IsDeleted() {
			continue
		}

		var gTag *items.Tag
		if tag.GetContentType() == "Tag" {
			gTag = tag.(*items.Tag)
		} else {
			continue
		}

		if StringInSlice(gTag.GetUUID(), tagUUIDs, true) ||
			StringInSlice(gTag.Content.Title, tagTitles, true) {
			gTag.Deleted = true
			tagsToDelete = append(tagsToDelete, gTag)
		}
	}

	var eTagsToDelete items.EncryptedItems

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
	replace   bool
}

type addTagsOutput struct {
	added    []string
	existing []string
}

func addTags(ati addTagsInput) (ato addTagsOutput, err error) {
	// get all tags
	addTagsFilter := items.Filter{
		Type: "Tag",
	}

	filters := []items.Filter{addTagsFilter}

	addFilter := items.ItemFilters{
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

	var gItems items.Items
	gItems, err = allPersistedItems.ToItems(ati.session)
	if err != nil {
		return
	}

	if err = so.DB.Close(); err != nil {
		return
	}

	gItems.Filter(addFilter)

	var allTags items.Tags

	for _, item := range gItems {
		if item.IsDeleted() {
			continue
		}

		if item.GetContentType() == "Tag" {
			tag := item.(*items.Tag)
			allTags = append(allTags, *tag)
		}
	}

	var tagsToAdd items.Tags

	for _, tag := range ati.tagTitles {
		if tagExists(allTags, tag) {
			ato.existing = append(ato.existing, tag)
			continue
		}

		newTag, _ := items.NewTag(tag, nil)

		tagsToAdd = append(tagsToAdd, newTag)
		ato.added = append(ato.added, tag)
	}

	if len(tagsToAdd) > 0 {
		so, err = Sync(putItemsInput, true)
		if err != nil {
			return
		}

		var eTagsToAdd items.EncryptedItems

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

func upsertTagReferences(tag items.Tag, typeUUIDs map[string][]string) (items.Tag, bool) {
	// create item reference
	var newReferences []items.ItemReference

	var changed bool

	for k, v := range typeUUIDs {
		for _, ref := range v {
			if !referenceExists(tag, ref) {
				newReferences = append(newReferences, items.ItemReference{
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

func tagExists(existing []items.Tag, find string) bool {
	for _, tag := range existing {
		if tag.Content.GetTitle() == find {
			return true
		}
	}

	return false
}
