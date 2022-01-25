package sncli

import (
	"fmt"
	"github.com/jonhadfield/gosn-v2"
)

func ItemKeysHealthcheck(s *gosn.Session, fix bool) error {
	fmt.Printf("apply fix: %+v\n", fix)

	so, err := gosn.Sync(gosn.SyncInput{
		Session: s,
	})
	if err != nil {
		return err
	}

	seenItemsKeys := make(map[string]int)
	for x := range so.Items {
		switch so.Items[x].ContentType {
		case "SN|ItemsKey":
			seenItemsKeys[so.Items[x].UUID]++
			if so.Items[x].UUID == "" {
				fmt.Printf("items key without UUID: %+v\n", so.Items[x])
			}
			if so.Items[x].EncItemKey == "" {
				fmt.Printf("items key without enc_item_key: %+v\n", so.Items[x])
			}
		}
	}

	for k, v := range seenItemsKeys {
		fmt.Printf("ItemsKey %s has %d instances\n", k, v)
	}

	var itemsWithMissingKeys []string
	var encitemsNotSpecifyingItemsKeyID []string
	var itemsKeysInUse []string
	// check for unused ItemsKeys
	for x := range so.Items {
		if so.Items[x].ContentType != "SN|ItemsKey" && !isUnsupportedType(so.Items[x].ContentType) {
			if so.Items[x].ItemsKeyID == nil {
				fmt.Printf("%s %s has no ItemsKeyID\n", so.Items[x].ContentType, so.Items[x].UUID)
				encitemsNotSpecifyingItemsKeyID = append(encitemsNotSpecifyingItemsKeyID, so.Items[x].UUID)

				continue
			}

			itemsKeysInUse = append(itemsKeysInUse, *so.Items[x].ItemsKeyID)

			if !itemsKeyExists(*so.Items[x].ItemsKeyID, seenItemsKeys) {
				fmt.Printf("No matching items key found for %s %s specifying key: %s\n", so.Items[x].ContentType, so.Items[x].UUID, *so.Items[x].ItemsKeyID)
				itemsWithMissingKeys = append(itemsWithMissingKeys, fmt.Sprintf("Type: %s UUID: %s| ItemsKeyID: %s", so.Items[x].ContentType, so.Items[x].UUID, *so.Items[x].ItemsKeyID))
			}
		}
	}

	if len(encitemsNotSpecifyingItemsKeyID) > 0 {
		fmt.Println("No matching ItemsKey specified for these encrypted items:")
		for x := range encitemsNotSpecifyingItemsKeyID {
			fmt.Printf("%s\n", encitemsNotSpecifyingItemsKeyID[x])
		}
	}

	if len(itemsWithMissingKeys) > 0 {
		fmt.Println("No matching ItemsKey was found these items:")
		for x := range itemsWithMissingKeys {
			fmt.Printf("%s\n", itemsWithMissingKeys[x])
		}
	}

	for k := range seenItemsKeys {
		var seen bool
		for x := range itemsKeysInUse {
			if k == itemsKeysInUse[x] {
				seen = true

				break
			}
		}
		if !seen {
			fmt.Printf("ItemsKey with UUID: %s is unused\n", k)
		}
	}

	return nil
}

func itemsKeyExists(uuid string, seenItemsKeys map[string]int) bool {
	if seenItemsKeys[uuid] > 0 {
		return true
	}

	return false
}

func isEncryptedWithMasterKey(t string) bool {
	return t == "SN|ItemsKey"
}

func isUnsupportedType(t string) bool {
	return false
	//return strings.HasPrefix(t, "SF|")
}
