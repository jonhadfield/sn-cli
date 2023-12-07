package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/asdine/storm/v3/q"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

func getTagByUUID(sess *cache.Session, uuid string) (tag items.Tag, err error) {
	if sess.CacheDBPath == "" {
		return tag, errors.New("CacheDBPath missing from sess")
	}

	if uuid == "" {
		return tag, errors.New("uuid not supplied")
	}

	var so cache.SyncOutput

	si := cache.SyncInput{
		Session: sess,
		Close:   false,
	}

	so, err = cache.Sync(si)
	if err != nil {
		return
	}

	defer func() {
		_ = so.DB.Close()
	}()

	var encTags cache.Items

	query := so.DB.Select(q.And(q.Eq("UUID", uuid), q.Eq("Deleted", false)))
	if err = query.Find(&encTags); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return tag, errors.New(fmt.Sprintf("could not find tag with UUID %s", uuid))
		}

		return
	}

	var rawEncItems items.Items
	rawEncItems, err = encTags.ToItems(sess)

	return *rawEncItems[0].(*items.Tag), err
}

func getTagsByTitle(sess cache.Session, title string) (tags items.Tags, err error) {
	var so cache.SyncOutput

	si := cache.SyncInput{
		Session: &sess,
		Close:   false,
	}

	so, err = cache.Sync(si)
	if err != nil {
		return
	}

	defer func() {
		_ = so.DB.Close()
	}()

	var allEncTags cache.Items

	query := so.DB.Select(q.And(q.Eq("ContentType", "Tag"), q.Eq("Deleted", false)))

	err = query.Find(&allEncTags)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("could not find any tags")
		}

		return
	}

	// decrypt all tags
	var allRawTags items.Items
	allRawTags, err = allEncTags.ToItems(&sess)

	var matchingRawTags items.Tags

	for _, rt := range allRawTags {
		t := rt.(*items.Tag)
		if t.Content.Title == title {
			matchingRawTags = append(matchingRawTags, *t)
		}
	}

	return matchingRawTags, err
}

func processEditTag(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	inUUID := c.String("uuid")
	inTitle := c.String("title")

	if inTitle == "" && inUUID == "" || inTitle != "" && inUUID != "" {
		_ = cli.ShowSubcommandHelp(c)

		return "", errors.New("title or UUID is required")
	}

	var sess cache.Session

	sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return "", err
	}

	var cacheDBPath string

	cacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
	if err != nil {
		return "", err
	}

	sess.CacheDBPath = cacheDBPath

	var tag items.Tag

	var tags items.Tags

	// if uuid was passed then retrieve tag from db using uuid
	if inUUID != "" {
		if tag, err = getTagByUUID(&sess, inUUID); err != nil {
			return
		}
	}

	// if title was passed then retrieve tag(s) matching that title
	if inTitle != "" {
		if tags, err = getTagsByTitle(sess, inTitle); err != nil {
			return
		}

		if len(tags) == 0 {
			return "", errors.New("tag not found")
		}

		if len(tags) > 1 {
			return "", errors.New("multiple tags found with same title")
		}

		tag = tags[0]
	}

	// only show existing title information if uuid was passed
	if inUUID != "" {
		fmt.Printf("existing title: %s\n", tag.Content.Title)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("new title: ")

	text, _ := reader.ReadString('\n')

	text = strings.TrimSuffix(text, "\n")
	if len(text) == 0 {
		return "", errors.New("new tag title not entered")
	}

	tag.Content.Title = text

	si := cache.SyncInput{
		Session: &sess,
		Close:   false,
	}

	var so cache.SyncOutput

	so, err = cache.Sync(si)
	if err != nil {
		return
	}

	tags = items.Tags{tag}

	if err = cache.SaveTags(so.DB, &sess, tags, false); err != nil {
		return
	}

	if _, err = cache.Sync(si); err != nil {
		return
	}

	return "", err
}

