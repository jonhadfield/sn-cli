package sncli

import (
	"testing"

	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/stretchr/testify/require"
)

func testNote(uuid, title, duplicateOf string, deleted bool) *items.Note {
	note := &items.Note{
		ItemCommon: items.ItemCommon{
			UUID:        uuid,
			ContentType: common.SNItemTypeNote,
			DuplicateOf: duplicateOf,
			Deleted:     deleted,
		},
	}
	note.Content.SetTitle(title)

	return note
}

func TestFindDuplicateNotesMarkedOnly(t *testing.T) {
	in := items.Items{
		testNote("original", "Shopping", "", false),
		testNote("copy", "Shopping copy", "original", false),
		testNote("unrelated", "Ideas", "", false),
	}

	dups := FindDuplicateNotes(in)

	require.Len(t, dups, 1)
	require.Equal(t, "copy", dups[0].UUID)
	require.Equal(t, "Shopping copy", dups[0].Title)
	require.Equal(t, "original", dups[0].OriginalUUID)
	require.True(t, dups[0].OriginalPresent)
}

// A duplicate whose original is gone is the only copy of that content left,
// so it must not be treated as safe to delete.
func TestFindDuplicateNotesOriginalMissing(t *testing.T) {
	in := items.Items{
		testNote("copy", "Shopping copy", "gone", false),
	}

	dups := FindDuplicateNotes(in)

	require.Len(t, dups, 1)
	require.False(t, dups[0].OriginalPresent)
}

func TestFindDuplicateNotesOriginalDeleted(t *testing.T) {
	in := items.Items{
		testNote("original", "Shopping", "", true),
		testNote("copy", "Shopping copy", "original", false),
	}

	dups := FindDuplicateNotes(in)

	require.Len(t, dups, 1)
	require.False(t, dups[0].OriginalPresent, "a deleted original should not count as present")
}

func TestFindDuplicateNotesSkipsDeletedDuplicates(t *testing.T) {
	in := items.Items{
		testNote("original", "Shopping", "", false),
		testNote("copy", "Shopping copy", "original", true),
	}

	require.Empty(t, FindDuplicateNotes(in))
}

func TestFindDuplicateNotesNone(t *testing.T) {
	in := items.Items{
		testNote("a", "One", "", false),
		testNote("b", "Two", "", false),
	}

	require.Empty(t, FindDuplicateNotes(in))
}

func TestFindDuplicateNotesChain(t *testing.T) {
	// a copy of a copy: both are duplicates, and both originals are present
	in := items.Items{
		testNote("original", "Shopping", "", false),
		testNote("copy", "Shopping copy", "original", false),
		testNote("copy2", "Shopping copy 2", "copy", false),
	}

	dups := FindDuplicateNotes(in)

	require.Len(t, dups, 2)

	for _, dup := range dups {
		require.True(t, dup.OriginalPresent)
	}
}
