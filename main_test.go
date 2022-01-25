package sncli

import (
	"os"
	"strings"
	"time"
)

// prevent throttling when using official server.
func testDelay() {
	if strings.Contains(os.Getenv("SN_SERVER"), "api.standardnotes.com") {
		time.Sleep(2 * time.Second)
	}
}
