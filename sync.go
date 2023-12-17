package sncli

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
)

func Sync(si cache.SyncInput, useStdErr bool) (so cache.SyncOutput, err error) {
	if !si.Debug {
		prefix := color.HiWhite.Sprintf("syncing ")
		if _, err = os.Stat(si.Session.CacheDBPath); os.IsNotExist(err) {
			prefix = color.HiWhite.Sprintf("initializing ")
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stdout))
		if useStdErr {
			s = spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
		}

		s.Prefix = prefix
		s.Start()

		so, err = sync(si)

		s.Stop()

		return
	}

	so, err = sync(si)

	return
}

func sync(si cache.SyncInput) (so cache.SyncOutput, err error) {
	return cache.Sync(cache.SyncInput{
		Session: si.Session,
		Close:   si.Close,
	})
}
