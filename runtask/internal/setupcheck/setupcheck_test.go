package setupcheck_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "keybook/pocketbaseserver/migrations"
	"keybook/runtask/internal/config"
	"keybook/runtask/internal/pbclient"
	"keybook/runtask/internal/setupcheck"
)

const seekSettingsCheckName = "seek settings"
const joraSettingsCheckName = "jora settings"
const pocketBaseSignInCheckName = "pocketbase sign-in"
const chromiumCheckName = "chromium"

var everyCheckNameInReportedOrder = []string{
	seekSettingsCheckName,
	joraSettingsCheckName,
	pocketBaseSignInCheckName,
	chromiumCheckName,
}

const seekSettingsThatLoad = `{"BaseUrl": "https://seek.example.com", "Pages": 1, "ApiParameters": []}`
const joraSettingsThatLoad = `{"BaseUrl": "https://jora.example.com", "Pages": 1, "SearchCriterias": []}`

type completeSetupWithRunningPocketBase struct {
	pocketBaseApp          core.App
	pocketBaseUrl          string
	serviceAccountEmail    string
	serviceAccountPassword string
	runtaskConfig          config.Config
}

func startTestPocketBaseWithMigrationsApplied(t *testing.T) (core.App, string) {
	t.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: t.TempDir()})
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	serverUrl := fmt.Sprintf("http://127.0.0.1:%d", port)

	var httpServer *http.Server
	app.OnServe().BindFunc(func(serveEvent *core.ServeEvent) error {
		httpServer = serveEvent.Server
		return serveEvent.Next()
	})

	app.RootCmd.SetArgs([]string{"serve", "--http", fmt.Sprintf("127.0.0.1:%d", port)})
	go func() { _ = app.Start() }()

	startDeadline, cancelStartDeadline := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartDeadline()
	for {
		select {
		case <-startDeadline.Done():
			t.Fatal("PocketBase test server did not start in time")
		default:
		}
		response, err := http.Get(serverUrl + "/api/health")
		if err == nil && response.StatusCode == http.StatusOK {
			response.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Cleanup(func() {
		if httpServer != nil {
			shutdownDeadline, cancelShutdownDeadline := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelShutdownDeadline()
			_ = httpServer.Shutdown(shutdownDeadline)
		}
	})

	return app, serverUrl
}

func createServiceAccountWithWebscraperRole(t *testing.T, app core.App) (email, password string) {
	t.Helper()
	email = "scraper-svc@example.com"
	password = "testtest123"

	usersCollection, _ := app.FindCollectionByNameOrId("_pb_users_auth_")
	user := core.NewRecord(usersCollection)
	user.Set("email", email)
	user.Set("password", password)
	user.Set("passwordConfirm", password)
	user.Set("verified", true)
	if err := app.Save(user); err != nil {
		t.Fatalf("save service account: %v", err)
	}

	webscraperRole, err := app.FindFirstRecordByData("roles", "name", "webscraper")
	if err != nil {
		t.Fatalf("find webscraper role: %v", err)
	}
	userRolesCollection, _ := app.FindCollectionByNameOrId("userRoles")
	userRole := core.NewRecord(userRolesCollection)
	userRole.Set("user", user.Id)
	userRole.Set("role", webscraperRole.Id)
	if err := app.Save(userRole); err != nil {
		t.Fatalf("assign webscraper role: %v", err)
	}
	return email, password
}

func writeFileInTemporaryFolder(t *testing.T, fileName, fileContents string) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), fileName)
	if err := os.WriteFile(filePath, []byte(fileContents), 0o600); err != nil {
		t.Fatalf("write %s: %v", fileName, err)
	}
	return filePath
}

