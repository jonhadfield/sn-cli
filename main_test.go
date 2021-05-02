package sncli

import (
	"os"
	"strings"
	"time"
)

func removeDB(dbPath string) {
	if err := os.Remove(dbPath); err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			panic(err)
		}
	}
}

func testDelay() {
	time.Sleep(2 * time.Second)
}