func processGetTags(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	inTitle := strings.TrimSpace(c.String("title"))
	inUUID := strings.TrimSpace(c.String("uuid"))

	var matchAny bool
	if c.Bool("match-all") {
		matchAny = false
	}

	regex := c.Bool("regex")
	count := c.Bool("count")

	getTagsIF := items.ItemFilters{
		MatchAny: matchAny,
	}

	// add uuid filters
	if inUUID != "" {
		for _, uuid := range sncli.CommaSplit(inUUID) {
			titleFilter := items.Filter{
				Type:       "Tag",
				Key:        "uuid",
				Comparison: "==",
				Value:      uuid,
			}
			getTagsIF.Filters = append(getTagsIF.Filters, titleFilter)
		}
	}

	comparison := "contains"
	if regex {
		comparison = "~"
	}

	if inTitle != "" {
		for _, title := range sncli.CommaSplit(inTitle) {
			titleFilter := items.Filter{
				Type:       "Tag",
				Key:        "Title",
				Comparison: comparison,
				Value:      title,
			}
			getTagsIF.Filters = append(getTagsIF.Filters, titleFilter)
		}
	}

	if inTitle == "" && inUUID == "" {
		getTagsIF.Filters = append(getTagsIF.Filters, items.Filter{
			Type: "Tag",
		})
	}

	var sess cache.Session

	sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return "", err
	}

	// TODO: validate output
	output := c.String("output")

	var cacheDBPath string

	cacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
	if err != nil {
		return "", err
	}

	sess.CacheDBPath = cacheDBPath

	appGetTagConfig := sncli.GetTagConfig{
		Session: &sess,
		Filters: getTagsIF,
		Output:  output,
		Debug:   opts.debug,
	}

	var rawTags items.Items

	rawTags, err = appGetTagConfig.Run()
	if err != nil {
		return "", err
	}

	// strip deleted items
	rawTags = sncli.RemoveDeleted(rawTags)

	var tagsYAML []sncli.TagYAML

	var tagsJSON []sncli.TagJSON

	var numResults int

	for _, rt := range rawTags {
		numResults++

		if !count && sncli.StringInSlice(output, yamlAbbrevs, false) {
			tagContentOrgStandardNotesSNDetailYAML := sncli.OrgStandardNotesSNDetailYAML{
				ClientUpdatedAt: rt.(*items.Tag).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
			}
			tagContentAppDataContent := sncli.AppDataContentYAML{
				OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailYAML,
			}

			tagContentYAML := sncli.TagContentYAML{
				Title:          rt.(*items.Tag).Content.GetTitle(),
				ItemReferences: sncli.ItemRefsToYaml(rt.(*items.Tag).Content.References()),
				AppData:        tagContentAppDataContent,
			}

			tagsYAML = append(tagsYAML, sncli.TagYAML{
				UUID:        rt.(*items.Tag).UUID,
				ContentType: rt.(*items.Tag).ContentType,
				Content:     tagContentYAML,
				UpdatedAt:   rt.(*items.Tag).UpdatedAt,
				CreatedAt:   rt.(*items.Tag).CreatedAt,
			})
		}

		if !count && strings.ToLower(output) == "json" {
			tagContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
				ClientUpdatedAt: rt.(*items.Tag).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
			}
			tagContentAppDataContent := sncli.AppDataContentJSON{
				OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailJSON,
			}

			tagContentJSON := sncli.TagContentJSON{
				Title:          rt.(*items.Tag).Content.GetTitle(),
				ItemReferences: sncli.ItemRefsToJSON(rt.(*items.Tag).Content.References()),
				AppData:        tagContentAppDataContent,
			}

			tagsJSON = append(tagsJSON, sncli.TagJSON{
				UUID:        rt.(*items.Tag).UUID,
				ContentType: rt.(*items.Tag).ContentType,
				Content:     tagContentJSON,
				UpdatedAt:   rt.(*items.Tag).UpdatedAt,
				CreatedAt:   rt.(*items.Tag).CreatedAt,
			})
		}
	}
	if !opts.useStdOut {
		return
	} else if numResults <= 0 {
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
			bOutput, err = json.MarshalIndent(tagsJSON, "", "    ")
		case "yaml":
			bOutput, err = yaml.Marshal(tagsYAML)
		}

		if len(bOutput) > 0 {
			fmt.Print("{\n  \"tags\": ")
			fmt.Print(string(bOutput))
			fmt.Print("\n}")
		}
	}
	return msg, err
}

