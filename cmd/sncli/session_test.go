package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonhadfield/gosn-v2/session"
	"github.com/stretchr/testify/require"
)

func TestUseSessionFile(t *testing.T) {
	t.Cleanup(func() { session.SetDefaultKeyring(nil) })

	path := filepath.Join(t.TempDir(), "session")
	require.NoError(t, os.WriteFile(path, []byte(`{"access_token":"token"}`), 0o600))

	require.NoError(t, useSessionFile(path))
	require.NoError(t, session.SessionExists(nil))

	require.Equal(t, session.MsgSessionRemovalSuccess, session.RemoveSession(nil))
	require.NoFileExists(t, path)
}

func TestUseSessionFileExpandsHome(t *testing.T) {
	t.Cleanup(func() { session.SetDefaultKeyring(nil) })

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // os.UserHomeDir on Windows

	require.NoError(t, os.WriteFile(filepath.Join(home, "sn-session"), []byte(`{"access_token":"token"}`), 0o600))

	require.NoError(t, useSessionFile("~/sn-session"))
	require.NoError(t, session.SessionExists(nil))
}

func TestUseSessionFileMissingFile(t *testing.T) {
	t.Cleanup(func() { session.SetDefaultKeyring(nil) })

	require.NoError(t, useSessionFile(filepath.Join(t.TempDir(), "missing")))
	require.Error(t, session.SessionExists(nil))
}
