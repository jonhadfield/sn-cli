package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

func processGetTags(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	inTitle := strings.TrimSpace(c.String("title"))
	inUUID := strings.TrimSpace(c.String("uuid"))

	var matchAny bool
	if c.Bool("match-all") {
		matchAny = false
	}

	regex := c.Bool("regex")
	count := c.Bool("count")

	getTagsIF := gosn.ItemFilters{
		MatchAny: matchAny,
	}

	// add uuid filters
	if inUUID != "" {
		for _, uuid := range sncli.CommaSplit(inUUID) {
			titleFilter := gosn.Filter{
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
			titleFilter := gosn.Filter{
				Type:       "Tag",
				Key:        "Title",
				Comparison: comparison,
				Value:      title,
			}
			getTagsIF.Filters = append(getTagsIF.Filters, titleFilter)
		}
	}

	if inTitle == "" && inUUID == "" {
		getTagsIF.Filters = append(getTagsIF.Filters, gosn.Filter{
			Type: "Tag",
		})
	}

	session, _, err := cache.GetSession(opts.useSession, opts.sessKey, opts.server)
	if err != nil {
		return "", err
	}

	// TODO: validate output
	output := c.String("output")

	var cacheDBPath string

	cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return "", err
	}

	session.CacheDBPath = cacheDBPath

	appGetTagConfig := sncli.GetTagConfig{
		Session: session,
		Filters: getTagsIF,
		Output:  output,
		Debug:   opts.debug,
	}

	var rawTags gosn.Items

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
				ClientUpdatedAt: rt.(*gosn.Tag).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
			}
			tagContentAppDataContent := sncli.AppDataContentYAML{
				OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailYAML,
			}

			tagContentYAML := sncli.TagContentYAML{
				Title:          rt.(*gosn.Tag).Content.GetTitle(),
				ItemReferences: sncli.ItemRefsToYaml(rt.(*gosn.Tag).Content.References()),
				AppData:        tagContentAppDataContent,
			}

			tagsYAML = append(tagsYAML, sncli.TagYAML{
				UUID:        rt.(*gosn.Tag).UUID,
				ContentType: rt.(*gosn.Tag).ContentType,
				Content:     tagContentYAML,
				UpdatedAt:   rt.(*gosn.Tag).UpdatedAt,
				CreatedAt:   rt.(*gosn.Tag).CreatedAt,
			})
		}

		if !count && strings.ToLower(output) == "json" {
			tagContentOrgStandardNotesSNDetailJSON := sncli.OrgStandardNotesSNDetailJSON{
				ClientUpdatedAt: rt.(*gosn.Tag).Content.GetAppData().OrgStandardNotesSN.ClientUpdatedAt,
			}
			tagContentAppDataContent := sncli.AppDataContentJSON{
				OrgStandardNotesSN: tagContentOrgStandardNotesSNDetailJSON,
			}

			tagContentJSON := sncli.TagContentJSON{
				Title:          rt.(*gosn.Tag).Content.GetTitle(),
				ItemReferences: sncli.ItemRefsToJSON(rt.(*gosn.Tag).Content.References()),
				AppData:        tagContentAppDataContent,
			}

			tagsJSON = append(tagsJSON, sncli.TagJSON{
				UUID:        rt.(*gosn.Tag).UUID,
				ContentType: rt.(*gosn.Tag).ContentType,
				Content:     tagContentJSON,
				UpdatedAt:   rt.(*gosn.Tag).UpdatedAt,
				CreatedAt:   rt.(*gosn.Tag).CreatedAt,
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
			bOutput, err = json.MarshalIndent(tagsJSON, "", "    ")
		case "yaml":
			bOutput, err = yaml.Marshal(tagsYAML)
		}
		if len(bOutput) > 0 {
			fmt.Println(string(bOutput))
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
	session, _, err := cache.GetSession(opts.useSession,
		opts.sessKey, opts.server)
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
		Session: session,
		Tags:    tags,
		Debug:   opts.debug,
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

	session, _, err := cache.GetSession(opts.useSession,
		opts.sessKey, opts.server)
	if err != nil {
		return "", err
	}

	if findText == "" && findTitle == "" && findTag == "" {
		fmt.Println("you must provide either text, title, or tag to search for")
		return "", cli.ShowSubcommandHelp(c)
	}

	processedTags := sncli.CommaSplit(newTags)

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return "", err
	}

	appConfig := sncli.TagItemsConfig{
		Session:    session,
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
	uuidIn := strings.Replace(c.String("uuid"), " ", "", -1)

	if titleIn == "" && uuidIn == "" {
		_ = cli.ShowSubcommandHelp(c)
		return msg, errors.New("title or uuid required")
	}

	session, _, err := cache.GetSession(opts.useSession,
		opts.sessKey, opts.server)
	if err != nil {
		return msg, err
	}

	tags := sncli.CommaSplit(titleIn)
	uuids := sncli.CommaSplit(uuidIn)

	var cacheDBPath string

	cacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return msg, err
	}

	session.CacheDBPath = cacheDBPath

	DeleteTagConfig := sncli.DeleteTagConfig{
		Session:   session,
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
