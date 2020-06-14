package sncli

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/jonhadfield/gosn-v2/cache"
	"os"
	"time"
)

func Sync(si cache.SyncInput, showProgress bool) (so cache.SyncOutput, err error) {
	if showProgress && ! si.Debug {
		prefix := "syncing "
		if _, err := os.Stat(si.Session.CacheDBPath); os.IsNotExist(err) {
			prefix = "initialising "
		}

		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond) // Build our new spinner
		s.Prefix = prefix
		s.Start()
		so, err = sync(si)
		s.Stop()
		return
	}

	return sync(si)
}

func sync(si cache.SyncInput) (so cache.SyncOutput, err error) {
	return cache.Sync(cache.SyncInput{
		Session: si.Session,
		Close:   si.Close,
		Debug:   si.Debug,
	})
}