func prepareCompleteSetupWithRunningPocketBase(t *testing.T) completeSetupWithRunningPocketBase {
	t.Helper()
	pocketBaseApp, pocketBaseUrl := startTestPocketBaseWithMigrationsApplied(t)
	serviceAccountEmail, serviceAccountPassword := createServiceAccountWithWebscraperRole(t, pocketBaseApp)
	return completeSetupWithRunningPocketBase{
		pocketBaseApp:          pocketBaseApp,
		pocketBaseUrl:          pocketBaseUrl,
		serviceAccountEmail:    serviceAccountEmail,
		serviceAccountPassword: serviceAccountPassword,
		runtaskConfig: config.Config{
			PocketBaseUrl:          pocketBaseUrl,
			ServiceAccountEmail:    serviceAccountEmail,
			ServiceAccountPassword: serviceAccountPassword,
			SeekConfigFile:         writeFileInTemporaryFolder(t, "seek.json", seekSettingsThatLoad),
			JoraConfigFile:         writeFileInTemporaryFolder(t, "jora.json", joraSettingsThatLoad),
		},
	}
}

func connectPocketBaseClient(t *testing.T, pocketBaseUrl, email, password string) *pbclient.Client {
	t.Helper()
	pocketBaseClient, err := pbclient.New(pocketBaseUrl, email, password)
	if err != nil {
		t.Fatalf("pbclient.New: %v", err)
	}
	return pocketBaseClient
}

func requireEveryCheckReportedOnceInOrder(t *testing.T, outcomes []setupcheck.CheckOutcome) map[string]error {
	t.Helper()
	if len(outcomes) != len(everyCheckNameInReportedOrder) {
		t.Fatalf("expected %d check outcomes, got %d: %+v", len(everyCheckNameInReportedOrder), len(outcomes), outcomes)
	}
	failureByCheckName := make(map[string]error, len(outcomes))
	for outcomeIndex, outcome := range outcomes {
		if outcome.CheckName != everyCheckNameInReportedOrder[outcomeIndex] {
			t.Fatalf("expected check %d to be %q, got %q", outcomeIndex, everyCheckNameInReportedOrder[outcomeIndex], outcome.CheckName)
		}
		failureByCheckName[outcome.CheckName] = outcome.Failure
	}
	return failureByCheckName
}

func requireExactlyTheseChecksFailed(t *testing.T, outcomes []setupcheck.CheckOutcome, checkNamesExpectedToFail ...string) {
	t.Helper()
	failureByCheckName := requireEveryCheckReportedOnceInOrder(t, outcomes)
	checkIsExpectedToFail := map[string]bool{}
	for _, checkName := range checkNamesExpectedToFail {
		checkIsExpectedToFail[checkName] = true
	}
	for _, checkName := range everyCheckNameInReportedOrder {
		failure := failureByCheckName[checkName]
		if checkIsExpectedToFail[checkName] && failure == nil {
			t.Errorf("expected check %q to fail, but it passed", checkName)
		}
		if !checkIsExpectedToFail[checkName] && failure != nil {
			t.Errorf("expected check %q to pass, got: %v", checkName, failure)
		}
	}
}

// Each real Chromium start competes with the scrape and extractor tests that go test runs alongside this package,
// and on a busy CI host that makes chromedp's 10-second connection limit flaky. Only the all-pass test needs a
// browser that really starts. The other failure tests use a missing one, so they start no browser and still prove
// that a failure leaves every other check reported.
func useMissingBrowserForChromiumCheck(t *testing.T) {
	t.Helper()
	setupcheck.UseBrowserExecutableForChromiumCheck(t, filepath.Join(t.TempDir(), "chrome-that-was-never-installed"))
}

func countRecordsInCollection(t *testing.T, app core.App, collectionName string) int64 {
	t.Helper()
	recordCount, err := app.CountRecords(collectionName)
	if err != nil {
		t.Fatalf("count %s: %v", collectionName, err)
	}
	return recordCount
}

