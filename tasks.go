package sncli

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/asdine/storm/v3/q"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/pterm/pterm"
)

const (
	txtOrderingLastUpdated = "last-updated"
	defaultMaxLength       = 80
	selectListHeight       = 10
)

type ListTasklistsInput struct {
	Session  *cache.Session
	Ordering string
	Debug    bool
}

func outputTime(updated time.Time, created time.Time) string {
	updatedAt := humanize.Time(updated)
	if updated.IsZero() {
		updatedAt = humanize.Time(created)
	}

	return updatedAt
}

func (ci *ListTasklistsInput) Run() error {
	stdLists, advLists, err := getAllLists(ci.Session)
	if err != nil {
		return err
	}

	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: color.Bold.Text("title")},
			{Align: simpletable.AlignCenter, Text: color.Bold.Text("type")},
			{Align: simpletable.AlignCenter, Text: color.Bold.Text("updated")},
			{Align: simpletable.AlignCenter, Text: color.Bold.Text("uuid")},
		},
	}

	if ci.Ordering == txtOrderingLastUpdated {
		advLists.Sort()
	}

	// get advanced checklist rows
	for _, row := range stdLists {
		r := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: row.Title},
			{Align: simpletable.AlignLeft, Text: "std"},
			{Align: simpletable.AlignLeft, Text: outputTime(row.UpdatedAt, time.Time{})},
			// {Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", taskListsConflictedWarning(row.Duplicates))},
			{Align: simpletable.AlignLeft, Text: row.UUID},
		}

		table.Body.Cells = append(table.Body.Cells, r)
	}

	// get advanced checklist rows
	for _, row := range advLists {
		r := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: row.Title},
			{Align: simpletable.AlignLeft, Text: "adv"},
			{Align: simpletable.AlignLeft, Text: outputTime(row.UpdatedAt, time.Time{})},
			// {Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", conflictedWarning(row.Duplicates))},
			{Align: simpletable.AlignLeft, Text: row.UUID},
		}

		table.Body.Cells = append(table.Body.Cells, r)
	}

	table.SetStyle(simpletable.StyleRounded)
	fmt.Println(table.String())

	return nil
}

func itemsToListNotes(sess *cache.Session, agitems cache.Items, listType string) (items.Notes, error) {
	gis, err := agitems.ToItems(sess)
	if err != nil {
		return nil, err
	}

	gis.Filter(items.ItemFilters{
		Filters: []items.Filter{
			{
				Type:       common.SNItemTypeNote,
				Key:        "editor",
				Comparison: "==",
				Value:      listType,
			},
		},
	})

	return gis.Notes(), nil
}

