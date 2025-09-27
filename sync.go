package sncli

import (
	"os"

	"github.com/briandowns/spinner"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
)

func Sync(si cache.SyncInput, useStdErr bool) (cache.SyncOutput, error) {
	if !si.Debug {
		prefix := color.HiWhite.Sprintf("syncing ")
		if _, err := os.Stat(si.Session.CacheDBPath); os.IsNotExist(err) {
			prefix = color.HiWhite.Sprintf("initializing ")
		}

		s := spinner.New(spinner.CharSets[14], spinnerDelay, spinner.WithWriter(os.Stdout))
		if useStdErr {
			s = spinner.New(spinner.CharSets[14], spinnerDelay, spinner.WithWriter(os.Stderr))
		}

		s.Prefix = prefix
		s.Start()

		so, err := sync(si)

		s.Stop()

		return so, err
	}

	return sync(si)
}

func sync(si cache.SyncInput) (cache.SyncOutput, error) {
	return cache.Sync(cache.SyncInput{
		Session: si.Session,
		Close:   si.Close,
	})
}
