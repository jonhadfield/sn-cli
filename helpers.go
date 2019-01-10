package sncli

import (
	"encoding/gob"
	"os"
	"strings"

	"github.com/jonhadfield/gosn"
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
		_ = encoder.Encode(object)
	}
	_ = file.Close()
	return err
}

func readGob(filePath string, object interface{}) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	_ = file.Close()
	return err
}

func ItemRefsToYaml(irs []gosn.ItemReference) []ItemReferenceYAML {
	iRefs := make([]ItemReferenceYAML, len(irs))
	for _, ref := range irs {
		iRef := ItemReferenceYAML{
			UUID:        ref.UUID,
			ContentType: ref.ContentType,
		}
		iRefs = append(iRefs, iRef)
	}
	return iRefs
}

func ItemRefsToJSON(irs []gosn.ItemReference) []ItemReferenceJSON {
	iRefs := make([]ItemReferenceJSON, len(irs))
	for _, ref := range irs {
		iRef := ItemReferenceJSON{
			UUID:        ref.UUID,
			ContentType: ref.ContentType,
		}
		iRefs = append(iRefs, iRef)
	}
	return iRefs
}

func CommaSplit(input string) []string {
	o := strings.Split(input, ",")
	if len(o) == 1 && len(o[0]) == 0 {
		return nil
	}
	return o
}
