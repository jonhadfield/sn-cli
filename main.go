package sncli

import (
	"time"

	gosn "github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
)

const (
	timeLayout  = "2006-01-02T15:04:05.000Z"
	SNServerURL = "https://sync.standardnotes.org"
	SNPageSize  = 600
	SNAppName   = "sn-cli"
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
	Session    *cache.Session
	FindTitle  string
	FindText   string
	FindTag    string
	NewTags    []string
	Replace    bool
	IgnoreCase bool
	Debug      bool
}

type AddTagsInput struct {
	Session *cache.Session
	Tags    []string
	Debug   bool
}

type AddTagsOutput struct {
	Added, Existing []string
}

type GetTagConfig struct {
	Session *cache.Session
	Filters gosn.ItemFilters
	Output  string
	Debug   bool
}

type GetSettingsConfig struct {
	Session *cache.Session
	Filters gosn.ItemFilters
	Output  string
	Debug   bool
}

type GetNoteConfig struct {
	Session    *cache.Session
	Filters    gosn.ItemFilters
	NoteTitles []string
	TagTitles  []string
	TagUUIDs   []string
	PageSize   int
	BatchSize  int
	Debug      bool
}

type DeleteTagConfig struct {
	Session   *cache.Session
	Email     string
	TagTitles []string
	TagUUIDs  []string
	Regex     bool
	Debug     bool
}

type AddNoteInput struct {
	Session *cache.Session
	Title   string
	Text    string
	Tags    []string
	Replace bool
	Debug   bool
}

type DeleteNoteConfig struct {
	Session    *cache.Session
	NoteTitles []string
	NoteText   string
	NoteUUIDs  []string
	Regex      bool
	Debug      bool
}

type WipeConfig struct {
	Session  *cache.Session
	Debug    bool
	Settings bool
}

type StatsConfig struct {
	Session cache.Session
	Debug   bool
}

func referenceExists(tag gosn.Tag, refID string) bool {
	for _, ref := range tag.Content.References() {
		if ref.UUID == refID {
			return true
		}
	}

	return false
}

func filterEncryptedItemsByTypes(ei gosn.EncryptedItems, types []string) (o gosn.EncryptedItems) {
	for _, i := range ei {
		if StringInSlice(i.ContentType, types, true) {
			o = append(o, i)
		}
	}

	return o
}

func filterItemsByTypes(ei gosn.Items, types []string) (o gosn.Items) {
	for _, i := range ei {
		if StringInSlice(i.GetContentType(), types, true) {
			o = append(o, i)
		}
	}

	return o
}

func filterCacheItemsByTypes(ei cache.Items, types []string) (o cache.Items) {
	for _, i := range ei {
		if StringInSlice(i.ContentType, types, true) {
			o = append(o, i)
		}
	}

	return o
}

var supportedContentTypes = []string{"Note", "Tag", "SN|Component"}

func (i *WipeConfig) Run() (int, error) {
	syncInput := cache.SyncInput{
		Session: i.Session,
		Debug:   i.Debug,
	}

	var err error
	// get all existing Tags and Notes and mark for deletion
	var so cache.SyncOutput

	so, err = Sync(syncInput, true)
	if err != nil {
		return 0, err
	}

	// get all items
	var allPersistedItems cache.Items
	err = so.DB.All(&allPersistedItems)

	if err != nil {
		return 0, err
	}

	filteredItems := filterCacheItemsByTypes(allPersistedItems, supportedContentTypes)

	var itemsToDel int

	for _, fi := range filteredItems {
		if fi.ContentType == "SN|ItemsKey" {
			panic("attempted to delete SN|ItemsKey")
		}
		itemsToDel++

		fi.Deleted = true
		fi.Dirty = true
		fi.DirtiedDate = time.Now()

		err = so.DB.Save(&fi)
		if err != nil {
			return 0, err
		}
	}

	// TODO: Close DB after each Sync?
	err = so.DB.Close()
	if err != nil {
		return 0, err
	}

	so, err = Sync(syncInput, true)
	if err != nil {
		return 0, err
	}

	err = so.DB.Close()

	return itemsToDel, err
}
