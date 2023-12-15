package sncli

import (
	"fmt"
	"sort"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/alexeyco/simpletable"

	// "github.com/fatih/color"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/ryanuber/columnize"
)

type StatsData struct {
	CoreTypeCounter   typeCounter
	OtherTypeCounter  typeCounter
	LargestNotes      []*items.Note
	ItemsOrphanedRefs []ItemOrphanedRefs
	LastUpdatedNote   *items.Note
	NewestNote        *items.Note
	OldestNote        *items.Note
}

func (i *StatsConfig) GetData() (StatsData, error) {
	var err error

	var so cache.SyncOutput

	so, err = Sync(cache.SyncInput{
		Session: &i.Session,
	}, true)
	if err != nil {
		return StatsData{}, err
	}

	var allPersistedItems cache.Items

	if err = so.DB.All(&allPersistedItems); err != nil {
		return StatsData{}, err
	}

	var gitems items.Items
	gitems, err = allPersistedItems.ToItems(&i.Session)
	if err != nil {
		return StatsData{}, err
	}

	var notes items.Items

	var itemsOrphanedRefs []ItemOrphanedRefs

	var missingItemsKey []string

	var missingContentUUIDs []string

	var missingContentTypeUUIDs []string

	allUUIDs := make([]string, len(allPersistedItems))
	for x := range allPersistedItems {
		allUUIDs = append(allUUIDs, allPersistedItems[x].UUID)
	}

	var ctCounter, otCounter typeCounter

	ctCounter.counts = make(map[string]int64)
	otCounter.counts = make(map[string]int64)

	var oldestNoteTime, newestNoteTime, lastUpdatedNoteTime time.Time
	var oldestNote, newestNote, lastUpdatedNote *items.Note

	for _, item := range gitems {
		// check if item is trashed note
		var isTrashedNote bool
		if item.GetContentType() == "Note" {
			n := item.(*items.Note)
			if n.Content.Trashed != nil && *n.Content.Trashed {
				isTrashedNote = true
			}
		}

		switch {
		case isTrashedNote:
			ctCounter.update("Notes (In Trash)")
		case item.GetContentType() == "Note":
			ctCounter.update("Note")
		case item.GetContentType() == "Tag":
			ctCounter.update("Tag")
		default:
			otCounter.update(item.GetContentType())
		}

		if item.GetItemsKeyID() == "" {
			missingItemsKey = append(missingItemsKey, fmt.Sprintf("- type: %s uuid: %s %s", item.GetContentType(), item.GetUUID(), item.GetItemsKeyID()))
		}

		if item.GetContentType() == "" {
			missingContentTypeUUIDs = append(missingContentTypeUUIDs, item.GetUUID())
		}

		refs := item.GetContent().References()
		for _, ref := range refs {
			if !StringInSlice(ref.UUID, allUUIDs, false) {
				itemsOrphanedRefs = append(itemsOrphanedRefs, ItemOrphanedRefs{
					ContentType:  item.GetContentType(),
					Item:         item,
					OrphanedRefs: []string{ref.UUID},
				})

				break
			}
		}

		if item.GetContentType() == "Note" && !isTrashedNote {
			if item.GetContent() == nil {
				missingContentUUIDs = append(missingContentUUIDs, item.GetUUID())
			}

			var cTime time.Time
			cTime, err = time.Parse(timeLayout, item.GetCreatedAt())

			if err != nil {
				return StatsData{}, err
			}

			var uTime time.Time

			uTime, err = time.Parse(timeLayout, item.GetUpdatedAt())
			if err != nil {
				return StatsData{}, err
			}

			if oldestNoteTime.IsZero() || cTime.Before(oldestNoteTime) {
				oldestNote = item.(*items.Note)

				oldestNoteTime = cTime
			}

			if newestNoteTime.IsZero() || cTime.After(newestNoteTime) {
				newestNote = item.(*items.Note)

				newestNoteTime = cTime
			}

			if lastUpdatedNoteTime.IsZero() || uTime.After(lastUpdatedNoteTime) {
				lastUpdatedNote = item.(*items.Note)

				lastUpdatedNoteTime = uTime
			}

			// create slice of notes with non-zero content size
			if item.GetContentSize() > 0 {
				notes = append(notes, item)
			}
		}
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].GetContentSize() > notes[j].GetContentSize()
	})

	var largestNotes []*items.Note

	if len(notes) > 0 {
		finalItem := len(notes)

		if len(notes) >= 5 {
			finalItem = 4
		}

		for x := 0; x < finalItem; x++ {
			largestNotes = append(largestNotes, notes[x].(*items.Note))
		}
	}

	return StatsData{
		CoreTypeCounter:   ctCounter,
		OtherTypeCounter:  otCounter,
		LargestNotes:      largestNotes,
		ItemsOrphanedRefs: itemsOrphanedRefs,
		LastUpdatedNote:   lastUpdatedNote,
		NewestNote:        newestNote,
		OldestNote:        oldestNote,
	}, nil
}

