package sncli

import (
	"fmt"
	"slices"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
)

func conflictedWarning([]items.Tasklist) string {
	if len(items.Tasklists{}) > 0 {
		return color.Yellow.Sprintf("%d conflicted versions", len(items.Tasklists{}))
	}

	return "-"
}

func (ci *ListChecklistsInput) Run() (err error) {
	checklists, err := getChecklists(ci.Session)
	if err != nil {
		return err
	}

	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Title"},
			{Align: simpletable.AlignCenter, Text: "Last Updated"},
			{Align: simpletable.AlignCenter, Text: "UUID"},
			{Align: simpletable.AlignCenter, Text: "Issues"},
		},
	}

	for _, row := range checklists {
		r := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", row.Title)},
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", row.UpdatedAt.String())},
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", row.UUID)},
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", conflictedWarning(row.Duplicates))},
		}

		table.Body.Cells = append(table.Body.Cells, r)
	}

	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())

	return nil
}

// construct a map of duplicates.
func getChecklistsDuplicatesMap(checklistNotes items.Notes) (map[string][]items.Tasklist, error) {
	duplicates := make(map[string][]items.Tasklist)

	for x := range checklistNotes {
		if checklistNotes[x].DuplicateOf != "" {
			// checklist is a duplicate
			// get the checklist content
			cl, err := checklistNotes[x].Content.ToTaskList()
			if err != nil {
				return map[string][]items.Tasklist{}, err
			}

			// skip trashed content
			if cl.Trashed {
				continue
			}

			cl.UUID = checklistNotes[x].UUID
			cl.UpdatedAt, err = time.Parse(timeLayout, checklistNotes[x].UpdatedAt)
			if err != nil {
				return map[string][]items.Tasklist{}, err
			}

			duplicates[checklistNotes[x].DuplicateOf] = append(duplicates[checklistNotes[x].DuplicateOf], cl)
		}
	}

	return duplicates, nil
}

func getChecklists(sess *cache.Session) (items.Tasklists, error) {
	var so cache.SyncOutput

	so, err := Sync(cache.SyncInput{
		Session: sess,
	}, true)
	if err != nil {
		return items.Tasklists{}, err
	}

	var allPersistedItems cache.Items

	if err = so.DB.All(&allPersistedItems); err != nil {
		return items.Tasklists{}, err
	}

	allItemUUIDs := allPersistedItems.UUIDs()

	var gitems items.Items
	gitems, err = allPersistedItems.ToItems(sess)
	if err != nil {
		return items.Tasklists{}, err
	}

	gitems.Filter(items.ItemFilters{
		Filters: []items.Filter{
			{
				Type:       "Note",
				Key:        "editor",
				Comparison: "==",
				Value:      "com.sncommunity.advanced-checklist",
			},
		},
	})

	var checklists items.Tasklists
	checklistNotes := gitems.Notes()

	duplicatesMap, err := getChecklistsDuplicatesMap(checklistNotes)
	// strip any duplicated items that no longer exist
	for k := range duplicatesMap {
		if !slices.Contains(allItemUUIDs, k) {
			delete(duplicatesMap, k)
		}
	}

	// second pass to get all non-deleted and non-trashed checklists
	for x := range checklistNotes {
		// strip deleted and trashed
		if checklistNotes[x].Deleted || checklistNotes[x].Content.Trashed != nil && *checklistNotes[x].Content.Trashed {
			continue
		}

		var cl items.Tasklist
		cl, err = checklistNotes[x].Content.ToTaskList()
		if err != nil {
			return items.Tasklists{}, err
		}

		cl.UUID = checklistNotes[x].UUID
		cl.UpdatedAt, err = time.Parse(timeLayout, checklistNotes[x].UpdatedAt)
		if err != nil {
			return items.Tasklists{}, err
		}

		cl.Duplicates = duplicatesMap[checklistNotes[x].UUID]

		checklists = append(checklists, cl)
	}

	return checklists, nil
}
