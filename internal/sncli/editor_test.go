package sncli

import (
	"testing"

	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/stretchr/testify/require"
)

func newTestUserPreferences(defaultEditor string) *items.UserPreferences {
	prefs := items.NewUserPreferences()
	prefs.ContentType = common.SNItemTypeUserPreferences
	prefs.Content = *items.NewUserPreferencesContent()

	if defaultEditor != "" {
		prefs.Content.SetDefaultEditorIdentifier(defaultEditor)
	}

	return &prefs
}

func newTestNote(t *testing.T, title string) *items.Note {
	t.Helper()

	note, err := items.NewNote(title, "text", nil)
	require.NoError(t, err)

	return &note
}

func TestFindUserPreferences(t *testing.T) {
	t.Parallel()

	prefs := newTestUserPreferences("com.standardnotes.super-editor")
	its := items.Items{newTestNote(t, "a note"), prefs}

	found, ok := findUserPreferences(its)
	require.True(t, ok)
	require.Equal(t, prefs.GetUUID(), found.GetUUID())
}

func TestFindUserPreferencesAbsent(t *testing.T) {
	t.Parallel()

	_, ok := findUserPreferences(items.Items{newTestNote(t, "a note")})
	require.False(t, ok, "an account with no preferences item should report absence, not panic")
}

func TestFindUserPreferencesSkipsDeleted(t *testing.T) {
	t.Parallel()

	deleted := newTestUserPreferences("com.standardnotes.super-editor")
	deleted.Deleted = true

	_, ok := findUserPreferences(items.Items{deleted})
	require.False(t, ok, "a deleted preferences item must not be used")
}

func TestDefaultEditorIdentifier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		its  items.Items
		want string
	}{
		{
			name: "preference set",
			its:  items.Items{newTestUserPreferences("com.standardnotes.super-editor")},
			want: "com.standardnotes.super-editor",
		},
		{
			name: "preference unset falls back to plain text",
			its:  items.Items{newTestUserPreferences("")},
			want: items.EditorPlainText,
		},
		{
			name: "no preferences item falls back to plain text",
			its:  items.Items{},
			want: items.EditorPlainText,
		},
		{
			name: "nil items falls back to plain text",
			its:  nil,
			want: items.EditorPlainText,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, DefaultEditorIdentifier(tc.its))
		})
	}
}

// TestDefaultEditorIdentifierMatchesStandardNotesFallback documents that an
// unresolvable identifier is returned as-is rather than corrected. Standard
// Notes itself falls back to plain text when rendering, so the stored value
// stays visible here for diagnosis instead of being silently rewritten.
func TestDefaultEditorIdentifierReturnsUnknownValueAsIs(t *testing.T) {
	t.Parallel()

	its := items.Items{newTestUserPreferences("com.example.no-such-editor")}

	require.Equal(t, "com.example.no-such-editor", DefaultEditorIdentifier(its))
}
