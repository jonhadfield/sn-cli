package sncli

import (
	"fmt"
	"log"
	"time"

	"github.com/jonhadfield/gosn"
)

const (
	timeLayout  = "2006-01-02T15:04:05.000Z"
	SNServerURL = "https://sync.standardnotes.org"
)

type TagItemsConfig struct {
	Session    gosn.Session
	FindText   string
	FindTag    string
	NewTags    []string
	Replace    bool
	IgnoreCase bool
	Debug      bool
}

type AddTagConfig struct {
	Session gosn.Session
	Tags    []string
	Debug   bool
}

type GetTagConfig struct {
	Session   gosn.Session
	Filters   gosn.ItemFilters
	TagTitles []string
	TagUUIDs  []string
	Output    string
	Debug     bool
}

type GetNoteConfig struct {
	Session    gosn.Session
	Filters    gosn.ItemFilters
	NoteTitles []string
	TagTitles  []string
	TagUUIDs   []string
	PageSize   int
	BatchSize  int
	Debug      bool
}

type DeleteTagConfig struct {
	Session   gosn.Session
	Email     string
	TagTitles []string
	TagUUIDs  []string
	Debug     bool
}

type AddNoteConfig struct {
	Session gosn.Session
	Title   string
	Text    string
	Tags    []string
	Replace bool
	Debug   bool
}

type DeleteNoteConfig struct {
	Session    gosn.Session
	NoteTitles []string
	NoteUUIDs  []string
	Debug      bool
}

type WipeConfig struct {
	Session gosn.Session
	Debug   bool
}

type StatsConfig struct {
	Session gosn.Session
	Debug   bool
}

func referenceExists(item gosn.Item, refID string) bool {
	for _, ref := range item.Content.References() {
		if ref.UUID == refID {
			return true
		}
	}
	return false
}

func (input *WipeConfig) Run() (int, error) {
	gosn.SetErrorLogger(log.Println)
	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}
	getItemsInput := gosn.GetItemsInput{
		Session: input.Session,
	}
	var err error
	// get all existing Tags and Notes and mark for deletion
	var output gosn.GetItemsOutput
	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return 0, err
	}
	output.DeDupe()
	var itemsToDel []gosn.Item
	for _, item := range output.Items {
		if item.Deleted {
			continue
		}
		if item.ContentType == "Tag" {
			item.Deleted = true
			itemsToDel = append(itemsToDel, item)
		} else if item.ContentType == "Note" {
			item.Deleted = true
			itemsToDel = append(itemsToDel, item)
		}
	}
	// delete items
	putItemsInput := gosn.PutItemsInput{
		Session:   input.Session,
		Items:     itemsToDel,
		SyncToken: output.SyncToken,
	}
	_, err = gosn.PutItems(putItemsInput)
	if err != nil {
		return 0, err
	}
	return len(itemsToDel), err
}

type FixupConfig struct {
	Session gosn.Session
	Debug   bool
}

func (input *FixupConfig) Run() error {
	gosn.SetErrorLogger(log.Println)
	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}
	getItemsInput := gosn.GetItemsInput{
		Session: input.Session,
	}
	var err error
	// get all existing Notes with nil content
	var output gosn.GetItemsOutput
	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return err
	}
	output.DeDupe()
	// notes with missing content to delete
	var missingContentType []gosn.Item
	var missingContent []gosn.Item
	var notesToTitleFix []gosn.Item

	for _, item := range output.Items {
		if !item.Deleted {
			switch {
			case item.ContentType == "":
				item.Deleted = true
				item.ContentType = "Note"
				missingContentType = append(missingContentType, item)
			case item.Content == nil:
				item.Deleted = true
				missingContent = append(missingContent, item)
			default:
				if item.ContentType == "Note" && item.Content.GetTitle() == "" {
					item.Content.SetUpdateTime(time.Now())
					item.Content.SetTitle("untitled")
					notesToTitleFix = append(notesToTitleFix, item)
				}
			}
		}
	}

	if len(missingContentType) > 0 {
		fmt.Printf("found %d notes with missing content type. delete? ", len(missingContentType))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   missingContentType,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed %d items\n", len(missingContentType))

		}
	} else {
		fmt.Println("no notes with missing content type")
	}

	if len(missingContent) > 0 {
		fmt.Printf("found %d notes with missing content. delete? ", len(missingContent))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   missingContent,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed %d items\n", len(missingContent))

		}
	} else {
		fmt.Println("no notes with missing content")
	}

	if len(notesToTitleFix) > 0 {
		fmt.Printf("found %d items with missing titles. fix? ", len(notesToTitleFix))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   notesToTitleFix,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed %d items", len(notesToTitleFix))

		}
	} else {
		fmt.Println("no items with missing titles")
	}
	return err
}