func getTasklists(sess *cache.Session, cacheItems cache.Items) (items.Tasklists, error) {
	allItemUUIDs := cacheItems.UUIDs()

	var gitems items.Items

	gitems, err := cacheItems.ToItems(sess)
	if err != nil {
		return items.Tasklists{}, err
	}

	gitems.Filter(items.ItemFilters{
		Filters: []items.Filter{
			{
				Type:       "Note",
				Key:        "editor",
				Comparison: "==",
				Value:      items.SimpleTaskEditorNoteType,
			},
		},
	})

	var checklists items.Tasklists

	checklistNotes := gitems.Notes()

	duplicatesMap, err := getTasklistsDuplicatesMap(checklistNotes)
	if err != nil {
		return nil, err
	}

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

func getAllLists(sess *cache.Session) (items.Tasklists, items.AdvancedChecklists, error) {
	var so cache.SyncOutput

	so, err := Sync(cache.SyncInput{
		Session: sess,
	}, true)
	if err != nil {
		return nil, nil, err
	}

	defer so.DB.Close()

	var cacheItems cache.Items

	if err = so.DB.All(&cacheItems); err != nil {
		return nil, nil, err
	}

	std, err := getTasklists(sess, cacheItems)
	if err != nil {
		return nil, nil, err
	}

	adv, err := getAdvancedChecklists(sess, cacheItems)
	if err != nil {
		return nil, nil, err
	}

	return std, adv, nil
}

func getAllMatchingLists(sess *cache.Session, title, uuid string) (items.Tasklists, items.AdvancedChecklists, error) {
	var so cache.SyncOutput

	so, err := Sync(cache.SyncInput{
		Session: sess,
	}, true)
	if err != nil {
		return nil, nil, err
	}

	var cacheItems cache.Items

	if err = so.DB.All(&cacheItems); err != nil {
		return nil, nil, err
	}

	allStd, err := getTasklists(sess, cacheItems)
	if err != nil {
		return nil, nil, err
	}

	var std items.Tasklists

	for x := range allStd {
		if allStd[x].Title == title || allStd[x].UUID == uuid {
			std = append(std, allStd[x])
		}
	}

	allAdv, err := getAdvancedChecklists(sess, cacheItems)
	if err != nil {
		return nil, nil, err
	}

	var adv items.AdvancedChecklists

	for x := range allAdv {
		if allAdv[x].Title == title || allAdv[x].UUID == uuid {
			adv = append(adv, allAdv[x])
		}
	}

	if len(std) == 0 && len(adv) == 0 {
		return nil, nil, errors.New("list not found")
	}

	return std, adv, nil
}

func getAdvancedChecklists(sess *cache.Session, cacheItems cache.Items) (items.AdvancedChecklists, error) {
	allItemUUIDs := cacheItems.UUIDs()

	var checklists items.AdvancedChecklists

	listNotes, err := itemsToListNotes(sess, cacheItems, items.AdvancedChecklistNoteType)
	if err != nil {
		return nil, err
	}

	duplicatesMap, err := getAdvancedChecklistsDuplicatesMap(listNotes)
	if err != nil {
		return nil, err
	}
	// strip any duplicated items that no longer exist
	for k := range duplicatesMap {
		if !slices.Contains(allItemUUIDs, k) {
			delete(duplicatesMap, k)
		}
	}

	// second pass to get all non-deleted and non-trashed checklists
	for x := range listNotes {
		// strip deleted and trashed
		if listNotes[x].Deleted || listNotes[x].Content.Trashed != nil && *listNotes[x].Content.Trashed {
			continue
		}

		var cl items.AdvancedChecklist

		cl, err = listNotes[x].Content.ToAdvancedCheckList()
		if err != nil {
			return items.AdvancedChecklists{}, err
		}

		cl.UUID = listNotes[x].UUID

		cl.UpdatedAt, err = time.Parse(timeLayout, listNotes[x].UpdatedAt)
		if err != nil {
			return items.AdvancedChecklists{}, err
		}

		cl.Duplicates = duplicatesMap[listNotes[x].UUID]

		checklists = append(checklists, cl)
	}

	return checklists, nil
}

func taskListsConflictedWarning(tasklists []items.Tasklist) string {
	if len(tasklists) > 0 {
		return color.Yellow.Sprintf("%d conflicted versions", len(tasklists))
	}

	return "-"
}

// construct a map of duplicates.
func getAdvancedChecklistsDuplicatesMap(checklistNotes items.Notes) (map[string][]items.AdvancedChecklist, error) {
	duplicates := make(map[string][]items.AdvancedChecklist)

	for x := range checklistNotes {
		if checklistNotes[x].DuplicateOf == "" {
			continue
		}
		// checklist is a duplicate
		// get the checklist content
		cl, err := checklistNotes[x].Content.ToAdvancedCheckList()
		if err != nil {
			return map[string][]items.AdvancedChecklist{}, err
		}

		// skip trashed content
		if cl.Trashed {
			continue
		}

		cl.UUID = checklistNotes[x].UUID

		cl.UpdatedAt, err = time.Parse(timeLayout, checklistNotes[x].UpdatedAt)
		if err != nil {
			return map[string][]items.AdvancedChecklist{}, err
		}

		duplicates[checklistNotes[x].DuplicateOf] = append(duplicates[checklistNotes[x].DuplicateOf], cl)
	}

	return duplicates, nil
}

// construct a map of duplicates.
func getTasklistsDuplicatesMap(checklistNotes items.Notes) (map[string][]items.Tasklist, error) {
	duplicates := make(map[string][]items.Tasklist)

	for x := range checklistNotes {
		if checklistNotes[x].DuplicateOf == "" {
			continue
		}
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

	return duplicates, nil
}

type AddAdvancedChecklistTaskInput struct {
	Session  *cache.Session
	Debug    bool
	UUID     string
	Title    string
	Group    string
	Tasklist string
}

type AddTaskInput struct {
	Session  *cache.Session
	Debug    bool
	Title    string
	UUID     string
	Tasklist string
}

func (ci *AddTaskInput) Run() error {
	tasklist, err := getTasklist(ci.Session, ci.Tasklist, ci.UUID)
	if err != nil {
		return err
	}

	// add task to the tasklist
	if err = tasklist.AddTask(ci.Title); err != nil {
		return err
	}

	// get corresponding note
	note, err := getNoteByUUID(ci.Session, tasklist.UUID)
	if err != nil {
		return err
	}

	clnt := items.TasksToNoteText(tasklist.Tasks)

	now := time.Now().UTC()
	note.Content.SetUpdateTime(now)
	note.Content.SetText(clnt)

	// save note to db
	si := cache.SyncInput{
		Session: ci.Session,
		Close:   false,
	}

	_, err = cache.Sync(si)
	if err != nil {
		return err
	}

	ci.Session.CacheDB = si.CacheDB

	if err = cache.SaveNotes(ci.Session, ci.Session.CacheDB, items.Notes{note}, true); err != nil {
		return err
	}

	si = cache.SyncInput{
		Session: ci.Session,
		Close:   true,
	}
	if _, err = cache.Sync(si); err != nil {
		return err
	}
	return nil
}

func (ci *AddAdvancedChecklistTaskInput) Run() error {
	// sync to get db
	_, err := Sync(cache.SyncInput{
		Session: ci.Session,
	}, true)
	if err != nil {
		return err
	}

	// get matching note
	notes, err := getNotesByTitleUUID(ci.Session, ci.Tasklist, ci.UUID, items.AdvancedChecklistNoteType)
	if err != nil {
		return err
	}

	var cl items.AdvancedChecklist

	cl, err = notes[0].Content.ToAdvancedCheckList()
	if err != nil {
		return err
	}

	// add task to the checklist
	if err = cl.AddTask(ci.Group, ci.Title); err != nil {
		return err
	}

	taskNoteText := items.AdvancedCheckListToNoteText(cl)

	now := time.Now().UTC()
	notes[0].Content.SetUpdateTime(now)
	notes[0].Content.SetText(taskNoteText)
	notes[0].Content.SetUpdateTime(now)

	// save note to db
	si := cache.SyncInput{
		Session: ci.Session,
		Close:   false,
	}

	if err = cache.SaveNotes(ci.Session, ci.Session.CacheDB, items.Notes{notes[0]}, true); err != nil {
		return err
	}

	if _, err = cache.Sync(si); err != nil {
		return err
	}

	return nil
}

func getNoteByUUID(sess *cache.Session, uuid string) (items.Note, error) {
	if sess.CacheDBPath == "" {
		return items.Note{}, errors.New("CacheDBPath missing from session")
	}

	if uuid == "" {
		return items.Note{}, errors.New("uuid not supplied")
	}

	var so cache.SyncOutput

	si := cache.SyncInput{
		Session: sess,
		Close:   false,
	}

	so, err := cache.Sync(si)
	if err != nil {
		return items.Note{}, err
	}

	defer func() {
		_ = so.DB.Close()
	}()

	var encNotes cache.Items

	query := so.DB.Select(q.And(q.Eq("UUID", uuid), q.Eq("Deleted", false)))
	if err = query.Find(&encNotes); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return items.Note{}, fmt.Errorf("could not find note with inUUID %s", uuid)
		}
		return items.Note{}, err
	}

	rawEncItems, err := encNotes.ToItems(sess)
	if err != nil {
		return items.Note{}, err
	}

	return *rawEncItems[0].(*items.Note), nil
}

