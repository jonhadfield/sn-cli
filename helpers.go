package sncli

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func outList(i []string, sep string) string {
	if len(i) == 0 {
		return "-"
	}

	return strings.Join(i, sep)
}

func writeGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		_ = encoder.Encode(object)
	}

	if file != nil {
		_ = file.Close()
	}

	return err
}

func writeJSON(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	var jsonExport []byte
	if err == nil {
		jsonExport, err = json.MarshalIndent(object,"","  ")
	}

	content := strings.Builder{}
	content.WriteString("{\n  \"items\": ")
	content.WriteString(string(jsonExport))
	content.WriteString("\n}")

	_, err = file.WriteString(content.String())

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

func readJSON(filePath string, items *gosn.EncryptedItems) error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open: %s", filePath)
	}
	err = json.Unmarshal(file, &items)
	if err != nil {
		return fmt.Errorf("failed to unmarshall json: %w", err)
	}

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

func CommaSplit(i string) []string {
	// split i
	o := strings.Split(i, ",")
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
