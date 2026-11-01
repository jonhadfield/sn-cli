package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGetOptsCacheDBDirFlag(t *testing.T) {
	set := flag.NewFlagSet("sn", flag.ContinueOnError)
	require.NoError(t, (&cli.StringFlag{Name: "cachedb-dir"}).Apply(set))
	require.NoError(t, set.Parse([]string{"--cachedb-dir", "/tmp/sn-cache"}))

	require.Equal(t, "/tmp/sn-cache", getOpts(cli.NewContext(cli.NewApp(), set, nil)).cacheDBDir)
}
