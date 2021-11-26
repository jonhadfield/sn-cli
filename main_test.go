package sncli

import (
	"os"
	"runtime"
	"strings"
	"time"
)

func removeDB(dbPath string) {
	if err := os.Remove(dbPath); err != nil {
		if StringInSlice(runtime.GOOS, []string{"linux", "darwin"}, false) {
			if !strings.Contains(err.Error(), "no such file or directory") {
				panic(err)
			}
		}
		if runtime.GOOS == "windows" && !strings.Contains(err.Error(), "cannot find the file specified") {
			panic(err)
		}
	}
}

func testDelay() {
	time.Sleep(2 * time.Second)
}
