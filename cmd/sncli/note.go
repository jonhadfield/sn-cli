package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asdine/storm/v3/q"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/divan/num2words"
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

func getNoteByUUID(sess cache.Session, uuid string, debug bool) (tag gosn.Note, err error) {
	if sess.CacheDBPath == "" {
		return tag, errors.New("CacheDBPath missing from sess")
	}

	if uuid == "" {
		return tag, errors.New("uuid not supplied")
	}

	var so cache.SyncOutput

	si := cache.SyncInput{
		Session: sess,
		Debug:   debug,
		Close:   false,
	}

	so, err = cache.Sync(si)
	if err != nil {
		return
	}

	defer func() {
		_ = so.DB.Close()
	}()

	var encNotes cache.Items

	query := so.DB.Select(q.And(q.Eq("UUID", uuid), q.Eq("Deleted", false)))
	if err = query.Find(&encNotes); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return tag, errors.New(fmt.Sprintf("could not find note with inUUID %s", uuid))
		}

		return
	}

	var rawEncItems gosn.Items
	rawEncItems, err = encNotes.ToItems(sess.Mk, sess.Ak)

	return *rawEncItems[0].(*gosn.Note), err
}

func getNotesByTitle(sess cache.Session, title string, debug bool) (notes gosn.Notes, err error) {
	var so cache.SyncOutput

	si := cache.SyncInput{
		Session: sess,
		Debug:   debug,
		Close:   false,
	}

	so, err = cache.Sync(si)
	if err != nil {
		return
	}

	defer func() {
		_ = so.DB.Close()
	}()

	var allEncNotes cache.Items

	query := so.DB.Select(q.And(q.Eq("ContentType", "Note"), q.Eq("Deleted", false)))
	if err = query.Find(&allEncNotes); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("could not find any notes")
		}

		return
	}

	// decrypt all notes
	var allRawNotes gosn.Items
	allRawNotes, err = allEncNotes.ToItems(sess.Mk, sess.Ak)

	var matchingRawNotes gosn.Notes

	for _, rt := range allRawNotes {
		t := rt.(*gosn.Note)
		if t.Content.Title == title {
			matchingRawNotes = append(matchingRawNotes, *t)
		}
	}

	return matchingRawNotes, err
}

func openInEditor(filename, editor string) error {
	if editor == "" {
		return errors.New("could not detect editor")
	}

	// Get the full executable path for the editor.
	executable, err := exec.LookPath(editor)
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func captureInputFromEditor(title, text, editor string) ([]byte, error) {
	file, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		return []byte{}, err
	}

	filename := file.Name()

	// write existing content
	_, err = io.WriteString(file, title+"\n"+text)
	if err != nil {
		return nil, err
	}

	// defer removal in case any of the next steps fail
	defer func() {
		_ = os.Remove(filename)
	}()

	if err = openInEditor(filename, editor); err != nil {
		return []byte{}, err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}

	// overwrite temporary file content in case deferred remove fails
	_, err = io.WriteString(file, "-")
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func processEditNote(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	inUUID := c.String("uuid")
	inTitle := c.String("title")
	inEditor := c.String("editor")

	if inTitle == "" && inUUID == "" || inTitle != "" && inUUID != "" {
		_ = cli.ShowSubcommandHelp(c)
		return "", errors.New("title or UUID is required")
	}

	var session cache.Session
	session, _, err = cache.GetSession(opts.useSession,
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

	var note gosn.Note

	var notes gosn.Notes

	// if uuid was passed then retrieve note from db using uuid
	if inUUID != "" {
		if note, err = getNoteByUUID(session, inUUID, opts.debug); err != nil {
			return
		}
	}

	// if title was passed then retrieve note(s) matching that title
	if inTitle != "" {
		if notes, err = getNotesByTitle(session, inTitle, opts.debug); err != nil {
			return
		}

		if len(notes) == 0 {
			return "", errors.New("note not found")
		}

		if len(notes) > 1 {
			return "", errors.New("multiple notes found with same title")
		}

		note = notes[0]
	}

	var b []byte

	b, err = captureInputFromEditor(note.Content.Title, note.Content.Text, inEditor)
	if err != nil {
		return "", err
	}

	var newTitle, newText string

	newTitle, newText, err = parseEditorOutput(b)
	if err != nil {
		return
	}

	note.Content.Title = newTitle
	note.Content.Text = newText

	// sync note content update
	si := cache.SyncInput{
		Session: session,
		Debug:   opts.debug,
		Close:   false,
	}

	var so cache.SyncOutput

	so, err = cache.Sync(si)
	if err != nil {
		return
	}

	notes = gosn.Notes{note}
	if err = cache.SaveNotes(so.DB, session.Mk, session.Ak, notes, true, opts.debug); err != nil {
		return
	}

	if _, err = cache.Sync(si); err != nil {
		return
	}

	return "", err
}

func parseEditorOutput(in []byte) (title, text string, err error) {
	lines := strings.Split(string(in), "\n")

	if len(lines) == 0 || len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		err = errors.New("no text saved")
		return
	}

	title = lines[0]

	if len(lines) >= 1 {
		text = strings.Join(lines[1:], "\n")
	}

	return
}

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

	sess, _, err := cache.GetSession(opts.useSession,
		opts.sessKey, opts.server)
	if err != nil {
		return msg, err
	}

	processedNotes := sncli.CommaSplit(title)

	var cacheDBPath string
	cacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)

	if err != nil {
		return msg, err
	}

	sess.CacheDBPath = cacheDBPath
	DeleteNoteConfig := sncli.DeleteNoteConfig{
		Session:    sess,
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
