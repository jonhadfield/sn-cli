package sncli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/asdine/storm/v3"
	"github.com/dustin/go-humanize"

	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
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
	DuplicateNotes    []*items.Note
}

func (i *StatsConfig) GetData() (StatsData, error) {
	var err error

	// Always sync first to get latest data
	var so cache.SyncOutput
	so, err = Sync(cache.SyncInput{
		Session: &i.Session,
	}, true)

	// Add delay after sync to prevent consecutive operations from hanging
	time.Sleep(1 * time.Second)

	if err != nil {
		// Handle sync timeout gracefully - fall back to cached data if available
		if strings.Contains(err.Error(), "giving up after") {
			fmt.Printf("Warning: Sync timed out, attempting to use cached data\n")

			var allPersistedItems cache.Items
			if i.Session.CacheDBPath != "" {
				if cacheDB, dbErr := storm.Open(i.Session.CacheDBPath); dbErr == nil {
					if dbErr = cacheDB.All(&allPersistedItems); dbErr == nil && len(allPersistedItems) > 0 {
						cacheDB.Close()
						fmt.Printf("Using cached data (%d items)\n", len(allPersistedItems))
						return i.processStatsFromCacheMetadata(allPersistedItems)
					}
					cacheDB.Close()
				}
			}

			// If no cached data available, return empty stats
			fmt.Printf("No cached data available, returning empty stats\n")
			return StatsData{
				CoreTypeCounter:  typeCounter{counts: make(map[string]int64)},
				OtherTypeCounter: typeCounter{counts: make(map[string]int64)},
			}, nil
		} else {
			return StatsData{}, err
		}
	}

	// Ensure we close the database when done
	defer func() {
		if so.DB != nil {
			so.DB.Close()
		}
	}()

	var allPersistedItems cache.Items
	if err = so.DB.All(&allPersistedItems); err != nil {
		return StatsData{}, err
	}

	return i.processStatsFromCache(allPersistedItems)
}

// processStatsFromCacheMetadata generates stats from cache metadata without decryption
// This is used for large datasets to avoid sync timeout issues
func (i *StatsConfig) processStatsFromCacheMetadata(allPersistedItems cache.Items) (StatsData, error) {
	var ctCounter, otCounter typeCounter
	ctCounter.counts = make(map[string]int64)
	otCounter.counts = make(map[string]int64)

	// Count items by type using cache metadata (no decryption needed)
	for _, item := range allPersistedItems {
		if item.Deleted {
			continue // Skip deleted items
		}

		switch item.ContentType {
		case common.SNItemTypeNote:
			ctCounter.update(common.SNItemTypeNote)
		case common.SNItemTypeTag:
			ctCounter.update(common.SNItemTypeTag)
		default:
			otCounter.update(item.ContentType)
		}
	}

	// Return simplified stats data (no note details since we can't decrypt)
	return StatsData{
		CoreTypeCounter:   ctCounter,
		OtherTypeCounter:  otCounter,
		LargestNotes:      []*items.Note{},      // Empty for metadata-only processing
		ItemsOrphanedRefs: []ItemOrphanedRefs{}, // Empty for metadata-only processing
		LastUpdatedNote:   nil,
		NewestNote:        nil,
		OldestNote:        nil,
		DuplicateNotes:    []*items.Note{}, // Empty for metadata-only processing
	}, nil
}

func (i *StatsConfig) processStatsFromCache(allPersistedItems cache.Items) (StatsData, error) {
	gitems, err := allPersistedItems.ToItems(&i.Session)
	if err != nil {
		return StatsData{}, err
	}

	allUUIDs := make([]string, 0, len(allPersistedItems))
	for _, item := range allPersistedItems {
		allUUIDs = append(allUUIDs, item.UUID)
	}

	ctCounter := typeCounter{counts: make(map[string]int64)}
	otCounter := typeCounter{counts: make(map[string]int64)}
	var notes items.Items
	var itemsOrphanedRefs []ItemOrphanedRefs
	var oldestNote, newestNote, lastUpdatedNote *items.Note
	var oldestTime, newestTime, lastUpdatedTime time.Time

	for _, item := range gitems {
		// Count items by type
		if err := i.processItemCount(item, &ctCounter, &otCounter); err != nil {
			return StatsData{}, err
		}

		// Check for orphaned references
		i.checkOrphanedRefs(item, allUUIDs, &itemsOrphanedRefs)

		// Process notes for size and time tracking
		if note, isNote := item.(*items.Note); isNote && !i.isNoteInTrash(note) {
			stats := &noteStats{
				notes:           &notes,
				oldestNote:      &oldestNote,
				newestNote:      &newestNote,
				lastUpdatedNote: &lastUpdatedNote,
				oldestTime:      &oldestTime,
				newestTime:      &newestTime,
				lastUpdatedTime: &lastUpdatedTime,
			}
			if err := i.processNoteStats(note, stats); err != nil {
				return StatsData{}, err
			}
		}
	}

	// Get largest notes (top 5)
	largestNotes := i.getLargestNotes(notes)

	// Get duplicate notes
	duplicateNotes := i.findDuplicateNotes(gitems)

	return StatsData{
		CoreTypeCounter:   ctCounter,
		OtherTypeCounter:  otCounter,
		LargestNotes:      largestNotes,
		ItemsOrphanedRefs: itemsOrphanedRefs,
		LastUpdatedNote:   lastUpdatedNote,
		NewestNote:        newestNote,
		OldestNote:        oldestNote,
		DuplicateNotes:    duplicateNotes,
	}, nil
}

