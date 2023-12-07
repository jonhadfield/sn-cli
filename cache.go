package sncli

import (
	"errors"
	"fmt"
	"github.com/jonhadfield/gosn-v2/cache"
	"os"
)

func Resync(s *cache.Session, cacheDBDir, appName string) error {
	var err error

	// check if cache db dir
	if cacheDBDir != "" {
		_, err = os.Stat(s.CacheDBPath)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.New("specified cache directory does not exist")
			}

			return err
		}
	}

	s.CacheDBPath, err = cache.GenCacheDBPath(*s, cacheDBDir, appName)
	if err != nil {
		return err
	}

	fmt.Printf("deleting cache db at %s\n", s.CacheDBPath)
	if s.CacheDBPath != "" {
		err = os.Remove(s.CacheDBPath)
		if err != nil {
			return err
		}
	}

	_, err = Sync(cache.SyncInput{
		Session: s,
	}, true)

	return err
}
