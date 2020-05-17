package sncli

import (
	"github.com/jonhadfield/gosn-v2"
)

const (
	timeLayout  = "2006-01-02T15:04:05.000Z"
	SNServerURL = "https://sync.standardnotes.org"
	SNPageSize  = 600
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

type AddTagsInput struct {
	Session gosn.Session
	Tags    []string
	Debug   bool
}

type AddTagsOutput struct {
	Added, Existing []string
	SyncToken       string
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

type AddNoteInput struct {
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
	Session  gosn.Session
	Debug    bool
	Settings bool
}

type StatsConfig struct {
	Session gosn.Session
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

func filterByTypes(ei gosn.EncryptedItems, types []string) (o gosn.EncryptedItems) {
	for _, i := range ei {
		if StringInSlice(i.ContentType, types, true) {
			o = append(o, i)
		}
	}

	return o
}

var supportedContentTypes = []string{"Note", "Tag", "SN|Component"}

func (input *WipeConfig) Run() (int, error) {
	getItemsInput := gosn.SyncInput{
		Session: input.Session,
		Debug: input.Debug,
	}

	var err error
	// get all existing Tags and Notes and mark for deletion
	var output gosn.SyncOutput

	output, err = gosn.Sync(getItemsInput)
	if err != nil {
		return 0, err
	}

	output.Items = filterByTypes(output.Items, supportedContentTypes)

	output.Items.DeDupe()

	var pi gosn.Items

	pi, err = output.Items.DecryptAndParse(input.Session.Mk, input.Session.Ak, input.Debug)
	if err != nil {
		return 0, err
	}

	var itemsToDel gosn.Items

	for _, item := range pi {
		if item == nil || item.IsDeleted() {
			continue
		}

		switch {
		case item.GetContentType() == "Tag":
			tag := item.(*gosn.Tag)
			tag.SetDeleted(true)
			tag.Content = *gosn.NewTagContent()
			itemsToDel = append(itemsToDel, tag)
		case item.GetContentType() == "Note":
			note := item.(*gosn.Note)
			note.Deleted = true
			note.Content = *gosn.NewNoteContent()
			itemsToDel = append(itemsToDel, note)
		case input.Settings:
			setting := item.(*gosn.Component)
			setting.Deleted = true
			itemsToDel = append(itemsToDel, setting)
		}
	}
	// delete items
	var eItemsToDel gosn.EncryptedItems

	eItemsToDel, err = itemsToDel.Encrypt(input.Session.Mk, input.Session.Ak, input.Debug)
	if err != nil {
		return 0, err
	}

	putItemsInput := gosn.SyncInput{
		Session:   input.Session,
		Items:     eItemsToDel,
		SyncToken: output.SyncToken,
	}

	_, err = gosn.Sync(putItemsInput)
	if err != nil {
		return 0, err
	}

	return len(itemsToDel), err
}