func getNotesByTitleUUID(sess *cache.Session, title string, uuid string, editor string) (items.Notes, error) {
	if title == "" && uuid == "" {
		return nil, errors.New("title or uuid required")
	}

	var allPersistedItems cache.Items

	if err := sess.CacheDB.All(&allPersistedItems); err != nil {
		return nil, err
	}

	var gitems items.Items

	gitems, err := allPersistedItems.ToItems(sess)
	if err != nil {
		return nil, err
	}

	filters := []items.Filter{
		{
			Type:       common.SNItemTypeNote,
			Key:        "editor",
			Comparison: "==",
			Value:      editor,
		},
	}

	if title != "" {
		filters = append(filters, items.Filter{
			Type:       common.SNItemTypeNote,
			Key:        "title",
			Comparison: "==",
			Value:      title,
		})
	}

	if uuid != "" {
		filters = append(filters, items.Filter{
			Type:       common.SNItemTypeNote,
			Key:        "uuid",
			Comparison: "==",
			Value:      uuid,
		})
	}

	gitems.Filter(items.ItemFilters{
		Filters: filters,
	})

	tasklistNotes := gitems.Notes()
	if len(tasklistNotes) == 0 {
		return nil, errors.New("list not found")
	}

	if len(tasklistNotes) > 1 {
		return nil, fmt.Errorf("%d lists found with same title/uuid", len(tasklistNotes))
	}

	return tasklistNotes, nil
}

