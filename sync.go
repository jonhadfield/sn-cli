package sncli

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jonhadfield/gosn-v2/cache"
)

func Sync(si cache.SyncInput, showProgress bool) (so cache.SyncOutput, err error) {
	if showProgress && !si.Debug {
		prefix := HiWhite("syncing ")
		if _, err = os.Stat(si.Session.CacheDBPath); os.IsNotExist(err) {
			prefix = HiWhite("initialising ")
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
		s.Prefix = prefix
		s.Start()

		so, err = sync(si)

		s.Stop()

		return
	}

	var a cache.SyncOutput
	a, err = sync(si)
	if err != nil {
		panic(err)
	}
	return a, err
}

func sync(si cache.SyncInput) (so cache.SyncOutput, err error) {
	return cache.Sync(cache.SyncInput{
		Session: si.Session,
		Close:   si.Close,
		Debug:   si.Debug,
	})
}
