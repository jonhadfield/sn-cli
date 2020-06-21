package main

import (
	"errors"
	"fmt"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
	"strings"
)
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