func getTasklist(sess *cache.Session, title string, uuid string) (items.Tasklist, error) {
	if title == "" && uuid == "" {
		return items.Tasklist{}, errors.New("title and/or uuid must be specified")
	}

	var so cache.SyncOutput

	so, err := Sync(cache.SyncInput{
		Session: sess,
	}, true)
	if err != nil {
		return items.Tasklist{}, err
	}

	defer func() {
		_ = so.DB.Close()
	}()

	var allPersistedItems cache.Items

	if err = so.DB.All(&allPersistedItems); err != nil {
		return items.Tasklist{}, err
	}

	allItemUUIDs := allPersistedItems.UUIDs()

	var gitems items.Items

	gitems, err = allPersistedItems.ToItems(sess)
	if err != nil {
		return items.Tasklist{}, err
	}

	filters := []items.Filter{
		{
			Type:       common.SNItemTypeNote,
			Key:        "editor",
			Comparison: "==",
			Value:      items.SimpleTaskEditorNoteType,
		},
	}

	if title != "" {
		filters = append(filters, items.Filter{
			Type:       common.SNItemTypeNote,
			Key:        "title",
			Comparison: "==",
			Value:      title,
		})
	}

	if uuid != "" {
		filters = append(filters, items.Filter{
			Type:       common.SNItemTypeNote,
			Key:        "uuid",
			Comparison: "==",
			Value:      uuid,
		})
	}

	gitems.Filter(items.ItemFilters{
		Filters: filters,
	})

	var tasklists items.Tasklists

	tasklistNotes := gitems.Notes()

	duplicatesMap, err := getTasklistsDuplicatesMap(tasklistNotes)
	// strip any duplicated items that no longer exist
	for k := range duplicatesMap {
		if !slices.Contains(allItemUUIDs, k) {
			delete(duplicatesMap, k)
		}
	}

	// second pass to get all non-deleted and non-trashed checklists
	for x := range tasklistNotes {
		// strip deleted and trashed
		if tasklistNotes[x].Deleted || tasklistNotes[x].Content.Trashed != nil && *tasklistNotes[x].Content.Trashed {
			continue
		}

		var cl items.Tasklist

		cl, err = tasklistNotes[x].Content.ToTaskList()
		if err != nil {
			return items.Tasklist{}, err
		}

		cl.UUID = tasklistNotes[x].UUID

		cl.UpdatedAt, err = time.Parse(timeLayout, tasklistNotes[x].UpdatedAt)
		if err != nil {
			return items.Tasklist{}, err
		}

		cl.Duplicates = duplicatesMap[tasklistNotes[x].UUID]

		tasklists = append(tasklists, cl)
	}

	numTasklists := len(tasklists)
	if numTasklists == 0 {
		return items.Tasklist{}, errors.New("tasklist not found")
	}

	if numTasklists > 1 {
		return items.Tasklist{}, errors.New("duplicate tasklists found")
	}

	return tasklists[0], nil
}

type ShowTasklistInput struct {
	Session       *cache.Session
	Debug         bool
	Group         string
	Title         string
	UUID          string
	ShowCompleted bool
	Ordering      string
}

func outputChars(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}

	return s
}

// func outputTime(updated time.Time, created time.Time) string {
// 	updatedAt := humanize.Time(updated)
// 	if updated.IsZero() {
// 		updatedAt = humanize.Time(created)
// 	}
//
// 	return updatedAt
// }

