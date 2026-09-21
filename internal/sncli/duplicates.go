package sncli

import (
	"fmt"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

// DuplicateNote is a note that Standard Notes marked as a copy of another,
// by setting duplicate_of when the note was duplicated.
type DuplicateNote struct {
	UUID string
	// Title of the duplicate itself
	Title string
	// OriginalUUID is the note this one was copied from
	OriginalUUID string
	// OriginalPresent reports whether that original is still in the account.
	// When it is not, deleting the duplicate would lose the content, so the
	// duplicate is kept.
	OriginalPresent bool
}

// FindDuplicateNotes returns the notes marked as duplicates, and whether the
// note each was copied from still exists.
func FindDuplicateNotes(in items.Items) []DuplicateNote {
	live := make(map[string]struct{})

	for _, item := range in {
		if item.GetContentType() == common.SNItemTypeNote && !item.IsDeleted() {
			live[item.GetUUID()] = struct{}{}
		}
	}

	var duplicates []DuplicateNote

	for _, item := range in {
		note, isNote := item.(*items.Note)
		if !isNote || note.IsDeleted() || note.GetDuplicateOf() == "" {
			continue
		}

		original := note.GetDuplicateOf()
		_, originalPresent := live[original]

		duplicates = append(duplicates, DuplicateNote{
			UUID:            note.GetUUID(),
			Title:           note.Content.GetTitle(),
			OriginalUUID:    original,
			OriginalPresent: originalPresent,
		})
	}

	return duplicates
}

// DeleteDuplicateNotesConfig deletes notes marked as duplicates of another
// note.
type DeleteDuplicateNotesConfig struct {
	Session *cache.Session
	// DryRun reports what would be deleted without deleting anything.
	DryRun bool
	Debug  bool
}

// DeleteDuplicateNotesOutput describes what was, or would be, deleted.
type DeleteDuplicateNotesOutput struct {
	// Deleted holds the duplicates removed, or, for a dry run, those that
	// would be removed.
	Deleted []DuplicateNote
	// KeptOriginalMissing holds duplicates left alone because the note they
	// were copied from is no longer in the account.
	KeptOriginalMissing []DuplicateNote
}

// Run finds duplicates and, unless this is a dry run, deletes them.
func (i *DeleteDuplicateNotesConfig) Run() (DeleteDuplicateNotesOutput, error) {
	var out DeleteDuplicateNotesOutput

	so, err := Sync(cache.SyncInput{Session: i.Session}, true)
	if err != nil {
		return out, err
	}

	var allPersistedItems cache.Items

	if err = so.DB.All(&allPersistedItems); err != nil {
		return out, fmt.Errorf("getting items from db: %w", err)
	}

	var all items.Items

	if all, err = allPersistedItems.ToItems(i.Session); err != nil {
		return out, err
	}

	byUUID := make(map[string]*items.Note)

	for _, item := range all {
		if note, isNote := item.(*items.Note); isNote {
			byUUID[note.GetUUID()] = note
		}
	}

	for _, dup := range FindDuplicateNotes(all) {
		if dup.OriginalPresent {
			out.Deleted = append(out.Deleted, dup)

			continue
		}

		out.KeptOriginalMissing = append(out.KeptOriginalMissing, dup)
	}

	if i.DryRun || len(out.Deleted) == 0 {
		_ = i.Session.CacheDB.Close()

		return out, nil
	}

	var notesToDelete items.Notes

	for _, dup := range out.Deleted {
		note := byUUID[dup.UUID]
		if note == nil || note.GetContent() == nil {
			continue
		}

		note.Content.SetText("")
		note.SetDeleted(true)
		notesToDelete = append(notesToDelete, *note)
	}

	if err = cache.SaveNotes(i.Session, so.DB, notesToDelete, true); err != nil {
		return out, err
	}

	if _, err = Sync(cache.SyncInput{Session: i.Session, Close: true}, true); err != nil {
		return out, err
	}

	return out, nil
}
