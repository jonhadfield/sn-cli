package sncli

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/jonhadfield/gosn"
	"github.com/ryanuber/columnize"
)

var (
	bold   = color.New(color.Bold).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

func (input *StatsConfig) Run() error {
	gosn.SetErrorLogger(log.Println)

	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}

	getItemsInput := gosn.GetItemsInput{
		Session:  input.Session,
		PageSize: SNPageSize,
	}

	var err error
	// get all existing Tags and Notes
	var output gosn.GetItemsOutput

	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return err
	}

	output.Items.DeDupe()

	var items gosn.Items

	items, err = output.Items.DecryptAndParse(input.Session.Mk, input.Session.Ak)
	if err != nil {
		return err
	}

	var notes gosn.Items

	var oldestNote, newestNote, lastUpdatedNote time.Time

	//var deletedItemsUUIDs []string

	var missingContentUUIDs []string

	var missingContentTypeUUIDs []string

	allUUIDs := make([]string, len(items))

	var duplicateUUIDs []string

	var tCounter typeCounter

	tCounter.counts = make(map[string]int64)

	for _, item := range items {
		tCounter.update(item.ContentType)

		if StringInSlice(item.UUID, allUUIDs, false) {
			duplicateUUIDs = append(duplicateUUIDs, item.UUID)
		}

		allUUIDs = append(allUUIDs, item.UUID)

		if item.Deleted {
			tCounter.update("Deleted")
			//deletedItemsUUIDs = append(deletedItemsUUIDs, item.UUID)
		}

		if !item.Deleted && item.ContentType == "" {
			missingContentTypeUUIDs = append(missingContentTypeUUIDs, item.UUID)
		}

		if item.ContentType == "Note" {
			if !item.Deleted && item.Content == nil {
				missingContentUUIDs = append(missingContentUUIDs, item.UUID)
			}

			var cTime time.Time
			cTime, err = time.Parse(timeLayout, item.CreatedAt)

			if err != nil {
				return err
			}

			var uTime time.Time

			uTime, err = time.Parse(timeLayout, item.UpdatedAt)
			if err != nil {
				return err
			}

			if !item.Deleted && oldestNote.IsZero() || cTime.Before(oldestNote) {
				oldestNote, err = time.Parse(timeLayout, item.CreatedAt)
				if err != nil {
					return err
				}
			}

			if !item.Deleted && newestNote.IsZero() || cTime.After(newestNote) {
				newestNote, err = time.Parse(timeLayout, item.CreatedAt)
				if err != nil {
					return err
				}
			}

			if !item.Deleted && lastUpdatedNote.IsZero() || uTime.After(lastUpdatedNote) {
				lastUpdatedNote, err = time.Parse(timeLayout, item.UpdatedAt)
				if err != nil {
					return err
				}
			}

			if !item.Deleted && item.ContentSize > 0 {
				notes = append(notes, item)
			}
		}
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].ContentSize > notes[j].ContentSize
	})

	fmt.Println(green("COUNTS"))
	tCounter.present()
	//fmt.Println("Deleted:", len(deletedItemsUUIDs))

	fmt.Println(green("\nSTATS"))
	var statLines []string
	if len(notes) > 0 {
		statLines = append(statLines, fmt.Sprintf("Oldest | %v", timeSince(oldestNote.Local())))
		statLines = append(statLines, fmt.Sprintf("Newest | %v", timeSince(newestNote.Local())))
		statLines = append(statLines, fmt.Sprintf("Updated | %v", timeSince(lastUpdatedNote.Local())))
		fmt.Println(columnize.SimpleFormat(statLines))

		fmt.Println("Largest:")

		var finalItem int

		if len(notes) >= 5 {
			finalItem = 4
		} else {
			finalItem = len(notes)
		}

		for x := 0; x < finalItem; x++ {
			fmt.Printf(" - %d bytes: \"%s\"\n", notes[x].ContentSize, notes[x].Content.GetTitle())
		}
	} else {
		fmt.Println("no notes returned")
	}

	fmt.Println(green("\nISSUES"))
	if allEmpty(duplicateUUIDs, missingContentUUIDs, missingContentTypeUUIDs) {
		fmt.Println("None")
	}
	if len(duplicateUUIDs) > 0 {
		fmt.Println("Duplicate note UUIDs: ", outList(duplicateUUIDs, ", "))
	}
	if len(missingContentUUIDs) > 0 {
		fmt.Println("Missing content UUIDs:", outList(missingContentUUIDs, ", "))
	}
	if len(missingContentTypeUUIDs) > 0 {
		fmt.Println("Missing content type UUIDs:", outList(missingContentTypeUUIDs, ", "))
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
		if name != "Tag" && name != "Note" {
			lines = append(lines, fmt.Sprintf("%s ^ %d", name, count))
		}
	}
	lines = append(lines, fmt.Sprintf("Deleted ^ %d", in.counts["Deleted"]))

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
