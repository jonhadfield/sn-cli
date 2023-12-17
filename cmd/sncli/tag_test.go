package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddTag(t *testing.T) {
	time.Sleep(1 * time.Second)
	var outputBuffer bytes.Buffer
	app, err := appSetup()
	require.NoError(t, err)
	app.Writer = &outputBuffer

	osArgs := []string{"sncli", "add", "tag", "--title", "testAddOneTagGetCount"}
	err = app.Run(osArgs)
	stdout := outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgAddSuccess)

	outputBuffer.Reset()

	osArgs = []string{"sncli", "add", "tag", "--title", "testAddOneTagGetCount"}
	err = app.Run(osArgs)
	stdout = outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgTagAlreadyExists)

	// err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})

	// cmd := cmdAdd()
	// cmd.
	// require.NoError(t, err)
	// err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testAddOneTagGetCount"})
	// require.NoError(t, err)
	// require.Contains(t, msg, msgAddSuccess)
}

func TestAddGetTag(t *testing.T) {
	time.Sleep(1 * time.Second)
	var outputBuffer bytes.Buffer
	app, _ := appSetup()
	app.Writer = &outputBuffer

	// wipe
	osArgs := []string{"sncli", "wipe", "--yes"}
	err := app.Run(osArgs)
	stdout := outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)

	// add tag
	osArgs = []string{"sncli", "add", "tag", "--title", "testAddOneTagGetCount"}
	err = app.Run(osArgs)
	stdout = outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgAddSuccess)

	// get tag
	osArgs = []string{"sncli", "get", "tag", "--title", "testAddOneTagGetCount"}
	err = app.Run(osArgs)
	stdout = outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgAddSuccess)
}

// func TestAddGetTagExport(t *testing.T) {
// 	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
// 	require.NoError(t, err)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testAddOneTagGetCount"})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgAddSuccess)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "get", "tag", "--title", "testAddOneTagGetCount"})
// 	require.NoError(t, err)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "export"})
// 	require.NoError(t, err)
// }

func TestAddDeleteTag(t *testing.T) {
	time.Sleep(1 * time.Second)
	var outputBuffer bytes.Buffer
	app, _ := appSetup()
	app.Writer = &outputBuffer

	// wipe
	osArgs := []string{"sncli", "wipe", "--yes"}
	err := app.Run(osArgs)
	stdout := outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)

	// add tag
	osArgs = []string{"sncli", "add", "tag", "--title", "testTag"}
	err = app.Run(osArgs)
	stdout = outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgAddSuccess)

	// get tag
	osArgs = []string{"sncli", "get", "tag", "--title", "testTag"}
	err = app.Run(osArgs)
	stdout = outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgAddSuccess)

	// delete tag
	osArgs = []string{"sncli", "delete", "tag", "--title", "testTag"}
	err = app.Run(osArgs)
	stdout = outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgTagDeleted)
}

// func TestAddTagExportDeleteTagReImport(t *testing.T) {
// 	msg, _, err := startCLI([]string{"sncli", "--debug", "--no-stdout", "wipe", "--yes"})
// 	require.NoError(t, err)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "add", "tag", "--title", "testAddOneTagGetCount"})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgAddSuccess)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "1", msg)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "export"})
// 	require.NoError(t, err)
// 	require.True(t, strings.HasPrefix(msg, "encrypted export written to:"))
// 	path := strings.TrimPrefix(msg, "encrypted export written to:")
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "--no-stdout", "delete", "tag", "--title", "testAddOneTagGetCount"})
// 	require.NoError(t, err)
// 	require.Contains(t, msg, msgDeleted)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "0", msg)
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "import", "--experiment", "--file", path})
// 	require.NoError(t, err)
// 	require.True(t, strings.HasPrefix(msg, "imported"))
// 	msg, _, err = startCLI([]string{"sncli", "--debug", "get", "tag", "--count"})
// 	require.NoError(t, err)
// 	require.Equal(t, "1", msg)
// }

func TestAddTagErrorMissingTitle(t *testing.T) {
	err := startCLI([]string{"sncli", "add", "tag"})
	require.Error(t, err, "error should be returned if title is unspecified")
}

func TestDeleteTagErrorMissingTitle(t *testing.T) {
	err := startCLI([]string{"sncli", "delete", "tag"})
	require.Error(t, err, "error should be returned if title is unspecified")
}
