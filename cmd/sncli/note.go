package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/divan/num2words"
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

func processGetNotes(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	uuid := c.String("uuid")
	title := c.String("title")
	text := c.String("text")
	count := c.Bool("count")
	output := c.String("output")
	noteFilter := gosn.Filter{
		Type: "Note",
	}
	getNotesIF := gosn.ItemFilters{
		MatchAny: false,
		Filters:  []gosn.Filter{noteFilter},
	}

	if uuid != "" {
		titleFilter := gosn.Filter{
			Type:       "Note",
			Key:        "uuid",
			Comparison: "==",
			Value:      uuid,
		}
		getNotesIF.Filters = append(getNotesIF.Filters, titleFilter)
	}
	if title != "" {
		titleFilter := gosn.Filter{
			Type:       "Note",
			Key:        "Title",
			Comparison: "contains",
			Value:      title,
		}
		getNotesIF.Filters = append(getNotesIF.Filters, titleFilter)
	}
	if text != "" {
		titleFilter := gosn.Filter{
			Type:       "Note",
			Key:        "Text",
			Comparison: "contains",
			Value:      text,
		}
		getNotesIF.Filters = append(getNotesIF.Filters, titleFilter)
	}
	processedTags := sncli.CommaSplit(c.String("tag"))

	if len(processedTags) > 0 {
		for _, t := range processedTags {
			titleFilter := gosn.Filter{
				Type:       "Note",
				Key:        "Tag",
				Comparison: "contains",
				Value:      t,
			}
			getNotesIF.Filters = append(getNotesIF.Filters, titleFilter)
		}
	}

	session, _, err := cache.GetSession(opts.useSession,
		opts.sessKey, opts.server)
	if err != nil {
		return "", err
	}

	var cacheDBPath string
	cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return "", err
	}

	session.CacheDBPath = cacheDBPath

	getNoteConfig := sncli.GetNoteConfig{
		Session: session,
		Filters: getNotesIF,
		Debug:   opts.debug,
	}

	var rawNotes gosn.Items
	rawNotes, err = getNoteConfig.Run()
	if err != nil {
		return "", err
	}
	// strip deleted items
	rawNotes = sncli.RemoveDeleted(rawNotes)

	var numResults int
	var notesYAML []sncli.NoteYAML
	var notesJSON []sncli.NoteJSON
	for _, rt := range rawNotes {
		numResults++
		if !count && sncli.StringInSlice(output, yamlAbbrevs, false) {
			noteContentOrgStandardNotesSNDetailYAML := sncli.OrgStandardNotesSNDetailYAML{
				ClientUpdatedAt: rt.(*gosn.Note).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
			}
			noteContentAppDataContent := sncli.AppDataContentYAML{
				OrgStandardNotesSN: noteContentOrgStandardNotesSNDetailYAML,
			}
			noteContentYAML := sncli.NoteContentYAML{
				Title:          rt.(*gosn.Note).Content.GetTitle(),
				Text:           rt.(*gosn.Note).Content.GetText(),
				ItemReferences: sncli.ItemRefsToYaml(rt.(*gosn.Note).Content.References()),
				AppData:        noteContentAppDataContent,
			}

			notesYAML = append(notesYAML, sncli.NoteYAML{
				UUID:        rt.(*gosn.Note).UUID,
				ContentType: rt.(*gosn.Note).ContentType,
				Content:     noteContentYAML,
				UpdatedAt:   rt.(*gosn.Note).UpdatedAt,
				CreatedAt:   rt.(*gosn.Note).CreatedAt,
			})
		}
		if !count && strings.ToLower(output) == "json" {
			noteContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
				ClientUpdatedAt: rt.(*gosn.Note).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
			}
			noteContentAppDataContent := sncli.AppDataContentJSON{
				OrgStandardNotesSN: noteContentOrgStandardNotesSNDetailJSON,
			}
			noteContentJSON := sncli.NoteContentJSON{
				Title:          rt.(*gosn.Note).Content.GetTitle(),
				Text:           rt.(*gosn.Note).Content.GetText(),
				ItemReferences: sncli.ItemRefsToJSON(rt.(*gosn.Note).Content.References()),
				AppData:        noteContentAppDataContent,
			}

			notesJSON = append(notesJSON, sncli.NoteJSON{
				UUID:        rt.(*gosn.Note).UUID,
				ContentType: rt.(*gosn.Note).ContentType,
				Content:     noteContentJSON,
				UpdatedAt:   rt.(*gosn.Note).UpdatedAt,
				CreatedAt:   rt.(*gosn.Note).CreatedAt,
			})
		}
	}

	if numResults <= 0 {
		if count {
			msg = "0"
		} else {
			msg = msgNoMatches
		}
	} else if count {
		msg = strconv.Itoa(numResults)
	} else {
		output = c.String("output")
		var bOutput []byte
		switch strings.ToLower(output) {
		case "json":
			bOutput, err = json.MarshalIndent(notesJSON, "", "    ")
		case "yaml":
			bOutput, err = yaml.Marshal(notesYAML)
		}
		if len(bOutput) > 0 {
			fmt.Println(string(bOutput))
		}
	}

	return msg, err
}

func processAddNotes(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	// get input
	title := c.String("title")
	text := c.String("text")
	if strings.TrimSpace(title) == "" {
		if cErr := cli.ShowSubcommandHelp(c); cErr != nil {
			panic(cErr)
		}

		return "", errors.New("note title not defined")
	}
	if strings.TrimSpace(text) == "" {
		_ = cli.ShowSubcommandHelp(c)
		return "", errors.New("note text not defined")
	}

	// get session
	session, _, err := cache.GetSession(opts.useSession,
		opts.sessKey, opts.server)
	if err != nil {
		return "", err
	}

	processedTags := sncli.CommaSplit(c.String("tag"))

	var cacheDBPath string
	cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return "", err
	}
	session.CacheDBPath = cacheDBPath
	AddNoteInput := sncli.AddNoteInput{
		Session: session,
		Title:   title,
		Text:    text,
		Tags:    processedTags,
		Replace: false,
		Debug:   opts.debug,
	}
	if err = AddNoteInput.Run(); err != nil {
		return "", fmt.Errorf("failed to add note. %+v", err)
	}

	msg = sncli.Green(msgAddSuccess + " note")

	return msg, err
}

func processDeleteNote(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	title := strings.TrimSpace(c.String("title"))
	uuid := strings.TrimSpace(c.String("uuid"))
	if title == "" && uuid == "" {
		_ = cli.ShowSubcommandHelp(c)
		return "", errors.New("")
	}
	session, _, err := cache.GetSession(opts.useSession,
		opts.sessKey, opts.server)
	if err != nil {
		return msg, err
	}
	processedNotes := sncli.CommaSplit(title)

	var cacheDBPath string
	cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return msg, err
	}
	session.CacheDBPath = cacheDBPath
	DeleteNoteConfig := sncli.DeleteNoteConfig{
		Session:    session,
		NoteTitles: processedNotes,
		Debug:      opts.debug,
	}
	var noDeleted int
	if noDeleted, err = DeleteNoteConfig.Run(); err != nil {
		return msg, fmt.Errorf("failed to delete note. %+v", err)
	}

	strNote := "notes"
	if noDeleted == 1 {
		strNote = "note"
	}

	msg = sncli.Green(fmt.Sprintf("%s %s %s", msgDeleted, num2words.Convert(noDeleted), strNote))
	return msg, err
}
