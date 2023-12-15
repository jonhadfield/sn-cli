package sncli

import (
	"encoding/base64"
	"fmt"
	"github.com/gookit/color"
	"strings"

	"github.com/jonhadfield/gosn-v2/crypto"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
)

type DecryptStringInput struct {
	Session   session.Session
	In        string
	UseStdOut bool
	Key       string
}

func DecryptString(input DecryptStringInput) (plaintext string, err error) {
	key1 := input.Session.MasterKey
	if input.Key != "" {
		key1 = input.Key
	}

	// trim noise
	if strings.HasPrefix(input.In, "enc_item_key") && len(input.In) > 13 {
		input.In = strings.TrimSpace(input.In)[13:]
	}
	if strings.HasPrefix(input.In, "content") && len(input.In) > 8 {
		input.In = strings.TrimSpace(input.In)[8:]
	}

	version, nonce, cipherText, authData := splitContent(input.In)
	if version != "004" {
		return plaintext, fmt.Errorf("only version 004 of encryption is supported")
	}

	bad, err := base64.StdEncoding.DecodeString(authData)
	if err != nil {
		err = fmt.Errorf("failed to base64 decode auth data: '%s' err: %+v", authData, err)

		return
	}
	fmt.Printf("Decoded Auth Data: %+v\n", string(bad))

	pb, err := crypto.DecryptCipherText(cipherText, key1, nonce, authData)
	if err != nil {
		return
	}

	return string(pb), nil
}

type OutputSessionInput struct {
	Session         session.Session
	In              string
	UseStdOut       bool
	OutputMasterKey bool
}

func OutputSession(input OutputSessionInput) error {
	fmt.Println(color.Bold.Sprintf("session"))
	fmt.Printf("debug: %t\n\n", input.Session.Debug)
	fmt.Println("key params")
	fmt.Printf("- identifier: %s\n", input.Session.KeyParams.Identifier)
	fmt.Printf("- nonce: %s\n", input.Session.KeyParams.PwNonce)
	fmt.Printf("- created: %s\n", input.Session.KeyParams.Created)
	fmt.Printf("- origination: %s\n", input.Session.KeyParams.Origination)
	fmt.Printf("- version: %s\n", input.Session.KeyParams.Version)
	fmt.Println()
	if input.OutputMasterKey {
		fmt.Printf("master key: %s\n", input.Session.MasterKey)
		fmt.Println()
	}

	_, err := items.Sync(items.SyncInput{Session: &input.Session})
	if err != nil {
		return err
	}

	// output default items key
	ik := input.Session.DefaultItemsKey
	fmt.Println("default items key")
	fmt.Printf("- uuid %s key %s created-at %d updated-at %d\n", ik.UUID, ik.ItemsKey, ik.CreatedAtTimestamp, ik.UpdatedAtTimestamp)

	// output all items keys in session
	fmt.Println("items keys")

	for _, ik = range input.Session.ItemsKeys {
		fmt.Printf("- uuid %s key %s created-at %d updated-at %d\n", ik.UUID, ik.ItemsKey, ik.CreatedAtTimestamp, ik.UpdatedAtTimestamp)
	}

	return nil
}

type CreateItemsKeyInput struct {
	Debug     bool
	MasterKey string
}

// func CreateItemsKey(input CreateItemsKeyInput) error {
// 	ik := items.NewItemsKey()
// 	fmt.Printf("%+v\n", ik.ItemsKey)
//
// 	return nil
// }

func splitContent(in string) (version, nonce, cipherText, authenticatedData string) {
	components := strings.Split(in, ":")
	if len(components) < 3 {
		panic(components)
	}

	version = components[0]           // protocol version
	nonce = components[1]             // encryption nonce
	cipherText = components[2]        // ciphertext
	authenticatedData = components[3] // authenticated data

	return
}