func processAddTags(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	// validate input
	tagInput := c.String("title")
	if strings.TrimSpace(tagInput) == "" {
		_ = cli.ShowSubcommandHelp(c)

		return "", errors.New("tag title not defined")
	}

	// get session
	session, _, err := cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return "", err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return "", err
	}

	// prepare input
	tags := sncli.CommaSplit(tagInput)
	addTagInput := sncli.AddTagsInput{
		Session:    &session,
		Tags:       tags,
		Parent:     c.String("parent"),
		ParentUUID: c.String("parent-uuid"),
		Debug:      opts.debug,
	}

	// attempt to add tags
	var ato sncli.AddTagsOutput

	ato, err = addTagInput.Run()
	if err != nil {
		return "", fmt.Errorf(sncli.Red(err))
	}

	// present results
	if len(ato.Added) > 0 {
		msg = sncli.Green(msgAddSuccess+": ", strings.Join(ato.Added, ", "))
	}

	if len(ato.Existing) > 0 {
		// add line break if output already added
		if len(msg) > 0 {
			msg += "\n"
		}

		msg += sncli.Yellow(msgAlreadyExisting + ": " + strings.Join(ato.Existing, ", "))
	}

	return msg, err
}

func processTagItems(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	findTitle := c.String("find-title")
	findText := c.String("find-text")
	findTag := c.String("find-tag")
	newTags := c.String("title")

	sess, _, err := cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return "", err
	}

	if findText == "" && findTitle == "" && findTag == "" {
		fmt.Println("you must provide either text, title, or tag to search for")

		return "", cli.ShowSubcommandHelp(c)
	}

	processedTags := sncli.CommaSplit(newTags)

	sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
	if err != nil {
		return "", err
	}

	appConfig := sncli.TagItemsConfig{
		Session:    &sess,
		FindText:   findText,
		FindTitle:  findTitle,
		FindTag:    findTag,
		NewTags:    processedTags,
		Replace:    c.Bool("replace"),
		IgnoreCase: c.Bool("ignore-case"),
		Debug:      opts.debug,
	}

	err = appConfig.Run()
	if err != nil {
		return "", err
	}

	msg = msgTagSuccess

	return msg, err
}

func processDeleteTags(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	titleIn := strings.TrimSpace(c.String("title"))
	uuidIn := strings.ReplaceAll(c.String("uuid"), " ", "")

	if titleIn == "" && uuidIn == "" {
		_ = cli.ShowSubcommandHelp(c)

		return msg, errors.New("title or uuid required")
	}

	var sess cache.Session

	sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return msg, err
	}

	tags := sncli.CommaSplit(titleIn)
	uuids := sncli.CommaSplit(uuidIn)

	var cacheDBPath string

	cacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
	if err != nil {
		return msg, err
	}

	sess.CacheDBPath = cacheDBPath

	DeleteTagConfig := sncli.DeleteTagConfig{
		Session:   &sess,
		TagTitles: tags,
		TagUUIDs:  uuids,
		Debug:     opts.debug,
	}

	var noDeleted int

	noDeleted, err = DeleteTagConfig.Run()
	if err != nil {
		return msg, fmt.Errorf("failed to delete tag. %+v", err)
	}

	if noDeleted > 0 {
		msg = sncli.Green(fmt.Sprintf("%s tag", msgDeleted))
	} else {
		msg = sncli.Yellow("Tag not found")
	}

	return msg, err
}
