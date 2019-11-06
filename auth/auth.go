package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"syscall"

	"github.com/jonhadfield/gosn"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/viper"
)

const (
	SNServerURL              = "https://sync.standardnotes.org"
	KeyringApplicationName   = "session"
	KeyringService           = "StandardNotesCLI"
	MsgSessionRemovalSuccess = "session removed successfully"
	MsgSessionRemovalFailure = "failed to remove session"
)

func GetCredentials(inServer string) (email, password, apiServer, errMsg string) {
	switch {
	case viper.GetString("email") != "":
		email = viper.GetString("email")
	default:
		fmt.Print("email: ")

		_, err := fmt.Scanln(&email)
		if err != nil || len(strings.TrimSpace(email)) == 0 {
			errMsg = "email required"
			return
		}
	}

	if viper.GetString("password") != "" {
		password = viper.GetString("password")
	} else {
		fmt.Print("password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err == nil {
			password = string(bytePassword)
		} else {
			errMsg = err.Error()
			return
		}
		if strings.TrimSpace(password) == "" {
			errMsg = "password not defined"
		}
	}

	switch {
	case inServer != "":
		apiServer = inServer
	case viper.GetString("server") != "":
		apiServer = viper.GetString("server")
	default:
		apiServer = SNServerURL
	}

	return email, password, apiServer, errMsg
}

// encrypt string to base64 crypto using AES
func Encrypt(key []byte, text string) string {
	key = padToAESBlockSize(key)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

func GetSessionFromKeyring(k keyring.Keyring) (s string, err error) {
	if k == nil {
		return keyring.Get(KeyringService, KeyringApplicationName)
	}

	return k.Get(KeyringService, KeyringApplicationName)
}

func AddSession(snServer, inKey string, k keyring.Keyring) (res string, err error) {
	// check if session exists in keyring
	var s string
	s, err = GetSessionFromKeyring(k)
	// only return an error if there's an issue accessing the keyring
	if err != nil && !strings.Contains(err.Error(), "secret not found in keyring") {
		return
	}

	if inKey == "." {
		var byteKey []byte

		fmt.Print("session key: ")

		byteKey, err = terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return
		}

		inKey = string(byteKey)

		fmt.Println()
	}

	if s != "" {
		fmt.Print("replace existing session (y|n): ")

		var resp string

		_, err := fmt.Scanln(&resp)
		if err != nil || strings.ToLower(resp) != "y" {
			// do nothing
			return "", nil
		}
	}

	var session gosn.Session

	var email string

	session, email, err = GetSessionFromUser(snServer)
	if err != nil {
		return fmt.Sprint("failed to get session: ", err), err
	}

	rS := makeSessionString(email, session)

	if inKey != "" {
		key := []byte(inKey)
		rS = Encrypt(key, MakeSessionString(email, session))
	}

	err = writeSession(rS, k)
	if err != nil {
		return fmt.Sprint("failed to set session: ", err), err
	}

	return "session added successfully", err
}

func writeSession(s string, k keyring.Keyring) error {
	if k == nil {
		return keyring.Set(KeyringService, KeyringApplicationName, s)
	}

	return k.Set(KeyringService, KeyringApplicationName, s)
}

func makeSessionString(email string, session gosn.Session) string {
	return fmt.Sprintf("%s;%s;%s;%s;%s", email, session.Server, session.Token, session.Ak, session.Mk)
}

func SessionExists(k keyring.Keyring) error {
	s, err := GetSessionFromKeyring(k)
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return errors.New("session is empty")
	}

	return nil
}

// RemoveSession removes the SN session from the keyring
func RemoveSession(k keyring.Keyring) string {
	var err error

	if err = SessionExists(k); err != nil {
		return fmt.Sprintf("%s: %s", MsgSessionRemovalFailure, err.Error())
	}

	if k == nil {
		err = keyring.Delete(KeyringService, KeyringApplicationName)
	} else {
		err = k.Delete(KeyringService, KeyringApplicationName)
	}

	if err != nil {
		return fmt.Sprintf("%s: %s", MsgSessionRemovalFailure, err.Error())
	}

	return MsgSessionRemovalSuccess
}

func MakeSessionString(email string, session gosn.Session) string {
	return fmt.Sprintf("%s;%s;%s;%s;%s", email, session.Server, session.Token, session.Ak, session.Mk)
}