func (i *StatsConfig) processItemCount(item items.Item, ctCounter, otCounter *typeCounter) error {
	isTrashedNote := i.isNoteInTrash(item)

	switch {
	case isTrashedNote:
		ctCounter.update("Notes (In Trash)")
	case item.GetContentType() == common.SNItemTypeNote:
		ctCounter.update(common.SNItemTypeNote)
	case item.GetContentType() == common.SNItemTypeTag:
		ctCounter.update(common.SNItemTypeTag)
	default:
		otCounter.update(item.GetContentType())
	}
	return nil
}

func (i *StatsConfig) isNoteInTrash(item items.Item) bool {
	if item.GetContentType() != common.SNItemTypeNote {
		return false
	}
	note, ok := item.(*items.Note)
	if !ok {
		return false
	}
	return note.Content.Trashed != nil && *note.Content.Trashed
}

func (i *StatsConfig) checkOrphanedRefs(item items.Item, allUUIDs []string, itemsOrphanedRefs *[]ItemOrphanedRefs) {
	refs := item.GetContent().References()
	for _, ref := range refs {
		if !StringInSlice(ref.UUID, allUUIDs, false) {
			*itemsOrphanedRefs = append(*itemsOrphanedRefs, ItemOrphanedRefs{
				ContentType:  item.GetContentType(),
				Item:         item,
				OrphanedRefs: []string{ref.UUID},
			})
			break
		}
	}
}

type noteStats struct {
	notes           *items.Items
	oldestNote      **items.Note
	newestNote      **items.Note
	lastUpdatedNote **items.Note
	oldestTime      *time.Time
	newestTime      *time.Time
	lastUpdatedTime *time.Time
}

func (i *StatsConfig) processNoteStats(note *items.Note, stats *noteStats) error {
	cTime, err := time.Parse(timeLayout, note.GetCreatedAt())
	if err != nil {
		return err
	}

	uTime, err := time.Parse(timeLayout, note.GetUpdatedAt())
	if err != nil {
		return err
	}

	if stats.oldestTime.IsZero() || cTime.Before(*stats.oldestTime) {
		*stats.oldestNote = note
		*stats.oldestTime = cTime
	}

	if stats.newestTime.IsZero() || cTime.After(*stats.newestTime) {
		*stats.newestNote = note
		*stats.newestTime = cTime
	}

	if stats.lastUpdatedTime.IsZero() || uTime.After(*stats.lastUpdatedTime) {
		*stats.lastUpdatedNote = note
		*stats.lastUpdatedTime = uTime
	}

	if note.GetContentSize() > 0 {
		*stats.notes = append(*stats.notes, note)
	}

	return nil
}

func (i *StatsConfig) getLargestNotes(notes items.Items) []*items.Note {
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].GetContentSize() > notes[j].GetContentSize()
	})

	var largestNotes []*items.Note
	maxNotes := 5
	if len(notes) < maxNotes {
		maxNotes = len(notes)
	}

	for i := 0; i < maxNotes; i++ {
		largestNotes = append(largestNotes, notes[i].(*items.Note))
	}

	return largestNotes
}

func (i *StatsConfig) findDuplicateNotes(gitems items.Items) []*items.Note {
	var duplicateNotes []*items.Note

	for _, item := range gitems {
		if note, isNote := item.(*items.Note); isNote {
			// Check if this note has a DuplicateOf field set
			if !note.IsDeleted() && note.GetDuplicateOf() != "" {
				duplicateNotes = append(duplicateNotes, note)
			}
		}
	}

	return duplicateNotes
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

	if data.OldestNote != nil {
		data.OldestNote.Content = items.NoteContent{}
	}

	if data.NewestNote != nil {
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: "Newest"},
			{Text: fmt.Sprintf("%s", data.NewestNote.Content.Title)},
			{Text: fmt.Sprintf("%s", humanize.Time(time.UnixMicro(data.NewestNote.CreatedAtTimestamp)))},
		})
	}
	if data.LastUpdatedNote != nil {
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: "Most Recently Updated"},
			{Text: fmt.Sprintf("%s", data.LastUpdatedNote.Content.Title)},
			{Text: fmt.Sprintf("%s", humanize.Time(time.UnixMicro(data.LastUpdatedNote.UpdatedAtTimestamp)))},
		})
	}

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
		{Text: fmt.Sprintf("%s", humanize.Comma(data.CoreTypeCounter.counts[common.SNItemTypeNote]))},
	})
	table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
		{Text: "Tags"},
		{Text: fmt.Sprintf("%s", humanize.Comma(data.CoreTypeCounter.counts[common.SNItemTypeTag]))},
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

func showDuplicateNotes(data StatsData) {
	if len(data.DuplicateNotes) == 0 {
		return
	}

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("Title")},
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("UUID")},
			{Align: simpletable.AlignLeft, Text: color.Bold.Sprintf("Duplicate Of")},
		},
	}

	for _, note := range data.DuplicateNotes {
		title := note.Content.Title
		if title == "" {
			title = "(Untitled)"
		}
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: title},
			{Text: note.UUID},
			{Text: note.GetDuplicateOf()},
		})
	}

	color.Bold.Println("Duplicate Notes")
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
	showDuplicateNotes(data)

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
	lines = append(lines, fmt.Sprintf("Notes ^ %d", in.counts[common.SNItemTypeNote]))
	lines = append(lines, fmt.Sprintf("Tags ^ %d", in.counts[common.SNItemTypeTag]))

	for name, count := range in.counts {
		if name != common.SNItemTypeTag && name != common.SNItemTypeNote && name != "Deleted" {
			lines = append(lines, fmt.Sprintf("%s ^ %d", name, count))
		}
	}

	config := columnize.DefaultConfig()
	config.Delim = "^"
	fmt.Println(columnize.Format(lines, config))
}
