package sncli

import (
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
	if len(input) <= 0 {
		return "-"
	}
	return strings.Join(input, sep)
}