type ItemOrphanedRefs struct {
	ContentType  string
	Item         items.Item
	OrphanedRefs []string
}

func showNoteHistory(data StatsData) {
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("")},
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("Title")},
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("Time")},
		},
	}
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "Oldest"},
		{Text: fmt.Sprintf("%s", data.OldestNote.Content.Title)},
		{Text: fmt.Sprintf("%s", humanize.Time(time.UnixMicro(data.OldestNote.CreatedAtTimestamp)))},
	})
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "Newest"},
		{Text: fmt.Sprintf("%s", data.NewestNote.Content.Title)},
		{Text: fmt.Sprintf("%s", humanize.Time(time.UnixMicro(data.NewestNote.CreatedAtTimestamp)))},
	})
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "Most Recently Updated"},
		{Text: fmt.Sprintf("%s", data.LastUpdatedNote.Content.Title)},
		{Text: fmt.Sprintf("%s", humanize.Time(time.UnixMicro(data.LastUpdatedNote.UpdatedAtTimestamp)))},
	})

	color.Bold.Println("Note History")

	table.Println()
}

func showItemCounts(data StatsData) {
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("Type")},
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("Count")},
		},
	}
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "Notes"},
		{Text: fmt.Sprintf("%s", humanize.Comma(data.CoreTypeCounter.counts["Note"]))},
	})
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "Tags"},
		{Text: fmt.Sprintf("%s", humanize.Comma(data.CoreTypeCounter.counts["Tag"]))},
	})
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "----------------"},
		{Text: "------"},
	})
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "Notes (In Trash)"},
		{Text: fmt.Sprintf("%s", humanize.Comma(data.CoreTypeCounter.counts["Notes (In Trash)"]))},
	})
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "----------------"},
		{Text: "------"},
	})

	var keys []string

	for key := range data.OtherTypeCounter.counts {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return data.OtherTypeCounter.counts[keys[i]] > data.OtherTypeCounter.counts[keys[j]]
	})

	for _, k := range keys {
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: k},
			{Text: fmt.Sprintf("%s", humanize.Comma(data.OtherTypeCounter.counts[k]))},
		})
	}

	color.Bold.Println("Item Counts")
	table.Println()
}

func showLargestNotes(data StatsData) {
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("Size")},
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("Title")},
		},
	}

	for _, note := range data.LargestNotes {
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: fmt.Sprintf("%s", humanize.Bytes(uint64(note.GetContentSize())))},
			{Text: fmt.Sprintf("%s", note.Content.Title)},
		})
	}

	color.Bold.Println("Largest Notes")

	table.Println()
}

func (i *StatsConfig) Run() error {
	data, err := i.GetData()
	if err != nil {
		return err
	}

	showItemCounts(data)
	showNoteHistory(data)
	showLargestNotes(data)

	return err
}

type typeCounter struct {
	counts map[string]int64
}

func (in *typeCounter) update(itemType string) {
	var found bool

	for name := range in.counts {
		if name == itemType {
			found = true
			in.counts[name]++
		}
	}

	if !found {
		in.counts[itemType] = 1
	}
}

func (in *typeCounter) present() {
	var lines []string
	lines = append(lines, fmt.Sprintf("Notes ^ %d", in.counts["Note"]))
	lines = append(lines, fmt.Sprintf("Tags ^ %d", in.counts["Tag"]))

	for name, count := range in.counts {
		if name != "Tag" && name != "Note" && name != "Deleted" {
			lines = append(lines, fmt.Sprintf("%s ^ %d", name, count))
		}
	}

	config := columnize.DefaultConfig()
	config.Delim = "^"
	fmt.Println(columnize.Format(lines, config))
}
