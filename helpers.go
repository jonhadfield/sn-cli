package sncli

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jonhadfield/gosn-v2/items"
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

type EncryptedItemExport struct {
	UUID        string `json:"uuid"`
	ItemsKeyID  string `json:"items_key_id,omitempty"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	// Deleted            bool    `json:"deleted"`
	EncItemKey         string  `json:"enc_item_key"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAtTimestamp int64   `json:"created_at_timestamp"`
	UpdatedAtTimestamp int64   `json:"updated_at_timestamp"`
	DuplicateOf        *string `json:"duplicate_of"`
}

func writeJSON(i ExportConfig, items items.EncryptedItems) error {
	// prepare for export
	var itemsExport []EncryptedItemExport
	for x := range items {
		itemsExport = append(itemsExport, EncryptedItemExport{
			UUID:       items[x].UUID,
			ItemsKeyID: items[x].ItemsKeyID,
			Content:    items[x].Content,
			// Deleted:            items[x].Deleted,
			ContentType:        items[x].ContentType,
			EncItemKey:         items[x].EncItemKey,
			CreatedAt:          items[x].CreatedAt,
			UpdatedAt:          items[x].UpdatedAt,
			CreatedAtTimestamp: items[x].CreatedAtTimestamp,
			UpdatedAtTimestamp: items[x].UpdatedAtTimestamp,
			DuplicateOf:        items[x].DuplicateOf,
		})
	}

	file, err := os.Create(i.File)
	if err != nil {
		return err
	}

	defer file.Close()

	var jsonExport []byte
	if err == nil {
		jsonExport, err = json.MarshalIndent(itemsExport, "", "  ")
	}

	content := strings.Builder{}
	content.WriteString("{\n  \"version\": \"004\",")
	content.WriteString("\n  \"items\": ")
	content.Write(jsonExport)
	content.WriteString(",")

	// add keyParams
	content.WriteString("\n  \"keyParams\": {")
	content.WriteString(fmt.Sprintf("\n    \"identifier\": \"%s\",", i.Session.KeyParams.Identifier))
	content.WriteString(fmt.Sprintf("\n    \"version\": \"%s\",", i.Session.KeyParams.Version))
	content.WriteString(fmt.Sprintf("\n    \"origination\": \"%s\",", i.Session.KeyParams.Origination))
	content.WriteString(fmt.Sprintf("\n    \"created\": \"%s\",", i.Session.KeyParams.Created))
	content.WriteString(fmt.Sprintf("\n    \"pw_nonce\": \"%s\"", i.Session.KeyParams.PwNonce))
	content.WriteString("\n  }")

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

type EncryptedItemsFile struct {
	Items items.EncryptedItems `json:"items"`
}

func readJSON(filePath string) (items items.EncryptedItems, err error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		err = fmt.Errorf("%w failed to open: %s", err, filePath)
		return
	}

	var eif EncryptedItemsFile

	err = json.Unmarshal(file, &eif)
	if err != nil {
		err = fmt.Errorf("failed to unmarshall json: %w", err)
		return
	}

	return eif.Items, err
}

func ItemRefsToYaml(irs []items.ItemReference) []ItemReferenceYAML {
	var iRefs []ItemReferenceYAML

	for _, ref := range irs {
		iRef := ItemReferenceYAML{
			UUID:          ref.UUID,
			ContentType:   ref.ContentType,
			ReferenceType: ref.ReferenceType,
		}
		iRefs = append(iRefs, iRef)
	}

	return iRefs
}

func ItemRefsToJSON(irs []items.ItemReference) []ItemReferenceJSON {
	var iRefs []ItemReferenceJSON

	for _, ref := range irs {
		iRef := ItemReferenceJSON{
			UUID:          ref.UUID,
			ContentType:   ref.ContentType,
			ReferenceType: ref.ReferenceType,
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

func RemoveDeleted(in items.Items) (out items.Items) {
	for _, i := range in {
		if !i.IsDeleted() {
			out = append(out, i)
		}
	}

	return
}
