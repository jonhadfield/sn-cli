package sncli

import (
	"fmt"
	"time"

	"github.com/jonhadfield/gosn"
)

type FixupConfig struct {
	Session gosn.Session
	Debug   bool
}

func (input *FixupConfig) Run() error {
	getItemsInput := gosn.GetItemsInput{
		Session: input.Session,
		Debug:   input.Debug,
	}

	var err error

	var output gosn.GetItemsOutput

	output, err = gosn.GetItems(getItemsInput)
	if err != nil {
		return err
	}

	output.Items.DeDupe()

	var pi gosn.Items
	pi, err = output.Items.DecryptAndParse(input.Session.Mk, input.Session.Ak, input.Debug)

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

					o := fmt.Sprintf("item: %s references missing item: %s\n", item.Content.GetTitle(), ref.UUID)
					fmt.Print(Yellow(o))
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

			eItemsWithRefsToUpdate, err = itemsWithRefsToUpdate.Encrypt(input.Session.Mk, input.Session.Ak, input.Debug)
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

			o := fmt.Sprintf("fixed references in %d items\n", len(itemsWithRefsToUpdate))
			fmt.Print(Green(o))
		}
	} else {
		fmt.Println(Green("no items with invalid references"))
	}

	// check for items without content type
	if len(missingContentType) > 0 {
		o := fmt.Sprintf("found %d notes with missing content type. delete? ", len(missingContentType))
		fmt.Print(Yellow(o))

		var response string

		_, err = fmt.Scanln(&response)
		if err == nil && StringInSlice(response, []string{"y", "yes"}, false) {
			var eMissingContentType gosn.EncryptedItems
			eMissingContentType, err = missingContentType.Encrypt(input.Session.Mk, input.Session.Ak, input.Debug)

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

			o := fmt.Sprintf("fixed %d items\n", len(missingContentType))
			fmt.Print(Yellow(o))
		}
	} else {
		fmt.Println(Green("no items with missing content type"))
	}

	// check for items with missing content
	if len(missingContent) > 0 {
		o := fmt.Sprintf("found %d notes with missing content. delete? ", len(missingContent))
		fmt.Print(Yellow(o))

		var response string
		_, err = fmt.Scanln(&response)

		if err != nil {
			return err
		}

		if StringInSlice(response, []string{"y", "yes"}, false) {
			var eMissingContent gosn.EncryptedItems

			eMissingContent, err = missingContent.Encrypt(input.Session.Mk, input.Session.Ak, input.Debug)
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
		fmt.Println(Green("no items with missing content"))
	}

	// check for items with missing titles
	if len(notesToTitleFix) > 0 {
		o := fmt.Sprintf("found %d items with missing titles. fix? ", len(notesToTitleFix))
		fmt.Print(Yellow(o))

		var response string

		_, err = fmt.Scanln(&response)
		if err != nil {
			return err
		}

		if StringInSlice(response, []string{"y", "yes"}, false) {
			var eNotesToTitleFix gosn.EncryptedItems

			eNotesToTitleFix, err = notesToTitleFix.Encrypt(input.Session.Mk, input.Session.Ak, input.Debug)
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
		fmt.Println(Green("no items with missing titles"))
	}

	return err
}
