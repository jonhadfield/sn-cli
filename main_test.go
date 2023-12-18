package sncli

import (
	"fmt"
	"github.com/jonhadfield/gosn-v2/auth"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
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
	ts, err := auth.CliSignIn(email, password, server, false)
	if err != nil {
		log.Fatal(err)
	}

	if server == "" {
		server = SNServerURL
	}

	httpClient := common.NewHTTPClient()
	httpClient.Logger = nil

	gTtestSession = &session.Session{
		Debug:             false,
		HTTPClient:        httpClient,
		SchemaValidation:  false,
		Server:            server,
		FilesServerUrl:    ts.FilesServerUrl,
		Token:             "",
		MasterKey:         ts.MasterKey,
		ItemsKeys:         nil,
		DefaultItemsKey:   session.SessionItemsKey{},
		KeyParams:         auth.KeyParams{},
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
		signIn(SNServerURL, os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"))
	}

	if _, err := items.Sync(items.SyncInput{Session: gTtestSession}); err != nil {
		log.Fatal(err)
	}

	if gTtestSession.DefaultItemsKey.ItemsKey == "" {
		panic("failed in TestMain due to empty default items key")
	}
	if strings.TrimSpace(gTtestSession.Server) == "" {
		panic("failed in TestMain due to empty server")
	}

	var err error
	testSession, err = cache.ImportSession(&auth.SignInResponseDataSession{
		Debug:             gTtestSession.Debug,
		HTTPClient:        gTtestSession.HTTPClient,
		SchemaValidation:  false,
		Server:            gTtestSession.Server,
		FilesServerUrl:    gTtestSession.FilesServerUrl,
		Token:             "",
		MasterKey:         gTtestSession.MasterKey,
		KeyParams:         gTtestSession.KeyParams,
		AccessToken:       gTtestSession.AccessToken,
		RefreshToken:      gTtestSession.RefreshToken,
		AccessExpiration:  gTtestSession.AccessExpiration,
		RefreshExpiration: gTtestSession.RefreshExpiration,
		ReadOnlyAccess:    gTtestSession.ReadOnlyAccess,
		PasswordNonce:     gTtestSession.PasswordNonce,
	}, "")
	if err != nil {
		return
	}

	testSession.CacheDBPath, err = cache.GenCacheDBPath(*testSession, "", common.LibName)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

// prevent throttling when using official server.
func testDelay() {
	if strings.Contains(os.Getenv("SN_SERVER"), "api.standardnotes.com") {
		time.Sleep(2 * time.Second)
	}
}
