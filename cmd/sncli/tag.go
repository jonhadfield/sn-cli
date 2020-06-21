package main

import (
	"errors"
	"fmt"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
	"strings"
)

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
		cli.ShowSubcommandHelp(c)
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
