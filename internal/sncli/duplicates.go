package sncli

import (
	"fmt"
	"sort"
	"time"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
)

// NoteSummary identifies a note in duplicate output.
type NoteSummary struct {
	UUID  string
	Title string
	// UpdatedAt is the note's last-updated timestamp, as the API reports it.
	UpdatedAt int64
	// UpdatedAtText is the same time as text, for display.
	UpdatedAtText string
	// Marked reports whether the note carries duplicate_of, meaning Standard
	// Notes created it by duplicating another note.
	Marked bool
	// Identical reports whether this note's title and text match the note
	// being kept in its set. It is only meaningful for notes being deleted:
	// when true, deleting the note loses nothing.
	Identical bool
}

// UpdatedDate returns the note's last-updated time as a date and time for
// display, or the raw value if it cannot be parsed.
func (n NoteSummary) UpdatedDate() string {
	t, err := time.Parse(timeLayout, n.UpdatedAtText)
	if err != nil {
		return n.UpdatedAtText
	}

	return t.Format("2006-01-02 15:04")
}

// DuplicateGroup is a note and the copies made from it, linked by
// duplicate_of. The most recently updated note is kept and the rest deleted,
// so that editing a copy after duplicating it does not lose that work.
type DuplicateGroup struct {
	Keep   NoteSummary
	Delete []NoteSummary
}

func summariseNote(note *items.Note) NoteSummary {
	return NoteSummary{
		UUID:          note.GetUUID(),
		Title:         note.Content.GetTitle(),
		UpdatedAt:     note.GetUpdatedAtTimestamp(),
		UpdatedAtText: note.GetUpdatedAt(),
		Marked:        note.GetDuplicateOf() != "",
	}
}

// sameContent reports whether two notes hold the same title and text.
func sameContent(a, b *items.Note) bool {
	if a == nil || b == nil {
		return false
	}

	return a.Content.GetTitle() == b.Content.GetTitle() && a.Content.GetText() == b.Content.GetText()
}

// newerThan reports whether a should be kept in preference to b. The most
// recently updated wins; on a tie an original is preferred over a copy, and
// then the lower UUID, so the choice is stable.
func newerThan(a, b NoteSummary) bool {
	switch {
	case a.UpdatedAt != b.UpdatedAt:
		return a.UpdatedAt > b.UpdatedAt
	case a.Marked != b.Marked:
		return !a.Marked
	default:
		return a.UUID < b.UUID
	}
}

// FindDuplicateGroups groups notes with the copies made from them and reports
// which to keep. It also returns copies whose original is no longer in the
// account: those are the only remaining copy of their content, so they are
// left alone.
func FindDuplicateGroups(in items.Items) (groups []DuplicateGroup, keptNoOriginal []NoteSummary) {
	live := make(map[string]*items.Note)

	for _, item := range in {
		if note, isNote := item.(*items.Note); isNote && !note.IsDeleted() {
			live[note.GetUUID()] = note
		}
	}

	// union-find over the live notes, joining each copy to the note it was
	// made from, so that a copy of a copy ends up in one group
	parent := make(map[string]string, len(live))

	var find func(string) string

	find = func(u string) string {
		if parent[u] != u {
			parent[u] = find(parent[u])
		}

		return parent[u]
	}

	for uuid := range live {
		parent[uuid] = uuid
	}

	for uuid, note := range live {
		original := note.GetDuplicateOf()
		if original == "" {
			continue
		}

		if _, ok := live[original]; !ok {
			keptNoOriginal = append(keptNoOriginal, summariseNote(note))

			continue
		}

		parent[find(uuid)] = find(original)
	}

	members := make(map[string][]NoteSummary)

	for uuid, note := range live {
		root := find(uuid)
		members[root] = append(members[root], summariseNote(note))
	}

	for _, group := range members {
		if len(group) < 2 {
			continue
		}

		keep := group[0]

		for _, candidate := range group[1:] {
			if newerThan(candidate, keep) {
				keep = candidate
			}
		}

		var toDelete []NoteSummary

		kept := live[keep.UUID]

		for _, candidate := range group {
			if candidate.UUID == keep.UUID {
				continue
			}

			candidate.Identical = sameContent(live[candidate.UUID], kept)
			toDelete = append(toDelete, candidate)
		}

		sort.Slice(toDelete, func(i, j int) bool { return toDelete[i].UUID < toDelete[j].UUID })

		groups = append(groups, DuplicateGroup{Keep: keep, Delete: toDelete})
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i].Keep.UUID < groups[j].Keep.UUID })
	sort.Slice(keptNoOriginal, func(i, j int) bool { return keptNoOriginal[i].UUID < keptNoOriginal[j].UUID })

	return groups, keptNoOriginal
}

// KeepOnlyIdentical drops notes whose content differs from the note being kept,
// returning the remaining groups and the notes left alone. A group with
// nothing left to delete is dropped.
func KeepOnlyIdentical(groups []DuplicateGroup) (filtered []DuplicateGroup, skipped []NoteSummary) {
	for _, group := range groups {
		var identical []NoteSummary

		for _, dup := range group.Delete {
			if dup.Identical {
				identical = append(identical, dup)

				continue
			}

			skipped = append(skipped, dup)
		}

		if len(identical) == 0 {
			continue
		}

		filtered = append(filtered, DuplicateGroup{Keep: group.Keep, Delete: identical})
	}

	return filtered, skipped
}

// DeleteDuplicateNotesConfig deletes the superseded notes in each set of
// duplicates.
type DeleteDuplicateNotesConfig struct {
	Session *cache.Session
	// DryRun reports what would be deleted without deleting anything.
	DryRun bool
	// IdenticalOnly restricts deletion to notes whose title and text match the
	// note being kept, leaving those that have diverged.
	IdenticalOnly bool
	Debug         bool
}

// DeleteDuplicateNotesOutput describes what was, or would be, deleted.
type DeleteDuplicateNotesOutput struct {
	// Groups holds each set of duplicates, with the note kept and those
	// deleted.
	Groups []DuplicateGroup
	// KeptNoOriginal holds copies left alone because the note they were made
	// from is no longer in the account.
	KeptNoOriginal []NoteSummary
	// KeptDiffering holds notes left alone under IdenticalOnly, because their
	// content differs from the note being kept.
	KeptDiffering []NoteSummary
}

// Deleted returns every note deleted, across all groups.
func (o DeleteDuplicateNotesOutput) Deleted() []NoteSummary {
	var all []NoteSummary

	for _, group := range o.Groups {
		all = append(all, group.Delete...)
	}

	return all
}

// Run finds duplicates and, unless this is a dry run, deletes all but the most
// recently updated note in each set.
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

	out.Groups, out.KeptNoOriginal = FindDuplicateGroups(all)

	if i.IdenticalOnly {
		out.Groups, out.KeptDiffering = KeepOnlyIdentical(out.Groups)
	}

	toDelete := out.Deleted()

	if i.DryRun || len(toDelete) == 0 {
		_ = i.Session.CacheDB.Close()

		return out, nil
	}

	var notesToDelete items.Notes

	for _, summary := range toDelete {
		note := byUUID[summary.UUID]
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
