package sncli

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/jonhadfield/gosn-v2"
	"os"
	"time"
)

type ItemsKeysHealthcheckInput struct {
	Session       gosn.Session
	UseStdOut     bool
	DeleteInvalid bool
}

func ItemKeysHealthcheck(input ItemsKeysHealthcheckInput) error {
	var so gosn.SyncOutput
	var err error
	var syncToken string

	// request all items from SN

	if !input.Session.Debug {
		prefix := HiWhite("retrieving items ")

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stdout))
		if input.UseStdOut {
			s = spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
		}

		s.Prefix = prefix
		s.Start()

		so, err = gosn.Sync(gosn.SyncInput{
			Session: &input.Session,
		})

		s.Stop()
	} else {
		so, err = gosn.Sync(gosn.SyncInput{
			Session: &input.Session,
		})
	}

	if err != nil {
		return err
	}

	syncToken = so.SyncToken

	// get a list of items keys and a count of items each one is used to encrypt

	var seenItemsKeys []string
	referencedItemsKeys := make(map[string]int)
	for x := range so.Items {
		if so.Items[x].Deleted {
			continue
		}
		switch so.Items[x].ContentType {
		case "SN|ItemsKey":
			// add to list of seen keys
			seenItemsKeys = append(seenItemsKeys, so.Items[x].UUID)

			if so.Items[x].UUID == "" {
				fmt.Printf("items key without UUID: %+v\n", so.Items[x])
			}
			if so.Items[x].EncItemKey == "" {
				fmt.Printf("items key without enc_item_key: %+v\n", so.Items[x])
			}

		default:
			if so.Items[x].ItemsKeyID != nil {
				// if an item has an items key id specified, then increment the count of how many
				// items the items key references
				referencedItemsKeys[*so.Items[x].ItemsKeyID]++
			}
		}
	}

	fmt.Println("existing Items Keys:")
	for x := range seenItemsKeys {
		fmt.Printf("- %s\n", seenItemsKeys[x])
	}

	fmt.Println()

	fmt.Println("item references per Items Key:")
	for k, v := range referencedItemsKeys {
		if v != 0 {
			fmt.Printf("- %s | %d\n", k, v)
		}
	}

	fmt.Println()

	var itemsWithMissingKeys []string
	var encitemsNotSpecifyingItemsKeyID gosn.EncryptedItems
	var itemsKeysNotEncryptedWithCurrentMasterKey []string
	var itemsKeysInUse []string
	// check for unused ItemsKeys
	for x := range so.Items {
		// skip deleted items
		if so.Items[x].Deleted {
			continue
		}

		switch {
		case so.Items[x].ContentType == "SN|ItemsKey":
			var ik gosn.ItemsKey
			ik, err = so.Items[x].Decrypt(input.Session.MasterKey)
			if err != nil || ik.ItemsKey == "" {
				itemsKeysNotEncryptedWithCurrentMasterKey = append(itemsKeysNotEncryptedWithCurrentMasterKey,
					so.Items[x].UUID)
			}

		case !isEncryptedWithMasterKey(so.Items[x].ContentType):
			if so.Items[x].ItemsKeyID == nil {
				fmt.Printf("%s %s has no ItemsKeyID\n", so.Items[x].ContentType, so.Items[x].UUID)
				encitemsNotSpecifyingItemsKeyID = append(encitemsNotSpecifyingItemsKeyID, so.Items[x])

				continue

			}
			itemsKeysInUse = append(itemsKeysInUse, *so.Items[x].ItemsKeyID)
			if !itemsKeyExists(*so.Items[x].ItemsKeyID, seenItemsKeys) {
				fmt.Printf("no matching items key found for %s %s specifying key: %s\n",
					so.Items[x].ContentType,
					so.Items[x].UUID,
					*so.Items[x].ItemsKeyID)
				itemsWithMissingKeys = append(itemsWithMissingKeys,
					fmt.Sprintf("Type: %s UUID: %s| ItemsKeyID: %s",
						so.Items[x].ContentType,
						so.Items[x].UUID,
						*so.Items[x].ItemsKeyID))
			}
		}
	}

	fmt.Println("unused Items Keys:")
	var numUnusedItemsKeys int64
	for x := range seenItemsKeys {
		var seen bool
		for y := range itemsKeysInUse {
			if seenItemsKeys[x] == itemsKeysInUse[y] {
				seen = true

				break
			}
		}
		if !seen {
			numUnusedItemsKeys++
			fmt.Printf("- %s\n", seenItemsKeys[x])
		}
	}
	if numUnusedItemsKeys == 0 {
		fmt.Println("none")
	}

	if len(encitemsNotSpecifyingItemsKeyID) > 0 {
		fmt.Println("no matching ItemsKey specified for these encrypted items:")
		for x := range encitemsNotSpecifyingItemsKeyID {
			fmt.Printf("- %s %s\n", encitemsNotSpecifyingItemsKeyID[x].ContentType, encitemsNotSpecifyingItemsKeyID[x].UUID)
		}
	}

	if len(itemsWithMissingKeys) > 0 && input.DeleteInvalid {
		fmt.Printf("wipe all encrypted items without items keys (account: %s)? ",
			input.Session.KeyParams.Identifier)

		var c string
		_, err = fmt.Scanln(&c)
		if err == nil && StringInSlice(c, []string{"y", "yes"}, false) {
			var itemsToDelete gosn.EncryptedItems
			for x := range encitemsNotSpecifyingItemsKeyID {
				itemToDelete := encitemsNotSpecifyingItemsKeyID[x]
				itemToDelete.Deleted = true
				itemsToDelete = append(itemsToDelete, itemToDelete)
			}

			so, err = gosn.Sync(gosn.SyncInput{
				SyncToken: syncToken,
				Session:   &input.Session,
				Items:     itemsToDelete,
			})

			if err != nil {
				return err
			}

			fmt.Printf("successfully deleted %d items\n", len(so.SavedItems))
		}
	}

	return nil
}

func itemsKeyExists(uuid string, seenItemsKeys []string) bool {
	for x := range seenItemsKeys {
		if seenItemsKeys[x] == uuid {
			return true
		}
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
