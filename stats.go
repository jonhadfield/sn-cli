package sncli

import (
	"fmt"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/ryanuber/columnize"
)

var (
	Bold    = color.New(color.Bold).SprintFunc()
	Red     = color.New(color.FgRed).SprintFunc()
	Green   = color.New(color.FgGreen).SprintFunc()
	Yellow  = color.New(color.FgYellow).SprintFunc()
	HiWhite = color.New(color.FgHiWhite).SprintFunc()
)

type stats struct {
	oldestNote, newestNote, lastUpdatedNote time.Time
	hasOrphanedRefs                         []gosn.Item
}

func (i *StatsConfig) Run() error {
	var st stats
	var err error

	var so cache.SyncOutput

	so, err = Sync(cache.SyncInput{
		Session: &i.Session,
	}, true)
	if err != nil {
		return err
	}

	var allPersistedItems cache.Items

	if err = so.DB.All(&allPersistedItems); err != nil {
		return err
	}

	var items gosn.Items
	items, err = allPersistedItems.ToItems(&i.Session)
	if err != nil {
		return err
	}

	var notes gosn.Items

	var missingItemsKey []string

	var missingContentUUIDs []string

	var missingContentTypeUUIDs []string

	allUUIDs := make([]string, len(allPersistedItems))
	for x := range allPersistedItems {
		allUUIDs = append(allUUIDs, allPersistedItems[x].UUID)
	}

	var duplicateUUIDs []string

	var tCounter typeCounter

	tCounter.counts = make(map[string]int64)

	for _, item := range items {
		// check if item is trashed note
		var isTrashedNote bool
		if item.GetContentType() == "Note" {
			n := item.(*gosn.Note)
			if n.Content.Trashed != nil && *n.Content.Trashed {
				isTrashedNote = true
			}
		}

		if isTrashedNote {
			tCounter.update("Notes (In Trash)")
		} else {
			tCounter.update(item.GetContentType())
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
				st.hasOrphanedRefs = append(st.hasOrphanedRefs, item)

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
				return err
			}

			var uTime time.Time

			uTime, err = time.Parse(timeLayout, item.GetUpdatedAt())
			if err != nil {
				return err
			}

			if st.oldestNote.IsZero() || cTime.Before(st.oldestNote) {
				st.oldestNote, err = time.Parse(timeLayout, item.GetCreatedAt())
				if err != nil {
					return err
				}
			}

			if st.newestNote.IsZero() || cTime.After(st.newestNote) {
				st.newestNote, err = time.Parse(timeLayout, item.GetCreatedAt())
				if err != nil {
					return err
				}
			}

			if st.lastUpdatedNote.IsZero() || uTime.After(st.lastUpdatedNote) {
				st.lastUpdatedNote, err = time.Parse(timeLayout, item.GetUpdatedAt())
				if err != nil {
					return err
				}
			}

			if item.GetContentSize() > 0 {
				notes = append(notes, item)
			}
		}
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].GetContentSize() > notes[j].GetContentSize()
	})

	fmt.Println(Green("COUNTS"))
	tCounter.present()
	fmt.Println(Green("\nSTATS"))

	var statLines []string

	if len(notes) > 0 {
		statLines = append(statLines, fmt.Sprintf("Oldest | %v", timeSince(st.oldestNote.Local())))
		statLines = append(statLines, fmt.Sprintf("Newest | %v", timeSince(st.newestNote.Local())))
		statLines = append(statLines, fmt.Sprintf("Updated | %v", timeSince(st.lastUpdatedNote.Local())))
		fmt.Println(columnize.SimpleFormat(statLines))

		fmt.Println("Largest:")

		var finalItem int

		if len(notes) >= 5 {
			finalItem = 4
		} else {
			finalItem = len(notes)
		}

		for x := 0; x < finalItem; x++ {
			note := notes[x].(*gosn.Note)
			fmt.Printf(" - %d bytes: \"%s\"\n", note.GetContentSize(), note.Content.Title)
		}
	} else {
		fmt.Println("no notes returned")
	}

	fmt.Println(Green("\nISSUES"))

	if allEmpty(duplicateUUIDs, missingContentUUIDs, missingContentTypeUUIDs, missingItemsKey) && len(st.hasOrphanedRefs) == 0 {
		fmt.Println("None")
	}

	if len(st.hasOrphanedRefs) > 0 {
		fmt.Println("items with dangling references")
		for x := range st.hasOrphanedRefs {
			o := st.hasOrphanedRefs[x]
			fmt.Printf("- %s %s\n", o.GetContentType(), o.GetUUID())
		}
	}

	if len(missingContentUUIDs) > 0 {
		fmt.Println("Missing content UUIDs:", outList(missingContentUUIDs, ", "))
	}

	if len(missingContentTypeUUIDs) > 0 {
		fmt.Println("Missing content type UUIDs:", outList(missingContentTypeUUIDs, ", "))
	}

	if len(missingItemsKey) > 0 {
		fmt.Println("Missing items key ID:\n", outList(missingItemsKey, "\n"))
	}

	return err
}

func allEmpty(in ...[]string) bool {
	for _, i := range in {
		if len(i) > 0 {
			return false
		}
	}

	return true
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

func timeSince(inTime time.Time) string {
	now := time.Now()
	if inTime.Location() != now.Location() {
		now = now.In(inTime.Location())
	}

	if inTime.After(now) {
		inTime, now = now, inTime
	}

	y1, M1, d1 := inTime.Date()
	y2, M2, d2 := now.Date()

	h1, m1, s1 := inTime.Clock()
	h2, m2, s2 := now.Clock()

	year := y2 - y1
	month := M2 - M1
	day := d2 - d1
	hour := h2 - h1
	min := m2 - m1
	sec := s2 - s1

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}

	if min < 0 {
		min += 60
		hour--
	}

	if hour < 0 {
		hour += 24
		day--
	}

	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}

	if month < 0 {
		month += 12
		year--
	}

	// determine output
	switch {
	case year > 0:
		return fmt.Sprintf("%2d years %2d months %2d days", year, month, day)
	case month > 0:
		return fmt.Sprintf("%2d months %2d days %2d hours", month, day, hour)
	case day > 0:
		return fmt.Sprintf("%2d days %2d hours %2d minutes", day, hour, min)
	case hour > 0:
		return fmt.Sprintf("%2d hours %2d minutes", hour, min)
	case min > 0:
		return fmt.Sprintf("%2d minutes %2d seconds", min, sec)
	case sec > 0:
		return fmt.Sprintf("%2d seconds", sec)
	default:
		return "0"
	}
}