func TestRunAllChecksPassesEveryCheckAndWritesNothingToPocketBaseWhenSetupIsComplete(t *testing.T) {
	setup := prepareCompleteSetupWithRunningPocketBase(t)
	collectionsThatScrapeAndReportWriteTo := []string{"sites", "jobPosts", "monthlyCountReports"}
	recordCountBeforeChecksByCollection := map[string]int64{}
	for _, collectionName := range collectionsThatScrapeAndReportWriteTo {
		recordCountBeforeChecksByCollection[collectionName] = countRecordsInCollection(t, setup.pocketBaseApp, collectionName)
	}

	outcomes := setupcheck.RunAllChecks(
		setup.runtaskConfig,
		connectPocketBaseClient(t, setup.pocketBaseUrl, setup.serviceAccountEmail, setup.serviceAccountPassword),
	)

	requireExactlyTheseChecksFailed(t, outcomes)
	for _, collectionName := range collectionsThatScrapeAndReportWriteTo {
		recordCountAfterChecks := countRecordsInCollection(t, setup.pocketBaseApp, collectionName)
		if recordCountAfterChecks != recordCountBeforeChecksByCollection[collectionName] {
			t.Errorf("expected %s to keep %d records, got %d", collectionName, recordCountBeforeChecksByCollection[collectionName], recordCountAfterChecks)
		}
	}
}

func TestRunAllChecksReportsWrongPocketBasePasswordAndStillReportsEveryOtherCheck(t *testing.T) {
	setup := prepareCompleteSetupWithRunningPocketBase(t)
	useMissingBrowserForChromiumCheck(t)

	outcomes := setupcheck.RunAllChecks(
		setup.runtaskConfig,
		connectPocketBaseClient(t, setup.pocketBaseUrl, setup.serviceAccountEmail, "not-the-service-account-password"),
	)

	requireExactlyTheseChecksFailed(t, outcomes, pocketBaseSignInCheckName, chromiumCheckName)
}

func TestRunAllChecksReportsSeekSettingsThatAreNotValidJsonAndStillReportsEveryOtherCheck(t *testing.T) {
	setup := prepareCompleteSetupWithRunningPocketBase(t)
	setup.runtaskConfig.SeekConfigFile = writeFileInTemporaryFolder(t, "seek.json", `{"BaseUrl": "https://seek.example.com",`)
	useMissingBrowserForChromiumCheck(t)

	outcomes := setupcheck.RunAllChecks(
		setup.runtaskConfig,
		connectPocketBaseClient(t, setup.pocketBaseUrl, setup.serviceAccountEmail, setup.serviceAccountPassword),
	)

	requireExactlyTheseChecksFailed(t, outcomes, seekSettingsCheckName, chromiumCheckName)
}

func TestRunAllChecksReportsMissingJoraSettingsFileAndStillReportsEveryOtherCheck(t *testing.T) {
	setup := prepareCompleteSetupWithRunningPocketBase(t)
	setup.runtaskConfig.JoraConfigFile = filepath.Join(t.TempDir(), "jora-settings-that-were-never-copied.json")
	useMissingBrowserForChromiumCheck(t)

	outcomes := setupcheck.RunAllChecks(
		setup.runtaskConfig,
		connectPocketBaseClient(t, setup.pocketBaseUrl, setup.serviceAccountEmail, setup.serviceAccountPassword),
	)

	requireExactlyTheseChecksFailed(t, outcomes, joraSettingsCheckName, chromiumCheckName)
}

// chromedp also tries /usr/local/bin/chrome by absolute path, which is where OpenBSD's package installs it, so
// emptying PATH cannot hide Chromium. Pointing the check at a missing executable is what a host without the
// package looks like to chromedp.
func TestRunAllChecksReportsChromiumThatCannotBeStartedAndStillReportsEveryOtherCheck(t *testing.T) {
	setup := prepareCompleteSetupWithRunningPocketBase(t)
	useMissingBrowserForChromiumCheck(t)

	outcomes := setupcheck.RunAllChecks(
		setup.runtaskConfig,
		connectPocketBaseClient(t, setup.pocketBaseUrl, setup.serviceAccountEmail, setup.serviceAccountPassword),
	)

	requireExactlyTheseChecksFailed(t, outcomes, chromiumCheckName)
}
