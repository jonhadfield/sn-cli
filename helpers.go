package sncli

import (
	"encoding/gob"
	"os"
	"strings"

	"github.com/jonhadfield/gosn-v2"
)

func StringInSlice(inStr string, inSlice []string, matchCase bool) bool {
	for i := range inSlice {
		if matchCase {
			if strings.EqualFold(inStr, inSlice[i]) {
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
	var iRefs []ItemReferenceYAML

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
	var iRefs []ItemReferenceJSON

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
	// split input
	o := strings.Split(input, ",")
	// strip leading and trailing space
	var s []string

	for _, i := range o {
		ti := strings.TrimSpace(i)
		if len(ti) > 0 {
			s = append(s, strings.TrimSpace(i))
		}
	}

	if len(s) == 1 && len(s[0]) == 0 {
		return nil
	}

	return s
}

func RemoveDeleted(in gosn.Items) (out gosn.Items) {
	for _, i := range in {
		if !i.IsDeleted() {
			out = append(out, i)
		}
	}

	return
}
