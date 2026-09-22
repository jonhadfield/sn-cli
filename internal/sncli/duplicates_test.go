package sncli

import (
	"testing"

	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/stretchr/testify/require"
)

func testNote(uuid, title, duplicateOf string, updatedAt int64, deleted bool) *items.Note {
	note := &items.Note{
		ItemCommon: items.ItemCommon{
			UUID:               uuid,
			ContentType:        common.SNItemTypeNote,
			DuplicateOf:        duplicateOf,
			UpdatedAtTimestamp: updatedAt,
			Deleted:            deleted,
		},
	}
	note.Content.SetTitle(title)

	return note
}

func testNoteWithText(uuid, title, text, duplicateOf string, updatedAt int64) *items.Note {
	note := testNote(uuid, title, duplicateOf, updatedAt, false)
	note.Content.SetText(text)

	return note
}

func TestFindDuplicateGroupsKeepsNewerCopy(t *testing.T) {
	// the copy was edited after it was made, so it holds the latest work
	in := items.Items{
		testNote("original", "Shopping", "", 100, false),
		testNote("copy", "Shopping copy", "original", 200, false),
	}

	groups, kept := FindDuplicateGroups(in)

	require.Empty(t, kept)
	require.Len(t, groups, 1)
	require.Equal(t, "copy", groups[0].Keep.UUID)
	require.Len(t, groups[0].Delete, 1)
	require.Equal(t, "original", groups[0].Delete[0].UUID)
}

func TestFindDuplicateGroupsKeepsNewerOriginal(t *testing.T) {
	// the usual case: the copy was never touched after being made
	in := items.Items{
		testNote("original", "Shopping", "", 300, false),
		testNote("copy", "Shopping copy", "original", 200, false),
	}

	groups, _ := FindDuplicateGroups(in)

	require.Len(t, groups, 1)
	require.Equal(t, "original", groups[0].Keep.UUID)
	require.Equal(t, "copy", groups[0].Delete[0].UUID)
}

func TestFindDuplicateGroupsTiePrefersOriginal(t *testing.T) {
	// identical timestamps: keep the note that is not marked as a copy
	in := items.Items{
		testNote("original", "Shopping", "", 100, false),
		testNote("copy", "Shopping copy", "original", 100, false),
	}

	groups, _ := FindDuplicateGroups(in)

	require.Len(t, groups, 1)
	require.Equal(t, "original", groups[0].Keep.UUID)
}

// A copy whose original is gone is the only note holding that content, so it
// must not be deleted.
func TestFindDuplicateGroupsOriginalMissing(t *testing.T) {
	in := items.Items{
		testNote("copy", "Shopping copy", "gone", 100, false),
	}

	groups, kept := FindDuplicateGroups(in)

	require.Empty(t, groups)
	require.Len(t, kept, 1)
	require.Equal(t, "copy", kept[0].UUID)
}

func TestFindDuplicateGroupsOriginalDeleted(t *testing.T) {
	in := items.Items{
		testNote("original", "Shopping", "", 100, true),
		testNote("copy", "Shopping copy", "original", 200, false),
	}

	groups, kept := FindDuplicateGroups(in)

	require.Empty(t, groups, "a deleted original is not a live duplicate")
	require.Len(t, kept, 1)
}

func TestFindDuplicateGroupsSkipsDeletedCopies(t *testing.T) {
	in := items.Items{
		testNote("original", "Shopping", "", 100, false),
		testNote("copy", "Shopping copy", "original", 200, true),
	}

	groups, kept := FindDuplicateGroups(in)

	require.Empty(t, groups)
	require.Empty(t, kept)
}

func TestFindDuplicateGroupsNone(t *testing.T) {
	in := items.Items{
		testNote("a", "One", "", 100, false),
		testNote("b", "Two", "", 200, false),
	}

	groups, kept := FindDuplicateGroups(in)

	require.Empty(t, groups)
	require.Empty(t, kept)
}

// A copy of a copy belongs to one set, so only the newest of the three is
// kept.
func TestFindDuplicateGroupsChain(t *testing.T) {
	in := items.Items{
		testNote("original", "Shopping", "", 100, false),
		testNote("copy", "Shopping copy", "original", 400, false),
		testNote("copy2", "Shopping copy 2", "copy", 200, false),
	}

	groups, kept := FindDuplicateGroups(in)

	require.Empty(t, kept)
	require.Len(t, groups, 1)
	require.Equal(t, "copy", groups[0].Keep.UUID)
	require.Len(t, groups[0].Delete, 2)
	require.Equal(t, []string{"copy2", "original"},
		[]string{groups[0].Delete[0].UUID, groups[0].Delete[1].UUID})
}

