package sncli

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
)

const (
	timeLayout        = "2006-01-02T15:04:05.000Z"
	SNServerURL       = "https://api.standardnotes.com"
	SNPageSize        = 600
	SNAppName         = "sn-cli"
	MinPasswordLength = 8
)

type ItemReferenceYAML struct {
	UUID          string `yaml:"uuid"`
	ContentType   string `yaml:"content_type"`
	ReferenceType string `yaml:"reference_type",omitempty`
}

type ItemReferenceJSON struct {
	UUID          string `json:"uuid"`
	ContentType   string `json:"content_type"`
	ReferenceType string `json:"reference_type",omitempty`
}

type OrgStandardNotesSNDetailJSON struct {
	ClientUpdatedAt    string `json:"client_updated_at"`
	PrefersPlainEditor bool   `json:"prefersPlainEditor"`
	Pinned             bool   `json:"pinned"`
}

type OrgStandardNotesSNComponentsDetailJSON map[string]interface{}

type OrgStandardNotesSNDetailYAML struct {
	ClientUpdatedAt string `yaml:"client_updated_at"`
}

type AppDataContentYAML struct {
	OrgStandardNotesSN           OrgStandardNotesSNDetailYAML             `yaml:"org.standardnotes.sn"`
	OrgStandardNotesSNComponents items.OrgStandardNotesSNComponentsDetail `yaml:"org.standardnotes.sn.components,omitempty"`
}

type AppDataContentJSON struct {
	OrgStandardNotesSN           OrgStandardNotesSNDetailJSON             `json:"org.standardnotes.sn"`
	OrgStandardNotesSNComponents items.OrgStandardNotesSNComponentsDetail `json:"org.standardnotes.sn.components,omitempty"`
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
	Title            string              `yaml:"title"`
	Text             string              `json:"text"`
	ItemReferences   []ItemReferenceYAML `yaml:"references"`
	AppData          AppDataContentYAML  `yaml:"appData"`
	EditorIdentifier string              `yaml:"editorIdentifier"`
	PreviewPlain     string              `yaml:"preview_plain"`
	PreviewHtml      string              `yaml:"preview_html"`
	Spellcheck       bool                `yaml:"spellcheck"`
	Trashed          *bool               `yaml:"trashed,omitempty"`
}

type NoteContentJSON struct {
	Title            string              `json:"title"`
	Text             string              `json:"text"`
	ItemReferences   []ItemReferenceJSON `json:"references"`
	AppData          AppDataContentJSON  `json:"appData"`
	EditorIdentifier string              `json:"editorIdentifier"`
	PreviewPlain     string              `json:"preview_plain"`
	PreviewHtml      string              `json:"preview_html"`
	Spellcheck       bool                `json:"spellcheck"`
	Trashed          *bool               `json:"trashed,omitempty"`
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

type GetItemsConfig struct {
	Session *cache.Session
	Filters items.ItemFilters
	Output  string
	Debug   bool
}

type AddTagsInput struct {
	Session    *cache.Session
	Tags       []string
	Parent     string
	ParentUUID string
	Debug      bool
	Replace    bool
}

type AddTagsOutput struct {
	Added, Existing []string
}

type GetTagConfig struct {
	Session *cache.Session
	Filters items.ItemFilters
	Output  string
	Debug   bool
}

type GetSettingsConfig struct {
	Session *cache.Session
	Filters items.ItemFilters
	Output  string
	Debug   bool
}

type GetNoteConfig struct {
	Session    *cache.Session
	Filters    items.ItemFilters
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

type ListChecklistsInput struct {
	Session *cache.Session
	Debug   bool
}

type ShowChecklistInput struct {
	Session *cache.Session
	Title   string
	UUID    string
	Debug   bool
}

type AddNoteInput struct {
	Session  *cache.Session
	Title    string
	Text     string
	FilePath string
	Tags     []string
	Replace  bool
	Debug    bool
}

type DeleteItemConfig struct {
	Session    *cache.Session
	NoteTitles []string
	NoteText   string
	ItemsUUIDs []string
	Regex      bool
	Debug      bool
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
	Session    *cache.Session
	UseStdOut  bool
	Debug      bool
	Everything bool
}

type StatsConfig struct {
	Session cache.Session
}

func referenceExists(tag items.Tag, refID string) bool {
	for _, ref := range tag.Content.References() {
		if ref.UUID == refID {
			return true
		}
	}

	return false
}

func filterEncryptedItemsByTypes(ei items.EncryptedItems, types []string) items.EncryptedItems {
	var o items.EncryptedItems
	for _, i := range ei {
		if StringInSlice(i.ContentType, types, true) {
			o = append(o, i)
		}
	}

	return o
}

func filterItemsByTypes(ei items.Items, types []string) items.Items {
	var o items.Items
	for _, i := range ei {
		if StringInSlice(i.GetContentType(), types, true) {
			o = append(o, i)
		}
	}

	return o
}

func filterCacheItemsByTypes(ei cache.Items, types []string) cache.Items {
	var o cache.Items
	for _, i := range ei {
		if StringInSlice(i.ContentType, types, true) {
			o = append(o, i)
		}
	}

	return o
}

var supportedContentTypes = []string{common.SNItemTypeNote, common.SNItemTypeTag, common.SNItemTypeComponent}

func (i *WipeConfig) Run() (int, error) {
	i.Session.RemoveDB()
	if !i.Session.Debug && i.UseStdOut {
		prefix := color.HiWhite.Sprintf("wiping ")

		s := spinner.New(spinner.CharSets[14], spinnerDelay, spinner.WithWriter(os.Stdout))
		if i.UseStdOut {
			s = spinner.New(spinner.CharSets[14], spinnerDelay, spinner.WithWriter(os.Stderr))
		}

		s.Prefix = prefix
		s.Start()

		deleted, err := items.DeleteContent(i.Session.Session, i.Everything)

		s.Stop()

		return deleted, err
	}

	return items.DeleteContent(i.Session.Session, i.Everything)
}
