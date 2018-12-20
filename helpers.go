package sncli

import (
	"encoding/gob"
	"os"
	"strings"
)

func StringInSlice(inStr string, inSlice []string, matchCase bool) bool {
	for i := range inSlice {
		if matchCase {
			if strings.ToLower(inStr) == strings.ToLower(inSlice[i]) {
				return true
			}
		} else {
			if inStr == inSlice[i] {
				return true
			}
		}

	}
	return false
}

func outList(input []string, sep string) string {
	if len(input) == 0 {
		return "-"
	}
	return strings.Join(input, sep)
}

func writeGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func readGob(filePath string, object interface{}) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}
