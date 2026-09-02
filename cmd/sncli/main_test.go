package main

import (
	"fmt"
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
		// Debug:     true,
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
		fmt.Println(err)

		os.Exit(1)
	}

	if server == "" {
		server = session.SNServerURL
	}

	httpClient := common.NewHTTPClient()

	debug := false
	if !debug {
		httpClient.Logger = nil
	}

	ts.HTTPClient = httpClient
	if httpClient.Logger != nil {
		panic("httpClient.Logger should be nil")
	}

	gTtestSession = &session.Session{
		Debug:             debug,
		HTTPClient:        httpClient,
		SchemaValidation:  false,
		Server:            server,
		FilesServerUrl:    ts.FilesServerUrl,
		Token:             "",
		MasterKey:         ts.MasterKey,
		ItemsKeys:         nil,
		DefaultItemsKey:   session.SessionItemsKey{},
		KeyParams:         ts.KeyParams,
		AccessToken:       ts.AccessToken,
		RefreshToken:      ts.RefreshToken,
		AccessExpiration:  ts.AccessExpiration,
		RefreshExpiration: ts.RefreshExpiration,
		ReadOnlyAccess:    ts.ReadOnlyAccess,
		PasswordNonce:     ts.PasswordNonce,
		Schemas:           nil,
		// Standard Notes issues "2:"-prefixed cookie-based access tokens, and
		// the sync request only authenticates if the Cookie header goes with
		// the bearer token. Dropping these yields a 401 from syncItemsViaAPI.
		AccessTokenCookie:  ts.AccessTokenCookie,
		RefreshTokenCookie: ts.RefreshTokenCookie,
	}

	testSession = &cache.Session{
		Session:     gTtestSession,
		CacheDB:     nil,
		CacheDBPath: "",
	}
}

// integrationCredentialsAvailable reports whether the live-server tests should
// run. They add and delete real items on a real Standard Notes account, so they
// are opt-in: set SN_INTEGRATION_TESTS=1 as well as supplying credentials.
func integrationCredentialsAvailable() bool {
	switch strings.ToLower(os.Getenv("SN_INTEGRATION_TESTS")) {
	case "1", "true", "yes":
	default:
		return false
	}

	// The local ramea server registers a throwaway account of its own.
	if strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
		return true
	}

	return os.Getenv("SN_EMAIL") != "" && os.Getenv("SN_PASSWORD") != ""
}

// requireIntegration skips a test that needs a live Standard Notes server.
func requireIntegration(t *testing.T) {
	t.Helper()

	if !integrationCredentialsAvailable() {
		t.Skip("live-server test: set SN_INTEGRATION_TESTS=1 with SN_EMAIL " +
			"and SN_PASSWORD to run it")
	}
}

func TestMain(m *testing.M) {
	// Without credentials there is no session to build, but the unit tests in
	// this package do not need one, so still run them. Tests that do need a
	// server call requireIntegration and skip themselves.
	if !integrationCredentialsAvailable() {
		fmt.Fprintln(os.Stderr,
			"SN_INTEGRATION_TESTS not set: running cmd/sncli unit tests only")

		os.Exit(m.Run())
	}

	if strings.Contains(os.Getenv("SN_SERVER"), "ramea") {
		localTestMain()
	} else {
		signIn(session.SNServerURL, os.Getenv("SN_EMAIL"), os.Getenv("SN_PASSWORD"))
	}

	if _, err := items.Sync(items.SyncInput{Session: gTtestSession}); err != nil {
		fmt.Println(err)

		os.Exit(1)
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

		AccessTokenCookie:  gTtestSession.AccessTokenCookie,
		RefreshTokenCookie: gTtestSession.RefreshTokenCookie,
	}, "")
	if err != nil {
		fmt.Println(err)

		os.Exit(1)
	}

	testSession.CacheDBPath, err = cache.GenCacheDBPath(*testSession, "", common.LibName)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