func getShowGroupTaskRows(name string, tasks items.AdvancedChecklistTasks, completed bool) [][]*simpletable.Cell {
	var rowsCells [][]*simpletable.Cell

	var rowCells []*simpletable.Cell

	var nameOutput bool

	for y, task := range tasks {
		if !nameOutput {
			// add group row
			// adding rowCells to output
			rowCells = []*simpletable.Cell{
				// {Align: simpletable.AlignLeft, Text: ""},
				{Span: 2, Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("%s", name)},
				// {Align: simpletable.AlignLeft, Text: ""},
				{Align: simpletable.AlignLeft, Text: ""},
			}

			// output completed column if we're showing completed tasks
			if completed {
				rowCells = append(rowCells, &simpletable.Cell{
					Align: simpletable.AlignLeft, Text: boolToText(task.Completed, "yes", "no"),
				})
			}

			rowsCells = append(rowsCells, rowCells)
			nameOutput = true
		}

		// adding rowCells to output
		rowCells = []*simpletable.Cell{
			{Align: simpletable.AlignRight, Text: strconv.Itoa(y + 1)},
			{Align: simpletable.AlignLeft, Text: outputChars(task.Description, defaultMaxLength)},
			{Align: simpletable.AlignLeft, Text: " " + outputTime(task.UpdatedAt, task.CreatedAt) + " "},
		}

		// output completed column if we're showing completed tasks
		if completed {
			rowCells = append(rowCells, &simpletable.Cell{
				Align: simpletable.AlignLeft, Text: boolToText(task.Completed, "yes", "no"),
			})
		}

		rowsCells = append(rowsCells, rowCells)
	}

	return rowsCells
}

func getShowGroupsTaskRows(group items.AdvancedChecklistGroup, showCompleted bool) [][]*simpletable.Cell {
	var rowsCells [][]*simpletable.Cell

	// output open and then completed tasks (if specified)
	seqs := []bool{false}
	if showCompleted {
		seqs = append(seqs, true)
	}

	// TODO: output empty groups
	for x := range seqs {
		// filter complete/incomplete tasks
		filtered := filterAdvancedChecklistTasks(group.Tasks, seqs[x])
		taskRows := getShowGroupTaskRows(group.Name, filtered, showCompleted)
		rowsCells = append(rowsCells, taskRows...)
	}

	return rowsCells
}

func filterAdvancedChecklistTasks(tasks items.AdvancedChecklistTasks, completed bool) items.AdvancedChecklistTasks {
	var filtered items.AdvancedChecklistTasks

	for x := range tasks {
		if tasks[x].Completed == completed {
			filtered = append(filtered, tasks[x])
		}
	}

	return filtered
}

func countCompletedTasks(group *items.AdvancedChecklistGroup) int {
	completedTaskCount := 0

	for taskIndex := range group.Tasks {
		if group.Tasks[taskIndex].Completed {
			completedTaskCount++
		}
	}

	return completedTaskCount
}

func getShowGroupsRows(groups []items.AdvancedChecklistGroup, showCompleted bool) [][]*simpletable.Cell {
	var rows [][]*simpletable.Cell

	for x := range groups {
		rows = append(rows, getShowGroupsTaskRows(groups[x], showCompleted)...)
	}

	return rows
}

func getShowTaskRows(tasklist items.Tasklist, showCompleted bool) [][]*simpletable.Cell {
	var rowsCells [][]*simpletable.Cell // table

	var rowCells []*simpletable.Cell // row

	// output open and then completed tasks (if specified)
	seqs := []bool{false}
	if showCompleted {
		seqs = append(seqs, true)
	}

	// for incomplete and then (if selected) completed tasks
	for _, seq := range seqs {
		// filter complete/incomplete tasks
		filtered := filterTasks(tasklist.Tasks, seq)
		for x, task := range filtered {
			// adding rowCells to output
			rowCells = []*simpletable.Cell{
				{Align: simpletable.AlignRight, Text: strconv.Itoa(x + 1)},
				{Align: simpletable.AlignLeft, Text: outputChars(task.Title, defaultMaxLength)},
			}

			// output completed column if we're showing completed tasks
			if showCompleted {
				rowCells = append(rowCells, &simpletable.Cell{
					Align: simpletable.AlignLeft, Text: boolToText(task.Completed, "yes", "no"),
				})
			}

			rowsCells = append(rowsCells, rowCells)
		}
	}

	return rowsCells
}

