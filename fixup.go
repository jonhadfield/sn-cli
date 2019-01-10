package sncli

import (
	"fmt"
	"log"
	"time"

	"github.com/jonhadfield/gosn"
)

type FixupConfig struct {
	Session gosn.Session
	Debug   bool
}

func (input *FixupConfig) Run() error {
	gosn.SetErrorLogger(log.Println)
	if input.Debug {
		gosn.SetDebugLogger(log.Println)
	}
	getItemsInput := gosn.GetItemsInput{
		Session: input.Session,
	}
	var err error
	var output gosn.GetItemsOutput
	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return err
	}
	output.Items.DeDupe()
	var pi gosn.Items
	pi, err = output.Items.DecryptAndParse(input.Session.Mk, input.Session.Ak)

	var missingContentType gosn.Items
	var missingContent gosn.Items
	var notesToTitleFix gosn.Items

	allIDs := make([]string, len(pi))
	var allItems gosn.Items

	for _, item := range pi {
		allIDs = append(allIDs, item.UUID)
		if !item.Deleted {
			allItems = append(allItems, item)
			switch {
			case item.ContentType == "":
				item.Deleted = true
				item.ContentType = "Note"
				missingContentType = append(missingContentType, item)
			case item.Content == nil && StringInSlice(item.ContentType, []string{"Note", "Tag"}, true):
				item.Deleted = true
				fmt.Println(item.ContentType)
				missingContent = append(missingContent, item)
			default:
				if item.ContentType == "Note" && item.Content.GetTitle() == "" {
					item.Content.SetUpdateTime(time.Now())
					item.Content.SetTitle("untitled")
					notesToTitleFix = append(notesToTitleFix, item)
				}
			}
		}
	}

	var itemsWithRefsToUpdate gosn.Items
	for _, item := range allItems {
		var newRefs []gosn.ItemReference
		var needsFix bool
		if item.Content != nil && item.Content.References() != nil && len(item.Content.References()) > 0 {
			for _, ref := range item.Content.References() {
				if !StringInSlice(ref.UUID, allIDs, false) {
					needsFix = true
					fmt.Printf("item: %s references missing item id: %s\n", item.Content.GetTitle(), ref.UUID)
				} else {
					newRefs = append(newRefs, ref)
				}
			}
			if needsFix {
				item.Content.SetReferences(newRefs)
				itemsWithRefsToUpdate = append(itemsWithRefsToUpdate, item)
			}
		}

	}

	// fix items with references to missing or deleted items
	if len(itemsWithRefsToUpdate) > 0 {
		fmt.Printf("found %d items with invalid references. fix? ", len(itemsWithRefsToUpdate))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eItemsWithRefsToUpdate gosn.EncryptedItems
			eItemsWithRefsToUpdate, err = itemsWithRefsToUpdate.Encrypt(input.Session.Mk, input.Session.Ak)
			if err != nil {
				return err
			}
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   eItemsWithRefsToUpdate,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed references in %d items\n", len(itemsWithRefsToUpdate))
		}
	} else {
		fmt.Println("no items with invalid references")
	}

	// check for items without content type
	if len(missingContentType) > 0 {
		fmt.Printf("found %d notes with missing content type. delete? ", len(missingContentType))
		var response string
		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eMissingContentType gosn.EncryptedItems
			eMissingContentType, err = missingContentType.Encrypt(input.Session.Mk, input.Session.Ak)
			if err != nil {
				return err
			}
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   eMissingContentType,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed %d items\n", len(missingContentType))

		}
	} else {
		fmt.Println("no items with missing content type")
	}

	// check for items with missing content
	if len(missingContent) > 0 {
		fmt.Printf("found %d notes with missing content. delete? ", len(missingContent))
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			return err
		}
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eMissingContent gosn.EncryptedItems
			eMissingContent, err = missingContent.Encrypt(input.Session.Mk, input.Session.Ak)
			if err != nil {
				return err
			}
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   eMissingContent,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed %d items\n", len(missingContent))

		}
	} else {
		fmt.Println("no items with missing content")
	}

	// check for items with missing titles
	if len(notesToTitleFix) > 0 {
		fmt.Printf("found %d items with missing titles. fix? ", len(notesToTitleFix))
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			return err
		}
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eNotesToTitleFix gosn.EncryptedItems
			eNotesToTitleFix, err = notesToTitleFix.Encrypt(input.Session.Mk, input.Session.Ak)
			if err != nil {
				return err
			}
			putItemsInput := gosn.PutItemsInput{
				Session: input.Session,
				Items:   eNotesToTitleFix,
			}
			_, err = gosn.PutItems(putItemsInput)
			if err != nil {
				return err
			}
			fmt.Printf("fixed %d items", len(notesToTitleFix))

		}
	} else {
		fmt.Println("no items with missing titles")
	}
	return err
}
