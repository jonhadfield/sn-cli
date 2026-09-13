package sncli

import (
	"errors"
	"fmt"
	"sort"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

// EditorInfo describes an editor that can be set as the default for new notes,
// along with whether it is the one currently set.
type EditorInfo struct {
	Identifier string
	Name       string
	BuiltIn    bool
	Deprecated bool
	Current    bool
}

// ListEditorsConfig lists the editors available to the account.
type ListEditorsConfig struct {
	Session *cache.Session
	Debug   bool
}

// SetDefaultEditorConfig sets the editor used for new notes.
type SetDefaultEditorConfig struct {
	Session *cache.Session
	// Editor is the display name or identifier of the editor to set.
	Editor string
	// Force writes the identifier without checking it against the available
	// editors. It exists so an editor newer than this client's built-in list
	// can still be set.
	Force bool
	Debug bool
}

// getAllItems syncs and returns every decrypted item for the session.
func getAllItems(session *cache.Session) (items.Items, error) {
	so, err := Sync(cache.SyncInput{Session: session}, true)
	if err != nil {
		return nil, err
	}

	var persisted cache.Items
	if err = so.DB.All(&persisted); err != nil {
		return nil, err
	}

	if err = so.DB.Close(); err != nil {
		return nil, err
	}

	return persisted.ToItems(session)
}

// findUserPreferences returns the account's SN|UserPreferences item.
func findUserPreferences(its items.Items) (*items.UserPreferences, bool) {
	for _, i := range its {
		if i.GetContentType() != common.SNItemTypeUserPreferences || i.IsDeleted() {
			continue
		}

		if prefs, ok := i.(*items.UserPreferences); ok {
			return prefs, true
		}
	}

	return nil, false
}

// Run returns the available editors, sorted with the current default first and
// deprecated editors last, so the most useful choices lead.
func (i *ListEditorsConfig) Run() ([]EditorInfo, error) {
	its, err := getAllItems(i.Session)
	if err != nil {
		return nil, err
	}

	var current string
	if prefs, ok := findUserPreferences(its); ok {
		current, _ = prefs.Content.GetDefaultEditorIdentifier()
	}

	// An unset preference means Standard Notes uses plain text.
	if current == "" {
		current = items.EditorPlainText
	}

	available := items.AvailableEditors(its)

	out := make([]EditorInfo, 0, len(available))
	for _, e := range available {
		out = append(out, EditorInfo{
			Identifier: e.Identifier,
			Name:       e.Name,
			BuiltIn:    e.BuiltIn,
			Deprecated: e.Deprecated,
			Current:    e.Identifier == current,
		})
	}

	sort.SliceStable(out, func(a, b int) bool {
		if out[a].Current != out[b].Current {
			return out[a].Current
		}

		if out[a].Deprecated != out[b].Deprecated {
			return !out[a].Deprecated
		}

		return out[a].Name < out[b].Name
	})

	return out, nil
}

// Run resolves the requested editor and writes it to the account's
// preferences.
func (i *SetDefaultEditorConfig) Run() (EditorInfo, error) {
	if i.Editor == "" {
		return EditorInfo{}, errors.New("an editor name or identifier is required")
	}

	its, err := getAllItems(i.Session)
	if err != nil {
		return EditorInfo{}, err
	}

	resolved := items.Editor{Identifier: i.Editor}

	// Standard Notes silently falls back to plain text when the identifier
	// matches no editor, so resolve it here and report a typo rather than
	// writing a preference that quietly does nothing.
	if !i.Force {
		var found bool

		resolved, found = items.FindEditor(items.AvailableEditors(its), i.Editor)
		if !found {
			return EditorInfo{}, fmt.Errorf(
				"unknown editor %q: run 'sn editor list' to see the available editors, or pass --force to set the identifier anyway",
				i.Editor)
		}
	}

	prefs, ok := findUserPreferences(its)
	if !ok {
		return EditorInfo{}, errors.New("no user preferences found for this account; open Standard Notes once to create them")
	}

	prefs.Content.SetDefaultEditorIdentifier(resolved.Identifier)

	so, err := Sync(cache.SyncInput{Session: i.Session, Close: false}, true)
	if err != nil {
		return EditorInfo{}, err
	}

	if err = cache.SaveItems(i.Session, so.DB, items.Items{prefs}, false); err != nil {
		_ = so.DB.Close()

		return EditorInfo{}, err
	}

	if err = so.DB.Close(); err != nil {
		return EditorInfo{}, err
	}

	if _, err = Sync(cache.SyncInput{Session: i.Session, Close: true}, true); err != nil {
		return EditorInfo{}, err
	}

	return EditorInfo{
		Identifier: resolved.Identifier,
		Name:       resolved.Name,
		BuiltIn:    resolved.BuiltIn,
		Deprecated: resolved.Deprecated,
		Current:    true,
	}, nil
}

// DefaultEditorIdentifier returns the editor configured for new notes, falling
// back to plain text when the preference is unset, which is what Standard
// Notes itself does.
func DefaultEditorIdentifier(its items.Items) string {
	if prefs, ok := findUserPreferences(its); ok {
		if identifier, set := prefs.Content.GetDefaultEditorIdentifier(); set {
			return identifier
		}
	}

	return items.EditorPlainText
}

// defaultEditorForSession returns the account's default editor identifier for
// use when creating a note. Any failure to read the preference falls back to
// plain text rather than erroring: not being able to determine the preferred
// editor is not a reason to refuse to add a note.
func defaultEditorForSession(session *cache.Session) (string, error) {
	its, err := getAllItems(session)
	if err != nil {
		return items.EditorPlainText, nil //nolint:nilerr // a note is still worth adding
	}

	return DefaultEditorIdentifier(its), nil
}
