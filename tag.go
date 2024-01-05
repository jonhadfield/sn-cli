package sncli

import (
	"errors"
	"strings"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
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
		Type: common.SNItemTypeNote,
	}
	getTagsFilter := items.Filter{
		Type: common.SNItemTypeTag,
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

		if item.GetContentType() == common.SNItemTypeTag {
			allTags = append(allTags, item.(*items.Tag))
		}

		if item.GetContentType() == common.SNItemTypeNote {
			allNotes = append(allNotes, item.(*items.Note))
		}
	}

	typeUUIDs := make(map[string][]string)
	// loop through all notes and create a list of those that
	// match the note title or exist in note text
	for _, note := range allNotes {
		switch {
		case StringInSlice(note.UUID, i.matchNoteUUIDs, false):
			typeUUIDs[common.SNItemTypeNote] = append(typeUUIDs[common.SNItemTypeNote], note.UUID)
		case strings.TrimSpace(i.matchTitle) != "" && strings.Contains(strings.ToLower(note.Content.GetTitle()), strings.ToLower(i.matchTitle)):
			typeUUIDs[common.SNItemTypeNote] = append(typeUUIDs[common.SNItemTypeNote], note.UUID)
		case strings.TrimSpace(i.matchText) != "" && strings.Contains(strings.ToLower(note.Content.GetText()), strings.ToLower(i.matchText)):
			typeUUIDs[common.SNItemTypeNote] = append(typeUUIDs[common.SNItemTypeNote], note.UUID)
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
	ati := addTagsInput{
		tagTitles:  i.Tags,
		parent:     i.Parent,
		parentUUID: i.ParentUUID,
		session:    i.Session,
		replace:    i.Replace,
	}

	var ato addTagsOutput

	ato, err = addTags(ati)
	if err != nil {
		return AddTagsOutput{}, err
	}

	output.Added = ato.added
	output.Existing = ato.existing

	return output, err
}

func (i *GetItemsConfig) Run() (items items.Items, err error) {
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

	items.FilterAllTypes(i.Filters)
	return items, err
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
		Type: common.SNItemTypeTag,
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
	if len(tags) == 0 {
		return 0, nil
	}

	var tagsToDelete items.Items

	for _, tag := range tags {
		if tag.IsDeleted() {
			continue
		}

		var gTag *items.Tag
		if tag.GetContentType() == common.SNItemTypeTag {
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

	if len(eTagsToDelete) == 0 {
		return 0, nil
	}

	if err = cache.SaveEncryptedItems(so.DB, eTagsToDelete, true); err != nil {
		return
	}

	pii := cache.SyncInput{
		Session: session,
		Close:   true,
	}

	_, err = Sync(pii, true)
	if err != nil {
		return
	}

	noDeleted = len(tagsToDelete)

	return noDeleted, nil
}

type addTagsInput struct {
	session    *cache.Session
	tagTitles  []string
	parent     string
	parentUUID string
	replace    bool
}

type addTagsOutput struct {
	added    []string
	existing []string
}

func addTags(ati addTagsInput) (ato addTagsOutput, err error) {
	// get all tags
	addTagsFilter := items.Filter{
		Type: common.SNItemTypeTag,
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
	var parentRef items.ItemReferences

	for _, item := range gItems {
		if item.IsDeleted() {
			continue
		}

		if item.GetContentType() == common.SNItemTypeTag {
			tag := item.(*items.Tag)
			allTags = append(allTags, *tag)
			if tag.Content.GetTitle() == ati.parent || tag.GetUUID() == ati.parentUUID {
				if parentRef != nil {
					return ato, errors.New("multiple parent tags found, specify by UUID")
				}
				itemRef := items.ItemReference{
					UUID:          tag.GetUUID(),
					ContentType:   common.SNItemTypeTag,
					ReferenceType: "TagToParentTag",
				}
				parentRef = items.ItemReferences{itemRef}
			}
		}
	}

	if ati.parent != "" && len(parentRef) == 0 {
		return ato, errors.New("parent tag not found by title")
	}

	if ati.parentUUID != "" && len(parentRef) == 0 {
		return ato, errors.New("parent tag not found by UUID")
	}

	var tagsToAdd items.Tags

	for _, tag := range ati.tagTitles {
		if tagExists(allTags, tag) {
			ato.existing = append(ato.existing, tag)
			continue
		}

		newTag, _ := items.NewTag(tag, parentRef)

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