func (ci *ShowTasklistInput) Run() error {
	if ci.Title == "" && ci.UUID == "" {
		var std items.Tasklists

		var adv items.AdvancedChecklists

		std, adv, err := getAllLists(ci.Session)
		if err != nil {
			return err
		}

		// Initialize an empty slice to hold the options
		var options []string

		for x := range std {
			options = append(options, std[x].Title)
		}

		for x := range adv {
			options = append(options, adv[x].Title)
		}

		selectedOption, _ := pterm.InteractiveSelectPrinter{
			TextStyle:       &pterm.ThemeDefault.TreeTextStyle,
			DefaultText:     "choose list to display",
			Options:         []string{},
			OptionStyle:     &pterm.ThemeDefault.DefaultText,
			DefaultOption:   "",
			MaxHeight:       selectListHeight,
			Selector:        ">",
			SelectorStyle:   &pterm.ThemeDefault.SecondaryStyle,
			OnInterruptFunc: nil,
			Filter:          true,
		}.WithOptions(options).Show()

		ci.Title = selectedOption
	}

	std, adv, err := getAllMatchingLists(ci.Session, ci.Title, ci.UUID)
	if err != nil {
		return err
	}

	if len(std)+len(adv) > 1 {
		return errors.New("more than one match found. use --uuid flag to specify.")
		// TODO: output table with all the item titles, type, uuid, and last updated
	}

	if len(std) > 0 {
		showTaskList(std[0], ci.ShowCompleted)
	} else if len(adv) > 0 {
		showAdvancedChecklist(adv[0], ci.ShowCompleted)
	}

	return nil
}

func showTaskList(tasklist items.Tasklist, showCompleted bool) {
	table := simpletable.New()

	headerCells := []*simpletable.Cell{
		{Align: simpletable.AlignCenter, Text: color.Bold.Text("-")},
		{Align: simpletable.AlignCenter, Text: color.Bold.Text("title")},
	}

	if showCompleted {
		headerCells = append(headerCells, &simpletable.Cell{Align: simpletable.AlignLeft, Text: color.Bold.Text("completed")})
	}

	table.Header = &simpletable.Header{Cells: headerCells}

	table.Body.Cells = getShowTaskRows(tasklist, showCompleted)

	table.SetStyle(simpletable.StyleRounded)
	fmt.Println(table.String())
}

func showAdvancedChecklist(tasklist items.AdvancedChecklist, showCompleted bool) {
	table := simpletable.New()

	headerCells := []*simpletable.Cell{
		{Span: 1, Align: simpletable.AlignCenter, Text: color.Bold.Text("group")},
		{Span: 1, Align: simpletable.AlignCenter, Text: color.Bold.Text("title")},
		// {Align: simpletable.AlignCenter, Text: color.Bold.Text("title")},
		{Align: simpletable.AlignCenter, Text: color.Bold.Text("updated")},
	}

	if showCompleted {
		headerCells = append(headerCells, &simpletable.Cell{Align: simpletable.AlignLeft, Text: color.Bold.Text("completed")})
	}

	table.Header = &simpletable.Header{Cells: headerCells}
	table.Body.Cells = getShowGroupsRows(tasklist.Groups, showCompleted)
	table.SetStyle(simpletable.StyleRounded)
	fmt.Println(table.String())
}

type DefaultSection struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AdvancedChecklistTasks []AdvancedChecklistTask

