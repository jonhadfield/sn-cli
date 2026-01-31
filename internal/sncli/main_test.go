package sncli

import (
	"fmt"
	"log" // nolint:depguard // log is acceptable in test files for setup/teardown
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jonhadfield/gosn-v2/auth"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
)

var (
	testSession      *cache.Session
	gTtestSession    *session.Session
	testUserEmail    string
	testUserPassword string
)

func localTestMain() {
	localServer := "http://ramea:3000"
	testUserEmail = fmt.Sprintf("ramea-%s", strconv.FormatInt(time.Now().UnixNano(), 16))
	testUserPassword = "secretsanta"

	rInput := auth.RegisterInput{
		Password:  testUserPassword,
		Email:     testUserEmail,
		APIServer: localServer,
		Version:   "004",
		Debug:     true,
	}

	_, err := rInput.Register()
	if err != nil {
		panic(fmt.Sprintf("failed to register with: %s", localServer))
	}

	signIn(localServer, testUserEmail, testUserPassword)
}

func signIn(server, email, password string) {
	// Enable debugging if requested
	debug := os.Getenv("SN_DEBUG") == "true"

	ts, err := auth.CliSignIn(email, password, server, debug)
	if err != nil {
		log.Fatal(err)
	}

	if server == "" {
		server = SNServerURL
	}

	// Enable schema validation if requested
	schemaValidation := os.Getenv("SN_SCHEMA_VALIDATION") == "yes" || os.Getenv("SN_SCHEMA_VALIDATION") == "true"

	// Disable logger if not debugging
	if !debug {
		ts.HTTPClient.Logger = nil
	}

	gTtestSession = &session.Session{
		Debug:            debug,
		HTTPClient:       ts.HTTPClient,
		SchemaValidation: schemaValidation,
		Server:           server,
		FilesServerUrl:   ts.FilesServerUrl,
		Token:            ts.Token,
		MasterKey:        ts.MasterKey,
		// ItemsKeys:         nil,
		// DefaultItemsKey:   session.SessionItemsKey{},
		KeyParams:         ts.KeyParams,
		AccessToken:       ts.AccessToken,
		RefreshToken:      ts.RefreshToken,
		AccessExpiration:  ts.AccessExpiration,
		RefreshExpiration: ts.RefreshExpiration,
		ReadOnlyAccess:    ts.ReadOnlyAccess,
		PasswordNonce:     ts.PasswordNonce,
		Schemas:           nil,
	}

	testSession = &cache.Session{
		Session:     gTtestSession,
		CacheDB:     nil,
		CacheDBPath: "",
	}
}

func TestMain(m *testing.M) {
	// if os.Getenv("SN_SERVER") == "" || strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
	if strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
		localTestMain()
	} else {
		email := os.Getenv("SN_EMAIL")
		password := os.Getenv("SN_PASSWORD")
		if email == "" {
			email = "gosn-v2-202509231858@lessknown.co.uk"
		}
		if password == "" {
			password = "gosn-v2-202509231858@lessknown.co.uk"
		}
		signIn(SNServerURL, email, password)
	}

	// Add delay to prevent rate limiting during initial sync
	time.Sleep(2 * time.Second)

	si := items.SyncInput{
		Session: gTtestSession,
	}

	_, syncErr := items.Sync(si)
	if syncErr != nil {
		log.Printf("WARNING: Initial sync failed: %v", syncErr)

		newKey, createErr := items.CreateItemsKey()
		if createErr != nil {
			log.Printf("ERROR: Failed to create new items key: %v", createErr)
			gTtestSession.DefaultItemsKey.ItemsKey = "test-placeholder-key"
		} else {
			gTtestSession.DefaultItemsKey = session.SessionItemsKey{
				ItemsKey: newKey.Content.ItemsKey,
				UUID:     newKey.UUID,
			}
			gTtestSession.ItemsKeys = append(gTtestSession.ItemsKeys, gTtestSession.DefaultItemsKey)
		}
	}

	if gTtestSession.DefaultItemsKey.ItemsKey == "" {
		panic("failed in TestMain due to empty default items key")
	}
	if strings.TrimSpace(gTtestSession.Server) == "" {
		panic("failed in TestMain due to empty server")
	}

	var importErr error
	testSession, importErr = cache.ImportSession(&auth.SignInResponseDataSession{
		Debug:             gTtestSession.Debug,
		HTTPClient:        gTtestSession.HTTPClient,
		SchemaValidation:  gTtestSession.SchemaValidation,
		Server:            gTtestSession.Server,
		FilesServerUrl:    gTtestSession.FilesServerUrl,
		Token:             gTtestSession.Token,
		MasterKey:         gTtestSession.MasterKey,
		KeyParams:         gTtestSession.KeyParams,
		AccessToken:       gTtestSession.AccessToken,
		RefreshToken:      gTtestSession.RefreshToken,
		AccessExpiration:  gTtestSession.AccessExpiration,
		RefreshExpiration: gTtestSession.RefreshExpiration,
		ReadOnlyAccess:    gTtestSession.ReadOnlyAccess,
		PasswordNonce:     gTtestSession.PasswordNonce,
	}, "")
	if importErr != nil {
		return
	}

	// Copy the items keys from gTtestSession to testSession
	testSession.Session.DefaultItemsKey = gTtestSession.DefaultItemsKey
	testSession.Session.ItemsKeys = gTtestSession.ItemsKeys

	testSession.CacheDBPath, importErr = cache.GenCacheDBPath(*testSession, "", common.LibName)
	if importErr != nil {
		panic(importErr)
	}

	// Run the tests
	exitCode := m.Run()

	// Clean up before exiting
	// Close any open cache database connections
	if testSession != nil && testSession.CacheDB != nil {
		if err := testSession.CacheDB.Close(); err != nil {
			log.Printf("WARNING: Failed to close cache database: %v", err)
		}
	}

	// Remove the cache database file
	if testSession != nil && testSession.CacheDBPath != "" {
		if err := os.Remove(testSession.CacheDBPath); err != nil && !os.IsNotExist(err) {
			log.Printf("WARNING: Failed to remove cache database file: %v", err)
		}
	}

	os.Exit(exitCode)
}

// prevent throttling when using official server.
func testDelay() {
	if strings.Contains(os.Getenv("SN_SERVER"), "api.standardnotes.com") {
		time.Sleep(5 * time.Second)
	}
}
