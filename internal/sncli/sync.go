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

		so, err := SyncQuiet(si)

		s.Stop()

		return so, err
	}

	return SyncQuiet(si)
}

// syncedWithServer records whether this process has synced with the server yet.
var syncedWithServer bool

// SyncQuiet syncs without a spinner. The first sync in each process always calls
// the API, so changes made in other clients since the last run are seen; later
// syncs in the same run may skip the call when nothing has changed locally.
func SyncQuiet(si cache.SyncInput) (cache.SyncOutput, error) {
	so, err := cache.Sync(cache.SyncInput{
		Session:    si.Session,
		Close:      si.Close,
		AlwaysSync: si.AlwaysSync || !syncedWithServer,
	})
	if err == nil {
		syncedWithServer = true
	}

	return so, err
}
