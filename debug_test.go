package sncli

import (
	"github.com/jonhadfield/gosn-v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDebugStringItemsKeyEncItemKey(t *testing.T) {
	plaintext, err := DecryptString(DecryptStringInput{
		Session:   gosn.Session{},
		In:        "004:966fb3ab4422f3b2a0ea20159d769cbee1d7f763d5aba425:xRmX5+LlNYk4bXI4OaeKNMI2LZE3QeLO5xFVwy+v3PY8hBUBOuDRa+n7m01gfA57fPL4JBrWhGJ9b8gaiOPFGC8ntlBL7+qhj/sQ/cbA4Po=:eyJrcCI6eyJjcmVhdGVkIjoiMTYwODQ3MzM4Nzc5OSIsImlkZW50aWZpZXIiOiJsZW1vbjIiLCJvcmlnaW5hdGlvbiI6InJlZ2lzdHJhdGlvbiIsInB3X25vbmNlIjoiNGZXc2RUOHJaY2NOTlR1aWVZcTF2bGJ6YkIzRmVkeE0iLCJ2ZXJzaW9uIjoiMDA0In0sInUiOiJhMWIxZDYwYy1mNjA1LTQ2MDQtOGE5ZS03NjE1NDkyODI4M2IiLCJ2IjoiMDA0In0=",
		UseStdOut: false,
		Key:       "9dbd97421d3981c433979fc8d86559734331f711372c4ad7a0a6830fff75af68",
	})
	require.NoError(t, err)
	require.Equal(t, "449adb29a39e770048dc6126565d2fe5c3a9b4094f19c1e109b84b17d0cf27bb", plaintext)
}

func TestDebugStringItemsKeyEncItemKeySession(t *testing.T) {
	plaintext, err := DecryptString(DecryptStringInput{
		Session: gosn.Session{
			MasterKey: "9dbd97421d3981c433979fc8d86559734331f711372c4ad7a0a6830fff75af68",
		},
		In:        "004:966fb3ab4422f3b2a0ea20159d769cbee1d7f763d5aba425:xRmX5+LlNYk4bXI4OaeKNMI2LZE3QeLO5xFVwy+v3PY8hBUBOuDRa+n7m01gfA57fPL4JBrWhGJ9b8gaiOPFGC8ntlBL7+qhj/sQ/cbA4Po=:eyJrcCI6eyJjcmVhdGVkIjoiMTYwODQ3MzM4Nzc5OSIsImlkZW50aWZpZXIiOiJsZW1vbjIiLCJvcmlnaW5hdGlvbiI6InJlZ2lzdHJhdGlvbiIsInB3X25vbmNlIjoiNGZXc2RUOHJaY2NOTlR1aWVZcTF2bGJ6YkIzRmVkeE0iLCJ2ZXJzaW9uIjoiMDA0In0sInUiOiJhMWIxZDYwYy1mNjA1LTQ2MDQtOGE5ZS03NjE1NDkyODI4M2IiLCJ2IjoiMDA0In0=",
		UseStdOut: false,
	})
	require.NoError(t, err)
	require.Equal(t, "449adb29a39e770048dc6126565d2fe5c3a9b4094f19c1e109b84b17d0cf27bb", plaintext)
}

func TestDebugStringItemsKeyContent(t *testing.T) {
	plaintext, err := DecryptString(DecryptStringInput{
		In:        "004:0f62ec0954de2aaf4f6bf529e6478b6725774cd3ce396d94:+keNC3SOAPp890NTrHTKnDn8tK8QgCNVA51U1L3XWfOK4lr65Ju4qtciY57NTDrXKok80CeyzY6lPwtW8dIExgHDKf+yjlPYHqxLwWOXytDvZA9o/8kQ0ciYG9XLdN9YCuUw3evV7jXkB5cVa6kUwqLhbQnerCXrXOaiFPkUoxaAxP7GP8ciYdwegRkag67DiZEbD5d5/iPGY2zN4u4ltapgWMU7BgbTMpvJUaMzYyrolmw6eY9KVS3x02IKHhaYbtd6Co5/YG1BGEY85F3vfuSkusR3Pwci2pk1nOcTKyukUoRCksyubr9G963HG/BKmAxmq02txd+D6ppZjoIfIoxeG+JNSvXcMp0iXMk3wrzHIKaA6/6v8n4fZAv61p1JEni3MAKwtMk6/XetdA==:eyJrcCI6eyJjcmVhdGVkIjoiMTYwODQ3MzM4Nzc5OSIsImlkZW50aWZpZXIiOiJsZW1vbjIiLCJvcmlnaW5hdGlvbiI6InJlZ2lzdHJhdGlvbiIsInB3X25vbmNlIjoiNGZXc2RUOHJaY2NOTlR1aWVZcTF2bGJ6YkIzRmVkeE0iLCJ2ZXJzaW9uIjoiMDA0In0sInUiOiJhMWIxZDYwYy1mNjA1LTQ2MDQtOGE5ZS03NjE1NDkyODI4M2IiLCJ2IjoiMDA0In0=",
		UseStdOut: false,
		Key:       "449adb29a39e770048dc6126565d2fe5c3a9b4094f19c1e109b84b17d0cf27bb",
	})
	require.NoError(t, err)
	require.Equal(t, "{\"itemsKey\":\"b710239b882127f663f2be3e2811a4add4e465b23458c8a770facc1aaed5b526\",\"version\":\"004\",\"references\":[],\"appData\":{\"org.standardnotes.sn\":{\"client_updated_at\":\"Thu Feb 10 2022 19:58:58 GMT+0000 (Greenwich Mean Time)\",\"prefersPlainEditor\":false,\"pinned\":false}},\"isDefault\":true}", plaintext)
}