type AdvancedChecklistTask struct {
	Id          string    `json:"id"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AdvancedChecklistSection struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Collapsed bool   `json:"collapsed"`
}
type AdvancedChecklistGroup struct {
	Name       string                     `json:"name"`
	LastActive time.Time                  `json:"lastActive"`
	Sections   []AdvancedChecklistSection `json:"sections"`
	Tasks      AdvancedChecklistTasks     `json:"tasks"`
	Collapsed  bool                       `json:"collapsed"`
}

type AdvancedChecklist struct {
	UUID            string                   `json:"-"`
	Duplicates      []AdvancedChecklist      `json:"-"`
	Title           string                   `json:"-"`
	SchemaVersion   string                   `json:"schemaVersion"`
	Groups          []AdvancedChecklistGroup `json:"groups"`
	DefaultSections []DefaultSection         `json:"defaultSections"`
	UpdatedAt       time.Time                `json:"updatedAt"`
	Trashed         bool                     `json:"trashed"`
}

func (t *AdvancedChecklistTasks) Sort() {
	// sort tasks by updated date descending
	sort.Slice(*t, func(i, j int) bool {
		dt := *t

		return dt[i].UpdatedAt.Unix() > dt[j].UpdatedAt.Unix()
	})
}

func (c *AdvancedChecklist) Sort() {
	// sort groups
	sort.Slice(c.Groups, func(i, j int) bool {
		return c.Groups[i].LastActive.Unix() > c.Groups[j].LastActive.Unix()
	})

	for x := range c.Groups {
		// sort group sections by name
		sort.Slice(c.Groups[x].Sections, func(i, j int) bool {
			return c.Groups[x].Sections[i].Name < c.Groups[x].Sections[j].Name
		})

		c.Groups[x].Tasks.Sort()
	}
}

type DeleteTaskInput struct {
	Session  *cache.Session
	Debug    bool
	Title    string
	Tasklist string
	UUID     string
}

type DeleteAdvancedChecklistTaskInput struct {
	Session   *cache.Session
	Debug     bool
	Title     string
	Group     string
	Checklist string
	UUID      string
}

func (ci *DeleteTaskInput) Run() error {
	// sync to get db
	_, err := Sync(cache.SyncInput{
		Session: ci.Session,
	}, true)
	if err != nil {
		return err
	}

	// get matching note
	notes, err := getNotesByTitleUUID(ci.Session, ci.Tasklist, ci.UUID, items.SimpleTaskEditorNoteType)
	if err != nil {
		return err
	}

	var cl items.Tasklist

	cl, err = notes[0].Content.ToTaskList()
	if err != nil {
		return err
	}

	err = cl.DeleteTask(ci.Title)
	if err != nil {
		return err
	}

	taskNoteText := items.TasksToNoteText(cl.Tasks)

	now := time.Now().UTC()
	notes[0].Content.SetUpdateTime(now)
	notes[0].Content.SetText(taskNoteText)
	notes[0].Content.SetUpdateTime(now)

	// save note to db
	si := cache.SyncInput{
		Session: ci.Session,
		Close:   false,
	}

	if err = cache.SaveNotes(ci.Session, ci.Session.CacheDB, items.Notes{notes[0]}, true); err != nil {
		return err
	}

	if _, err = cache.Sync(si); err != nil {
		return err
	}
	return nil
}

func (ci *DeleteAdvancedChecklistTaskInput) Run() error {
	// sync to get db
	_, err := Sync(cache.SyncInput{
		Session: ci.Session,
	}, true)
	if err != nil {
		return err
	}

	// get matching note
	notes, err := getNotesByTitleUUID(ci.Session, ci.Checklist, ci.UUID, items.AdvancedChecklistNoteType)
	if err != nil {
		return err
	}

	var cl items.AdvancedChecklist

	cl, err = notes[0].Content.ToAdvancedCheckList()
	if err != nil {
		return err
	}

	// delete task from the checklist
	if err = cl.DeleteTask(ci.Group, ci.Title); err != nil {
		return err
	}

	taskNoteText := items.AdvancedCheckListToNoteText(cl)

	now := time.Now().UTC()
	notes[0].Content.SetUpdateTime(now)
	notes[0].Content.SetText(taskNoteText)
	notes[0].Content.SetUpdateTime(now)

	// save note to db
	si := cache.SyncInput{
		Session: ci.Session,
		Close:   false,
	}

	if err = cache.SaveNotes(ci.Session, ci.Session.CacheDB, items.Notes{notes[0]}, true); err != nil {
		return err
	}

	if _, err = cache.Sync(si); err != nil {
		return err
	}

	return nil
}

func filterTasks(tasks items.Tasks, completed bool) items.Tasks {
	var filtered items.Tasks

	for _, task := range tasks {
		if task.Completed == completed {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

func boolToText(in bool, trueTxt, falseTxt string) string {
	if in {
		return trueTxt
	}

	return falseTxt
}

type CompleteAdvancedTaskInput struct {
	Session  *cache.Session
	Debug    bool
	Title    string
	Group    string
	Tasklist string
	UUID     string
}

type CompleteTaskInput struct {
	Session  *cache.Session
	Debug    bool
	Title    string
	Tasklist string
	UUID     string
}

func (ci *CompleteTaskInput) Run() error {
	// sync to get db
	_, err := Sync(cache.SyncInput{
		Session: ci.Session,
	}, true)
	if err != nil {
		return err
	}

	// get matching note
	notes, err := getNotesByTitleUUID(ci.Session, ci.Tasklist, ci.UUID, items.SimpleTaskEditorNoteType)
	if err != nil {
		return err
	}

	var cl items.Tasklist

	cl, err = notes[0].Content.ToTaskList()
	if err != nil {
		return err
	}

	err = cl.CompleteTask(ci.Title)
	if err != nil {
		return err
	}

	taskNoteText := items.TasksToNoteText(cl.Tasks)

	now := time.Now().UTC()
	notes[0].Content.SetUpdateTime(now)
	notes[0].Content.SetText(taskNoteText)
	notes[0].Content.SetUpdateTime(now)

	// save note to db
	si := cache.SyncInput{
		Session: ci.Session,
		Close:   false,
	}

	if err = cache.SaveNotes(ci.Session, ci.Session.CacheDB, items.Notes{notes[0]}, true); err != nil {
		return err
	}

	if _, err = cache.Sync(si); err != nil {
		return err
	}
	return nil
}

func (ci *CompleteAdvancedTaskInput) Run() error {
	// sync to get db
	_, err := Sync(cache.SyncInput{
		Session: ci.Session,
	}, true)
	if err != nil {
		return err
	}

	// get matching note
	notes, err := getNotesByTitleUUID(ci.Session, ci.Tasklist, ci.UUID, items.AdvancedChecklistNoteType)
	if err != nil {
		return err
	}

	var cl items.AdvancedChecklist

	cl, err = notes[0].Content.ToAdvancedCheckList()
	if err != nil {
		return err
	}

	err = cl.CompleteTask(ci.Group, ci.Title)
	if err != nil {
		return err
	}

	taskNoteText := items.AdvancedCheckListToNoteText(cl)

	now := time.Now().UTC()
	notes[0].Content.SetUpdateTime(now)
	notes[0].Content.SetText(taskNoteText)
	notes[0].Content.SetUpdateTime(now)

	// save note to db
	si := cache.SyncInput{
		Session: ci.Session,
		Close:   false,
	}

	if err = cache.SaveNotes(ci.Session, ci.Session.CacheDB, items.Notes{notes[0]}, true); err != nil {
		return err
	}

	if _, err = cache.Sync(si); err != nil {
		return err
	}

	return nil
}

type ReopenAdvancedTaskInput struct {
	Session  *cache.Session
	Debug    bool
	UUID     string
	Title    string
	Group    string
	Tasklist string
}

type ReopenTaskInput struct {
	Session  *cache.Session
	Debug    bool
	UUID     string
	Title    string
	Tasklist string
}

func (ci *ReopenTaskInput) Run() error {
	// sync to get db
	_, err := Sync(cache.SyncInput{
		Session: ci.Session,
	}, true)
	if err != nil {
		return err
	}

	// get matching note
	notes, err := getNotesByTitleUUID(ci.Session, ci.Tasklist, ci.UUID, items.SimpleTaskEditorNoteType)
	if err != nil {
		return err
	}

	var cl items.Tasklist

	cl, err = notes[0].Content.ToTaskList()
	if err != nil {
		return err
	}

	err = cl.ReopenTask(ci.Title)
	if err != nil {
		return err
	}

	taskNoteText := items.TasksToNoteText(cl.Tasks)

	now := time.Now().UTC()
	notes[0].Content.SetUpdateTime(now)
	notes[0].Content.SetText(taskNoteText)
	notes[0].Content.SetUpdateTime(now)

	// save note to db
	si := cache.SyncInput{
		Session: ci.Session,
		Close:   false,
	}

	if err = cache.SaveNotes(ci.Session, ci.Session.CacheDB, items.Notes{notes[0]}, true); err != nil {
		return err
	}

	if _, err = cache.Sync(si); err != nil {
		return err
	}
	return nil
}

func (ci *ReopenAdvancedTaskInput) Run() error {
	// sync to get db
	_, err := Sync(cache.SyncInput{
		Session: ci.Session,
	}, true)
	if err != nil {
		return err
	}

	// get matching note
	notes, err := getNotesByTitleUUID(ci.Session, ci.Tasklist, ci.UUID, items.AdvancedChecklistNoteType)
	if err != nil {
		return err
	}

	var cl items.AdvancedChecklist

	cl, err = notes[0].Content.ToAdvancedCheckList()
	if err != nil {
		return err
	}

	err = cl.ReopenTask(ci.Group, ci.Title)
	if err != nil {
		return err
	}

	taskNoteText := items.AdvancedCheckListToNoteText(cl)

	now := time.Now().UTC()
	notes[0].Content.SetUpdateTime(now)
	notes[0].Content.SetText(taskNoteText)
	notes[0].Content.SetUpdateTime(now)

	// save note to db
	si := cache.SyncInput{
		Session: ci.Session,
		Close:   false,
	}

	if err = cache.SaveNotes(ci.Session, ci.Session.CacheDB, items.Notes{notes[0]}, true); err != nil {
		return err
	}

	if _, err = cache.Sync(si); err != nil {
		return err
	}

	return nil
}
