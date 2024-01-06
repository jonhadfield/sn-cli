package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddDeleteNote(t *testing.T) {
	time.Sleep(250 * time.Millisecond)

	var outputBuffer bytes.Buffer

	app := appSetup()

	app.Writer = &outputBuffer

	osArgs := []string{"sncli", "add", "note", "--title", "testNote", "--text", "testAddNote"}

	err := app.Run(osArgs)
	require.NoError(t, err)

	stdout := outputBuffer.String()

	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgAddSuccess)

	outputBuffer.Reset()
	osArgs = []string{"sncli", "delete", "note", "--title", "testNote"}
	err = app.Run(osArgs)
	stdout = outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgDeleted)
}

func TestGetMissingNote(t *testing.T) {
	time.Sleep(250 * time.Millisecond)
	var outputBuffer bytes.Buffer
	app := appSetup()

	app.Writer = &outputBuffer
	osArgs := []string{"sncli", "get", "note", "--title", "missing note"}

	err := app.Run(osArgs)
	require.NoError(t, err)

	stdout := outputBuffer.String()
	fmt.Println(stdout)
	require.NoError(t, err)
	require.Contains(t, stdout, msgNoMatches)
}

func TestDeleteNonExistantNote(t *testing.T) {
	time.Sleep(250 * time.Millisecond)
	var outputBuffer bytes.Buffer
	app := appSetup()
	app.Writer = &outputBuffer

	outputBuffer.Reset()
	require.NoError(t, app.Run([]string{"sncli", "delete", "note", "--title", "testNote"}))

	stdout := outputBuffer.String()
	fmt.Println(stdout)
	require.Contains(t, stdout, msgNoteNotFound)
}
