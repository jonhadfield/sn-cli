package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/items"
	sncli "github.com/jonhadfield/sn-cli"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func processGetItems(c *cli.Context, opts configOptsOutput) (err error) {
	inUUID := strings.TrimSpace(c.String("uuid"))

	matchAny := true
	if c.Bool("match-all") {
		matchAny = false
	}

	getItemsIF := items.ItemFilters{
		Filters: []items.Filter{
			{
				Key:        "uuid",
				Comparison: "==",
				Value:      inUUID,
			},
		},
		MatchAny: matchAny,
	}

	var sess cache.Session

	sess, _, err = cache.GetSession(opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	// TODO: validate output
	output := c.String("output")

	var cacheDBPath string

	cacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	sess.CacheDBPath = cacheDBPath
	appGetItemsConfig := sncli.GetItemsConfig{
		Session: &sess,
		Filters: getItemsIF,
		Output:  output,
		Debug:   opts.debug,
	}

	var rawItems items.Items

	rawItems, err = appGetItemsConfig.Run()
	if err != nil {
		return err
	}

	// strip deleted items
	rawItems = sncli.RemoveDeleted(rawItems)
	numResults := len(rawItems)

	if numResults == 0 {
		// msg = msgNoMatches

		return nil
	}

	output = c.String("output")

	output = "json"
	var bOutput []byte
	switch strings.ToLower(output) {
	case "json":
		bOutput, err = json.MarshalIndent(rawItems, "", "    ")
	case "yaml":
		bOutput, err = yaml.Marshal(rawItems)
	}

	if len(bOutput) > 0 {
		fmt.Print("{\n  \"items\": ")
		fmt.Print(string(bOutput))
		fmt.Print("\n}")
	}

	return err
}
