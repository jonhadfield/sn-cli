package main

import (
	"errors"
	"fmt"
	"github.com/divan/num2words"
	"github.com/jonhadfield/gosn-v2/cache"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli"
	"strings"
)

func processDeleteNote(c *cli.Context, opts configOptsOutput) (msg string, err error) {
	title := strings.TrimSpace(c.String("title"))
	uuid := strings.TrimSpace(c.String("uuid"))
	if title == "" && uuid == "" {
		cli.ShowSubcommandHelp(c)
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