// Two unrelated sets of duplicates stay separate.
func TestFindDuplicateGroupsSeparateSets(t *testing.T) {
	in := items.Items{
		testNote("a", "One", "", 100, false),
		testNote("a-copy", "One copy", "a", 200, false),
		testNote("b", "Two", "", 300, false),
		testNote("b-copy", "Two copy", "b", 100, false),
	}

	groups, _ := FindDuplicateGroups(in)

	require.Len(t, groups, 2)

	keep := map[string]string{}
	for _, g := range groups {
		keep[g.Keep.UUID] = g.Delete[0].UUID
	}

	require.Equal(t, map[string]string{"a-copy": "a", "b": "b-copy"}, keep)
}

func TestDeleteDuplicateNotesOutputDeleted(t *testing.T) {
	out := DeleteDuplicateNotesOutput{
		Groups: []DuplicateGroup{
			{Keep: NoteSummary{UUID: "a"}, Delete: []NoteSummary{{UUID: "a-copy"}}},
			{Keep: NoteSummary{UUID: "b"}, Delete: []NoteSummary{{UUID: "b-copy"}, {UUID: "b-copy2"}}},
		},
	}

	require.Len(t, out.Deleted(), 3)
}

func TestNoteSummaryUpdatedDate(t *testing.T) {
	require.Equal(t, "2026-09-21 14:30",
		NoteSummary{UpdatedAtText: "2026-09-21T14:30:00.000Z"}.UpdatedDate())

	// an unparseable value is shown as it came from the API
	require.Equal(t, "not a time", NoteSummary{UpdatedAtText: "not a time"}.UpdatedDate())
}

func TestFindDuplicateGroupsFlagsIdenticalContent(t *testing.T) {
	in := items.Items{
		testNoteWithText("original", "Shopping", "milk", "", 100),
		testNoteWithText("copy", "Shopping", "milk", "original", 200),
	}

	groups, _ := FindDuplicateGroups(in)

	require.Len(t, groups, 1)
	require.Equal(t, "copy", groups[0].Keep.UUID)
	require.True(t, groups[0].Delete[0].Identical)
}

func TestFindDuplicateGroupsFlagsDifferingText(t *testing.T) {
	in := items.Items{
		testNoteWithText("original", "Shopping", "milk", "", 100),
		testNoteWithText("copy", "Shopping", "milk and eggs", "original", 200),
	}

	groups, _ := FindDuplicateGroups(in)

	require.Len(t, groups, 1)
	require.False(t, groups[0].Delete[0].Identical)
}

// A note with the same text under a different title has not been left
// untouched, so it does not count as identical.
func TestFindDuplicateGroupsFlagsDifferingTitle(t *testing.T) {
	in := items.Items{
		testNoteWithText("original", "Shopping", "milk", "", 100),
		testNoteWithText("copy", "Shopping copy", "milk", "original", 200),
	}

	groups, _ := FindDuplicateGroups(in)

	require.Len(t, groups, 1)
	require.False(t, groups[0].Delete[0].Identical)
}

func TestKeepOnlyIdentical(t *testing.T) {
	groups := []DuplicateGroup{
		{
			Keep: NoteSummary{UUID: "a"},
			Delete: []NoteSummary{
				{UUID: "a-same", Identical: true},
				{UUID: "a-diff"},
			},
		},
		{
			// nothing identical, so the whole set is dropped
			Keep:   NoteSummary{UUID: "b"},
			Delete: []NoteSummary{{UUID: "b-diff"}},
		},
	}

	filtered, skipped := KeepOnlyIdentical(groups)

	require.Len(t, filtered, 1)
	require.Equal(t, "a", filtered[0].Keep.UUID)
	require.Len(t, filtered[0].Delete, 1)
	require.Equal(t, "a-same", filtered[0].Delete[0].UUID)

	require.Len(t, skipped, 2)
	require.Equal(t, []string{"a-diff", "b-diff"}, []string{skipped[0].UUID, skipped[1].UUID})
}

func TestKeepOnlyIdenticalNoGroups(t *testing.T) {
	filtered, skipped := KeepOnlyIdentical(nil)

	require.Empty(t, filtered)
	require.Empty(t, skipped)
}