func GetSessionFromUser(server string) (gosn.Session, string, error) {
	var sess gosn.Session

	var err error

	var email, password, apiServer, errMsg string

	email, password, apiServer, errMsg = GetCredentials(server)
	if errMsg != "" {
		if strings.Contains(errMsg, "password not defined") {
			err = fmt.Errorf("password not defined")
		} else {
			fmt.Printf("\nerror: %s\n\n", errMsg)
		}

		return sess, email, err
	}

	sess, err = gosn.CliSignIn(email, password, apiServer)
	if err != nil {
		return sess, email, err
	}

	return sess, email, err
}

func GetSession(loadSession bool, sessionKey, server string) (session gosn.Session, email string, err error) {
	if loadSession {
		var rawSess string

		rawSess, err = keyring.Get(KeyringService, KeyringApplicationName)
		if err != nil {
			return
		}

		if !isUnencryptedSession(rawSess) {
			if sessionKey == "" {
				var byteKey []byte

				fmt.Print("session key: ")

				byteKey, err = terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return
				}

				fmt.Println()

				if len(byteKey) == 0 {
					err = fmt.Errorf("key not provided")
					return
				}

				sessionKey = string(byteKey)
			}

			if rawSess, err = Decrypt([]byte(sessionKey), rawSess); err != nil {
				return
			}
		}

		email, session, err = ParseSessionString(rawSess)
		if err != nil {
			return
		}
	} else {
		session, email, err = GetSessionFromUser(server)
		if err != nil {
			return
		}
	}

	return session, email, err
}

func isUnencryptedSession(in string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(strings.Split(in, ";")) == 5 && re.MatchString(strings.Split(in, ";")[0]) {
		return true
	}

	return false
}

func ParseSessionString(in string) (email string, session gosn.Session, err error) {
	if !isUnencryptedSession(in) {
		err = errors.New("session invalid, or encrypted and key was not provided")
		return
	}

	parts := strings.Split(in, ";")
	email = parts[0]
	session = gosn.Session{
		Token:  parts[2],
		Mk:     parts[4],
		Ak:     parts[3],
		Server: parts[1],
	}

	return
}

func padToAESBlockSize(b []byte) []byte {
	n := aes.BlockSize - (len(b) % aes.BlockSize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))

	return pb
}

// decrypt from base64 to decrypted string
func Decrypt(key []byte, cryptoText string) (pt string, err error) {
	var ciphertext []byte

	if ciphertext, err = base64.URLEncoding.DecodeString(cryptoText); err != nil {
		return
	}

	key = padToAESBlockSize(key)

	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	pt = fmt.Sprintf("%s", ciphertext)

	return
}

func getSessionContent(key, rawSession string) (session string, err error) {
	// check if Session is encrypted
	if len(strings.Split(rawSession, ";")) != 5 {
		if key == "" {
			fmt.Printf("encryption key: ")

			var byteKey []byte
			byteKey, err = terminal.ReadPassword(int(syscall.Stdin))

			fmt.Println()

			if err == nil {
				key = string(byteKey)
			}

			if len(strings.TrimSpace(key)) == 0 {
				err = fmt.Errorf("key required")
				return
			}
		}

		if session, err = Decrypt([]byte(key), rawSession); err != nil {
			return
		}

		if len(strings.Split(session, ";")) != 5 {
			err = fmt.Errorf("invalid session or wrong key provided")
		}
	} else {
		session = rawSession
	}

	return
}

func SessionStatus(sKey string, k keyring.Keyring) (msg string, err error) {
	var rawSession string
	rawSession, err = GetSessionFromKeyring(k)

	if err != nil {
		return
	}

	if len(rawSession) == 0 {
		return "", errors.New("keyring is empty")
	}
	// now decrypt if needed
	var session string
	session, err = getSessionContent(sKey, rawSession)

	if err != nil {
		if strings.Contains(err.Error(), "illegal base64") {
			err = errors.New("stored session is corrupt")
		}

		return
	}

	var email string
	email, _, err = ParseSessionString(session)

	if err != nil {
		msg = fmt.Sprint("failed to parse session: ", err)
		return
	}

	msg = fmt.Sprint("session found: ", email)

	return
}
