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

type ItemReferenceYAML struct {
	UUID        string `yaml:"uuid"`
	ContentType string `yaml:"content_type"`
}

type ItemReferenceJSON struct {
	UUID        string `json:"uuid"`
	ContentType string `json:"content_type"`
}

type OrgStandardNotesSNDetailJSON struct {
	ClientUpdatedAt string `json:"client_updated_at"`
}

type OrgStandardNotesSNDetailYAML struct {
	ClientUpdatedAt string `yaml:"client_updated_at"`
}
type AppDataContentYAML struct {
	OrgStandardNotesSN OrgStandardNotesSNDetailYAML `yaml:"org.standardnotes.sn"`
}
type AppDataContentJSON struct {
	OrgStandardNotesSN OrgStandardNotesSNDetailJSON `json:"org.standardnotes.sn"`
}

type TagContentYAML struct {
	Title          string              `yaml:"title"`
	ItemReferences []ItemReferenceYAML `yaml:"references"`
	AppData        AppDataContentYAML  `yaml:"appData"`
}
type TagContentJSON struct {
	Title          string              `json:"title"`
	ItemReferences []ItemReferenceJSON `json:"references"`
	AppData        AppDataContentJSON  `json:"appData"`
}
type SettingContentYAML struct {
	Title          string              `yaml:"title"`
	ItemReferences []ItemReferenceYAML `yaml:"references"`
	AppData        AppDataContentYAML  `yaml:"appData"`
}
type SettingContentJSON struct {
	Title          string              `json:"title"`
	ItemReferences []ItemReferenceJSON `json:"references"`
	AppData        AppDataContentJSON  `json:"appData"`
}
type NoteContentYAML struct {
	Title          string              `yaml:"title"`
	Text           string              `json:"text"`
	ItemReferences []ItemReferenceYAML `yaml:"references"`
	AppData        AppDataContentYAML  `yaml:"appData"`
}
type NoteContentJSON struct {
	Title          string              `json:"title"`
	Text           string              `json:"text"`
	ItemReferences []ItemReferenceJSON `json:"references"`
	AppData        AppDataContentJSON  `json:"appData"`
}
type TagJSON struct {
	UUID        string         `json:"uuid"`
	Content     TagContentJSON `json:"content"`
	ContentType string         `json:"content_type"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type TagYAML struct {
	UUID        string         `yaml:"uuid"`
	Content     TagContentYAML `yaml:"content"`
	ContentType string         `yaml:"content_type"`
	CreatedAt   string         `yaml:"created_at"`
	UpdatedAt   string         `yaml:"updated_at"`
}

type NoteJSON struct {
	UUID        string          `json:"uuid"`
	Content     NoteContentJSON `json:"content"`
	ContentType string          `json:"content_type"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type SettingYAML struct {
	UUID        string             `yaml:"uuid"`
	Content     SettingContentYAML `yaml:"content"`
	ContentType string             `yaml:"content_type"`
	CreatedAt   string             `yaml:"created_at"`
	UpdatedAt   string             `yaml:"updated_at"`
}

type SettingJSON struct {
	UUID        string             `json:"uuid"`
	Content     SettingContentJSON `json:"content"`
	ContentType string             `json:"content_type"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

type NoteYAML struct {
	UUID        string          `yaml:"uuid"`
	Content     NoteContentYAML `yaml:"content"`
	ContentType string          `yaml:"content_type"`
	CreatedAt   string          `yaml:"created_at"`
	UpdatedAt   string          `yaml:"updated_at"`
}

type TagItemsConfig struct {
	Session    gosn.Session
	FindTitle  string
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
	Session gosn.Session
	Filters gosn.ItemFilters
	Output  string
	Debug   bool
}

type GetSettingsConfig struct {
	Session gosn.Session
	Filters gosn.ItemFilters
	Output  string
	Debug   bool
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
	Regex     bool
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
	NoteText   string
	NoteUUIDs  []string
	Regex      bool
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
	output.Items.DeDupe()
	ei := output.Items
	var di gosn.DecryptedItems
	di, err = ei.Decrypt(input.Session.Mk, input.Session.Ak)
	if err != nil {
		return 0, err
	}
	var pi gosn.Items
	pi, err = di.Parse()
	if err != nil {
		return 0, err
	}
	var itemsToDel gosn.Items
	for _, item := range pi {
		if item.Deleted {
			continue
		}
		switch item.ContentType {
		case "Tag":
			item.Deleted = true
			item.Content = gosn.NewTagContent()
			itemsToDel = append(itemsToDel, item)
		case "Note":
			item.Deleted = true
			item.Content = gosn.NewNoteContent()
			itemsToDel = append(itemsToDel, item)
		}
	}
	// delete items
	var eItemsToDel gosn.EncryptedItems
	eItemsToDel, err = itemsToDel.Encrypt(input.Session.Mk, input.Session.Ak)
	if err != nil {
		return 0, err
	}
	putItemsInput := gosn.PutItemsInput{
		Session:   input.Session,
		Items:     eItemsToDel,
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
	var output gosn.GetItemsOutput
	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return err
	}
	output.Items.DeDupe()
	ei := output.Items
	var di gosn.DecryptedItems
	di, err = ei.Decrypt(input.Session.Mk, input.Session.Ak)
	var pi gosn.Items
	pi, err = di.Parse()
	var missingContentType gosn.Items
	var missingContent gosn.Items
	var notesToTitleFix gosn.Items

	var allIDs []string
	var allItems gosn.Items

	for _, item := range pi {
		allIDs = append(allIDs, item.UUID)
		if !item.Deleted {
			allItems = append(allItems, item)
			switch {
			case item.ContentType == "":
				item.Deleted = true
				item.ContentType = "Note"
				missingContentType = append(missingContentType, item)
			case item.Content == nil && StringInSlice(item.ContentType, []string{"Note", "Tag"}, true):
				item.Deleted = true
				fmt.Println(item.ContentType)
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

	var itemsWithRefsToUpdate gosn.Items
	for _, item := range allItems {
		var newRefs []gosn.ItemReference
		var needsFix bool
		if item.Content != nil && item.Content.References() != nil && len(item.Content.References()) > 0 {
			for _, ref := range item.Content.References() {
				if !StringInSlice(ref.UUID, allIDs, false) {
					needsFix = true
					fmt.Printf("item: %s references missing item id: %s\n", item.Content.GetTitle(), ref.UUID)
				} else {
					newRefs = append(newRefs, ref)
				}
			}
			if needsFix {
				item.Content.SetReferences(newRefs)
				itemsWithRefsToUpdate = append(itemsWithRefsToUpdate, item)
			}
		}

	}

	// fix items with references to missing or deleted items
	if len(itemsWithRefsToUpdate) > 0 {
		fmt.Printf("found %d items with invalid references. fix? ", len(itemsWithRefsToUpdate))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eItemsWithRefsToUpdate gosn.EncryptedItems
			eItemsWithRefsToUpdate, err = itemsWithRefsToUpdate.Encrypt(input.Session.Mk, input.Session.Ak)
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   eItemsWithRefsToUpdate,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed references in %d items\n", len(itemsWithRefsToUpdate))
		}
	} else {
		fmt.Println("no items with invalid references")
	}

	// check for items without content type
	if len(missingContentType) > 0 {
		fmt.Printf("found %d notes with missing content type. delete? ", len(missingContentType))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eMissingContentType gosn.EncryptedItems
			eMissingContentType, err = missingContentType.Encrypt(input.Session.Mk, input.Session.Ak)
			if err != nil {
				return err
			}
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   eMissingContentType,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed %d items\n", len(missingContentType))

		}
	} else {
		fmt.Println("no items with missing content type")
	}

	// check for items with missing content
	if len(missingContent) > 0 {
		fmt.Printf("found %d notes with missing content. delete? ", len(missingContent))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eMissingContent gosn.EncryptedItems
			eMissingContent, err = missingContent.Encrypt(input.Session.Mk, input.Session.Ak)
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   eMissingContent,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed %d items\n", len(missingContent))

		}
	} else {
		fmt.Println("no items with missing content")
	}

	// check for items with missing titles
	if len(notesToTitleFix) > 0 {
		fmt.Printf("found %d items with missing titles. fix? ", len(notesToTitleFix))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eNotesToTitleFix gosn.EncryptedItems
			eNotesToTitleFix, err = notesToTitleFix.Encrypt(input.Session.Mk, input.Session.Ak)
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   eNotesToTitleFix,
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
