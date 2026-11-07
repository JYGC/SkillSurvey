package deploy_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const exampleSiteAddress = "http://site.example.test"
const examplePocketBaseAddress = "http://site.example.test:8090"
const exampleSecondPocketBaseAddress = "http://api.example.test:8090"
const exampleSecondSiteAddress = "http://www.example.test"
const documentationOnlyListenAddress = "192.0.2.10"

const installedPocketBaseProgram = "usr/local/libexec/skillsurvey/pocketbaseserver"
const pocketBaseDataFolder = "var/skillsurvey/pb_data"
const pocketBaseSnapshotsFolder = "var/skillsurvey/snapshots"
const serverInstallRecord = "var/db/skillsurvey/BUNDLE-INFO"
const serverAddressesFile = "etc/skillsurvey/server.conf"
const webRoot = "var/www/htdocs/skillsurvey"

const expectedHttpdConfigurationForEveryAddress = `# Written by install-server.sh, which replaces this file.
server "default" {
	listen on * port 80
	root "/htdocs/skillsurvey/current"
	location not found "/*" {
		request rewrite "/index.html"
	}
}
`

type serverBundle struct {
	scriptPath     string
	directory      string
	commit         string
	programContent string
}

func (fixture *installFixture) buildServerBundle(commitLabel, programVersion, frontendVersion string) serverBundle {
	fixture.t.Helper()
	commit := commitNamed(commitLabel)
	bundleDirectory := filepath.Join(fixture.t.TempDir(), "skillsurvey-server-"+commit[:7])
	programContent := "fake pocketbaseserver\npocketbaseserver-version-" + programVersion + "\n"

	writeFileCreatingFolders(fixture.t, filepath.Join(bundleDirectory, "BUNDLE-INFO"), bundleInformation(commit), 0o644)
	copyFileFromDeploy(fixture.t, "install-server.sh", filepath.Join(bundleDirectory, "install-server.sh"), 0o644)
	copyFileFromDeploy(fixture.t, "install-common.sh", filepath.Join(bundleDirectory, "install-common.sh"), 0o644)
	copyFileFromDeploy(fixture.t, "pocketbaseserver.rc", filepath.Join(bundleDirectory, "pocketbaseserver.rc"), 0o644)
	writeFileCreatingFolders(fixture.t, filepath.Join(bundleDirectory, "pocketbaseserver"), programContent, 0o755)
	writeFileCreatingFolders(fixture.t, filepath.Join(bundleDirectory, "frontend", "index.html"),
		"<!DOCTYPE html><title>SkillSurvey</title><div id=\"app\"></div><!-- frontend-version-"+frontendVersion+" -->\n", 0o644)
	writeFileCreatingFolders(fixture.t, filepath.Join(bundleDirectory, "frontend", "js", "app."+frontendVersion+".js"),
		"console.log('frontend "+frontendVersion+"');\n", 0o644)
	writeChecksumFile(fixture.t, bundleDirectory)

	return serverBundle{
		scriptPath:     filepath.Join(bundleDirectory, "install-server.sh"),
		directory:      bundleDirectory,
		commit:         commit,
		programContent: programContent,
	}
}

func everyServerAddressOption() []string {
	return []string{
		"--listen-address", "0.0.0.0",
		"--site-address", exampleSiteAddress,
		"--pocketbase-address", examplePocketBaseAddress,
	}
}

func expectedPocketBaseFlags(listenAddress, siteAddress string) string {
	return "serve --http=" + listenAddress + ":8090 --origins=" + siteAddress + " --dir=/var/skillsurvey/pb_data"
}

func expectedHostConfiguration(pocketBaseAddress string) string {
	return "{ \"PocketBaseUrl\": \"" + pocketBaseAddress + "\" }\n"
}

func (fixture *installFixture) liveReleaseLink() string {
	fixture.t.Helper()
	linkTarget, err := os.Readlink(fixture.inRoot(webRoot + "/current"))
	if err != nil {
		fixture.t.Fatal(err)
	}
	return linkTarget
}

func (fixture *installFixture) putDatabaseInPocketBaseData() {
	fixture.t.Helper()
	writeFileCreatingFolders(fixture.t, fixture.inRoot(pocketBaseDataFolder+"/data.db"), "records written before the upgrade\n", 0o644)
	writeFileCreatingFolders(fixture.t, fixture.inRoot(pocketBaseDataFolder+"/storage/upload.txt"), "an uploaded file\n", 0o644)
}

func (fixture *installFixture) snapshotNames() []string {
	fixture.t.Helper()
	entries, err := os.ReadDir(fixture.inRoot(pocketBaseSnapshotsFolder))
	if err != nil {
		fixture.t.Fatal(err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func indexOfFirstCallStartingWith(t *testing.T, calls []string, prefix string) int {
	t.Helper()
	for callIndex, call := range calls {
		if strings.HasPrefix(call, prefix) {
			return callIndex
		}
	}
	t.Fatalf("expected a call starting with %q, got %v", prefix, calls)
	return -1
}

// rc.d finds a running service by its exact command line, so flags changed while it runs make the stop miss it.
func assertFlagsWereOnlySetWhilePocketBaseWasStopped(t *testing.T, calls []string) {
	t.Helper()
	for _, setFlagsCall := range callsStartingWith(calls, "rcctl set pocketbaseserver flags") {
		if !strings.HasPrefix(setFlagsCall, "rcctl set pocketbaseserver flags (running=no ") {
			t.Errorf("flags were set while pocketbaseserver was running: %q", setFlagsCall)
		}
	}
}

func TestFirstServerInstallSetsUpBothServicesAndInstallsNoPackages(t *testing.T) {
	fixture := newInstallFixture(t)
	bundle := fixture.buildServerBundle("first", "1", "1")

	output, calls := fixture.mustRunScript(bundle.scriptPath, everyServerAddressOption()...)

	useraddCalls := callsStartingWith(calls, "useradd ")
	if len(useraddCalls) != 1 {
		t.Fatalf("expected one useradd call, got %v", useraddCalls)
	}
	for _, expectedArgument := range []string{"-u 498", "-L daemon", "-d /var/skillsurvey", "-s /sbin/nologin", "_skillsurvey"} {
		if !strings.Contains(useraddCalls[0], expectedArgument) {
			t.Errorf("useradd call %q lacks %q", useraddCalls[0], expectedArgument)
		}
	}
	if fixture.readFileInRoot(installedPocketBaseProgram) != bundle.programContent {
		t.Error("installed pocketbaseserver differs from the bundle's")
	}
	assertMode(t, fixture.inRoot(installedPocketBaseProgram), 0o755)
	serviceDefinition, _ := os.ReadFile("pocketbaseserver.rc")
	if fixture.readFileInRoot("etc/rc.d/pocketbaseserver") != string(serviceDefinition) {
		t.Error("/etc/rc.d/pocketbaseserver differs from pocketbaseserver.rc")
	}
	assertMode(t, fixture.inRoot("etc/rc.d/pocketbaseserver"), 0o555)
	assertMode(t, fixture.inRoot(pocketBaseDataFolder), 0o700)
	assertMode(t, fixture.inRoot(pocketBaseSnapshotsFolder), 0o700)
	if len(callsStartingWith(calls, "chown _skillsurvey:_skillsurvey "+fixture.inRoot(pocketBaseDataFolder))) != 1 {
		t.Errorf("expected pb_data to be given to _skillsurvey, got %v", callsStartingWith(calls, "chown "))
	}
	expectedAddresses := "listen_address=0.0.0.0\nsite_address=" + exampleSiteAddress + "\npocketbase_address=" + examplePocketBaseAddress + "\n"
	if addresses := fixture.readFileInRoot(serverAddressesFile); addresses != expectedAddresses {
		t.Errorf("server.conf is:\n%s\nwant:\n%s", addresses, expectedAddresses)
	}
	if httpdConfiguration := fixture.readFileInRoot("etc/httpd.conf"); httpdConfiguration != expectedHttpdConfigurationForEveryAddress {
		t.Errorf("httpd.conf is:\n%s\nwant:\n%s", httpdConfiguration, expectedHttpdConfigurationForEveryAddress)
	}
	checkedConfiguration, err := os.ReadFile(filepath.Join(fixture.stubStateDirectory, "last-checked-httpd.conf"))
	if err != nil || string(checkedConfiguration) != expectedHttpdConfigurationForEveryAddress {
		t.Errorf("httpd -n did not check the configuration that was installed: %v", err)
	}
	if fixture.liveReleaseLink() != "releases/"+bundle.commit[:7] {
		t.Errorf("current points at %q, want releases/%s", fixture.liveReleaseLink(), bundle.commit[:7])
	}
	bundleStartPage, _ := os.ReadFile(filepath.Join(bundle.directory, "frontend", "index.html"))
	if fixture.readFileInRoot(webRoot+"/current/index.html") != string(bundleStartPage) {
		t.Error("the live start page differs from the bundle's")
	}
	if hostConfiguration := fixture.readFileInRoot(webRoot + "/current/host-config.json"); hostConfiguration != expectedHostConfiguration(examplePocketBaseAddress) {
		t.Errorf("host-config.json is %q", hostConfiguration)
	}
	for _, serviceName := range []string{"pocketbaseserver", "httpd"} {
		if !fixture.serviceIsEnabled(serviceName) || !fixture.serviceIsRunning(serviceName) {
			t.Errorf("%s should be enabled and running", serviceName)
		}
	}
	if flags := fixture.serviceFlags("pocketbaseserver"); flags != expectedPocketBaseFlags("0.0.0.0", exampleSiteAddress) {
		t.Errorf("pocketbaseserver flags are %q", flags)
	}
	if indexOfFirstCallStartingWith(t, calls, "rcctl set pocketbaseserver flags (running=no") > indexOfFirstCallStartingWith(t, calls, "rcctl start pocketbaseserver") {
		t.Error("flags were set after pocketbaseserver started")
	}
	assertFlagsWereOnlySetWhilePocketBaseWasStopped(t, calls)
	for _, checkedAddress := range []string{"http://127.0.0.1:8090/api/health", "http://127.0.0.1/", "http://127.0.0.1/user/monthly-count-report", "http://127.0.0.1/host-config.json"} {
		if len(callsStartingWith(calls, "ftp "+checkedAddress)) == 0 {
			t.Errorf("expected a check of %s, got %v", checkedAddress, callsStartingWith(calls, "ftp "))
		}
	}
	assertNoCallStartsWith(t, calls, "pkg_add")
	if fixture.readFileInRoot(serverInstallRecord) != bundleInformation(bundle.commit) {
		t.Error("the install record differs from the bundle's BUNDLE-INFO")
	}
	assertOutputContains(t, output, "superuser upsert", "runtask@skillsurvey.com", "another machine")
}

func TestSecondServerRunWithTheSameBundleChangesNothingAndRestartsNothing(t *testing.T) {
	fixture := newInstallFixture(t)
	bundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(bundle.scriptPath, everyServerAddressOption()...)
	before := fixture.fingerprintRoot()

	output, calls := fixture.mustRunScript(bundle.scriptPath, everyServerAddressOption()...)

	assertNothingChanged(t, before, fixture.fingerprintRoot())
	assertNoCallStartsWith(t, calls, "rcctl ", "useradd", "chown", "httpd ")
	assertOutputContains(t, output, "0 changes")
	if strings.Contains(output, "superuser upsert") {
		t.Error("the steps left by hand were printed again on a second run")
	}
}

func TestServerAddressesPersistWhenLaterRunsOmitThem(t *testing.T) {
	fixture := newInstallFixture(t)
	bundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(bundle.scriptPath, everyServerAddressOption()...)
	before := fixture.fingerprintRoot()

	output, calls := fixture.mustRunScript(bundle.scriptPath)

	assertNothingChanged(t, before, fixture.fingerprintRoot())
	assertNoCallStartsWith(t, calls, "rcctl ")
	assertOutputContains(t, output, "0 changes")
}

func TestChangedPocketBaseAddressRewritesOnlyTheLiveHostConfiguration(t *testing.T) {
	fixture := newInstallFixture(t)
	bundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(bundle.scriptPath, everyServerAddressOption()...)
	liveReleaseBefore := fixture.liveReleaseLink()

	output, calls := fixture.mustRunScript(bundle.scriptPath, "--pocketbase-address", exampleSecondPocketBaseAddress)

	if hostConfiguration := fixture.readFileInRoot(webRoot + "/current/host-config.json"); hostConfiguration != expectedHostConfiguration(exampleSecondPocketBaseAddress) {
		t.Errorf("host-config.json is %q", hostConfiguration)
	}
	if fixture.liveReleaseLink() != liveReleaseBefore {
		t.Errorf("current moved from %s to %s for an address change", liveReleaseBefore, fixture.liveReleaseLink())
	}
	if !strings.Contains(fixture.readFileInRoot(serverAddressesFile), "pocketbase_address="+exampleSecondPocketBaseAddress+"\n") {
		t.Error("server.conf does not hold the new PocketBase address")
	}
	assertNoCallStartsWith(t, calls, "rcctl stop", "rcctl start", "rcctl set", "rcctl reload")
	assertOutputContains(t, output, "2 changes")
}

func TestChangedSiteAddressSetsFlagsOnlyWhilePocketBaseIsStoppedAndTakesNoSnapshot(t *testing.T) {
	fixture := newInstallFixture(t)
	bundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(bundle.scriptPath, everyServerAddressOption()...)
	fixture.putDatabaseInPocketBaseData()

	_, calls := fixture.mustRunScript(bundle.scriptPath, "--site-address", exampleSecondSiteAddress)

	assertFlagsWereOnlySetWhilePocketBaseWasStopped(t, calls)
	if flags := fixture.serviceFlags("pocketbaseserver"); flags != expectedPocketBaseFlags("0.0.0.0", exampleSecondSiteAddress) {
		t.Errorf("pocketbaseserver flags are %q", flags)
	}
	if !fixture.serviceIsRunning("pocketbaseserver") {
		t.Error("pocketbaseserver is not running")
	}
	if snapshotNames := fixture.snapshotNames(); len(snapshotNames) != 0 {
		t.Errorf("a flags-only change took snapshots %v", snapshotNames)
	}
}

func TestServerUpgradeStopsSnapshotsReplacesSetsFlagsAndStartsInThatOrderLeavingDataUntouched(t *testing.T) {
	fixture := newInstallFixture(t)
	firstBundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(firstBundle.scriptPath, everyServerAddressOption()...)
	fixture.putDatabaseInPocketBaseData()
	dataBeforeUpgrade := fingerprintEveryPathUnder(t, fixture.inRoot(pocketBaseDataFolder))
	secondBundle := fixture.buildServerBundle("second", "2", "2")

	output, calls := fixture.mustRunScript(secondBundle.scriptPath, "--site-address", exampleSecondSiteAddress)

	pocketBaseCalls := callsStartingWith(calls, "rcctl ")
	stopIndex := indexOfFirstCallStartingWith(t, pocketBaseCalls, "rcctl stop pocketbaseserver (running=yes program=pocketbaseserver-version-1 snapshots=0")
	setFlagsIndex := indexOfFirstCallStartingWith(t, pocketBaseCalls, "rcctl set pocketbaseserver flags (running=no program=pocketbaseserver-version-2 snapshots=1")
	startIndex := indexOfFirstCallStartingWith(t, pocketBaseCalls, "rcctl start pocketbaseserver (running=no program=pocketbaseserver-version-2 snapshots=1 flags="+expectedPocketBaseFlags("0.0.0.0", exampleSecondSiteAddress))
	if !(stopIndex < setFlagsIndex && setFlagsIndex < startIndex) {
		t.Errorf("expected stop, then set flags, then start, got %v", pocketBaseCalls)
	}
	assertFlagsWereOnlySetWhilePocketBaseWasStopped(t, calls)
	snapshotNames := fixture.snapshotNames()
	if len(snapshotNames) != 1 || !strings.HasSuffix(snapshotNames[0], "-"+firstBundle.commit[:7]) {
		t.Fatalf("expected one snapshot named after %s, got %v", firstBundle.commit[:7], snapshotNames)
	}
	snapshotOfData := fingerprintEveryPathUnder(t, filepath.Join(fixture.inRoot(pocketBaseSnapshotsFolder), snapshotNames[0], "pb_data"))
	if snapshotDatabase, _ := os.ReadFile(filepath.Join(fixture.inRoot(pocketBaseSnapshotsFolder), snapshotNames[0], "pb_data", "data.db")); string(snapshotDatabase) != "records written before the upgrade\n" || len(snapshotOfData) != len(dataBeforeUpgrade) {
		t.Errorf("the snapshot does not hold the data from before the upgrade")
	}
	assertNothingChanged(t, dataBeforeUpgrade, fingerprintEveryPathUnder(t, fixture.inRoot(pocketBaseDataFolder)))
	if fixture.readFileInRoot(installedPocketBaseProgram) != secondBundle.programContent {
		t.Error("pocketbaseserver was not replaced")
	}
	if fixture.readFileInRoot(serverInstallRecord) != bundleInformation(secondBundle.commit) {
		t.Error("the install record still names the first bundle")
	}
	assertOutputContains(t, output, snapshotNames[0])
}

func TestFrontendOnlyUpgradeLeavesPocketBaseRunningAndKeepsThePreviousRelease(t *testing.T) {
	fixture := newInstallFixture(t)
	firstBundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(firstBundle.scriptPath, everyServerAddressOption()...)
	fixture.putDatabaseInPocketBaseData()
	secondBundle := fixture.buildServerBundle("second", "1", "2")

	_, calls := fixture.mustRunScript(secondBundle.scriptPath)

	assertNoCallStartsWith(t, calls, "rcctl stop", "rcctl start pocketbaseserver", "rcctl set", "rcctl reload")
	if snapshotNames := fixture.snapshotNames(); len(snapshotNames) != 0 {
		t.Errorf("a frontend-only upgrade took snapshots %v", snapshotNames)
	}
	if fixture.liveReleaseLink() != "releases/"+secondBundle.commit[:7] {
		t.Errorf("current points at %q, want the second release", fixture.liveReleaseLink())
	}
	if !strings.Contains(fixture.readFileInRoot(webRoot+"/current/index.html"), "frontend-version-2") {
		t.Error("the live start page is not the second bundle's")
	}
	if hostConfiguration := fixture.readFileInRoot(webRoot + "/current/host-config.json"); hostConfiguration != expectedHostConfiguration(examplePocketBaseAddress) {
		t.Errorf("the new release's host-config.json is %q", hostConfiguration)
	}
	if !strings.Contains(fixture.readFileInRoot(webRoot+"/previous/index.html"), "frontend-version-1") {
		t.Error("the previous release was not kept")
	}
	if fixture.readFileInRoot(serverInstallRecord) != bundleInformation(secondBundle.commit) {
		t.Error("the install record still names the first bundle")
	}
}

func TestServerKeepsOnlyTheCurrentAndPreviousReleases(t *testing.T) {
	fixture := newInstallFixture(t)
	for bundleNumber, frontendVersion := range []string{"1", "2", "3"} {
		bundle := fixture.buildServerBundle("bundle-"+frontendVersion, "1", frontendVersion)
		if bundleNumber == 0 {
			fixture.mustRunScript(bundle.scriptPath, everyServerAddressOption()...)
		} else {
			fixture.mustRunScript(bundle.scriptPath)
		}
	}

	releaseEntries, err := os.ReadDir(fixture.inRoot(webRoot + "/releases"))
	if err != nil {
		t.Fatal(err)
	}
	releaseNames := []string{}
	for _, entry := range releaseEntries {
		releaseNames = append(releaseNames, entry.Name())
	}
	expectedReleaseNames := []string{commitNamed("bundle-2")[:7], commitNamed("bundle-3")[:7]}
	if len(releaseNames) != 2 || !strings.Contains(strings.Join(releaseNames, " "), expectedReleaseNames[0]) || !strings.Contains(strings.Join(releaseNames, " "), expectedReleaseNames[1]) {
		t.Errorf("releases are %v, want only %v", releaseNames, expectedReleaseNames)
	}
}

func TestFailedSnapshotStartsTheInstalledProgramAgainAndReplacesNothing(t *testing.T) {
	fixture := newInstallFixture(t)
	firstBundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(firstBundle.scriptPath, everyServerAddressOption()...)
	fixture.putDatabaseInPocketBaseData()
	unreadableDataFile := fixture.inRoot(pocketBaseDataFolder + "/unreadable.db")
	writeFileCreatingFolders(t, unreadableDataFile, "cannot be copied by a non-root test\n", 0o000)
	secondBundle := fixture.buildServerBundle("second", "2", "2")

	output, calls := fixture.mustFailScript(secondBundle.scriptPath)

	if fixture.readFileInRoot(installedPocketBaseProgram) != firstBundle.programContent {
		t.Error("pocketbaseserver was replaced although the snapshot failed")
	}
	if !fixture.serviceIsRunning("pocketbaseserver") {
		t.Error("the installed pocketbaseserver was not started again")
	}
	if len(callsStartingWith(calls, "rcctl start pocketbaseserver (running=no program=pocketbaseserver-version-1")) != 1 {
		t.Errorf("expected the first program to be started again, got %v", callsStartingWith(calls, "rcctl "))
	}
	if snapshotNames := fixture.snapshotNames(); len(snapshotNames) != 0 {
		t.Errorf("a partial snapshot was left behind: %v", snapshotNames)
	}
	assertOutputContains(t, output, "snapshot")
}

func TestFailedHealthCheckNamesTheSnapshotPrintsHowToGoBackAndRestoresNothing(t *testing.T) {
	fixture := newInstallFixture(t)
	firstBundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(firstBundle.scriptPath, everyServerAddressOption()...)
	fixture.putDatabaseInPocketBaseData()
	dataBeforeUpgrade := fingerprintEveryPathUnder(t, fixture.inRoot(pocketBaseDataFolder))
	secondBundle := fixture.buildServerBundle("second", "2", "2")
	fixture.setStubState("fail-health", "")

	output, _ := fixture.mustFailScript(secondBundle.scriptPath)

	snapshotNames := fixture.snapshotNames()
	if len(snapshotNames) != 1 {
		t.Fatalf("expected one snapshot, got %v", snapshotNames)
	}
	assertOutputContains(t, output,
		"/var/skillsurvey/snapshots/"+snapshotNames[0],
		"doas rcctl stop pocketbaseserver",
		"doas mv /var/skillsurvey/pb_data /var/skillsurvey/pb_data.failed",
		"doas cp -Rp /var/skillsurvey/snapshots/"+snapshotNames[0]+"/pb_data /var/skillsurvey/pb_data",
	)
	assertNothingChanged(t, dataBeforeUpgrade, fingerprintEveryPathUnder(t, fixture.inRoot(pocketBaseDataFolder)))
	if fixture.readFileInRoot(installedPocketBaseProgram) != secondBundle.programContent {
		t.Error("the script put the old program back by itself")
	}
}

func TestServerUpgradeKeepsOnlyTheThreeNewestSnapshotsItMade(t *testing.T) {
	fixture := newInstallFixture(t)
	firstBundle := fixture.buildServerBundle("first", "1", "1")
	fixture.mustRunScript(firstBundle.scriptPath, everyServerAddressOption()...)
	fixture.putDatabaseInPocketBaseData()
	olderSnapshotNames := []string{"20250101T000000Z-aaaaaaa", "20250201T000000Z-bbbbbbb", "20250301T000000Z-ccccccc"}
	for _, snapshotName := range olderSnapshotNames {
		writeFileCreatingFolders(t, fixture.inRoot(pocketBaseSnapshotsFolder+"/"+snapshotName+"/pb_data/data.db"), "old\n", 0o644)
	}
	writeFileCreatingFolders(t, fixture.inRoot(pocketBaseSnapshotsFolder+"/kept-by-hand/pb_data/data.db"), "mine\n", 0o644)
	secondBundle := fixture.buildServerBundle("second", "2", "2")

	fixture.mustRunScript(secondBundle.scriptPath)

	remainingNames := strings.Join(fixture.snapshotNames(), " ")
	for _, removedName := range olderSnapshotNames[:1] {
		if strings.Contains(remainingNames, removedName) {
			t.Errorf("snapshot %s should have been removed, remaining: %s", removedName, remainingNames)
		}
	}
	for _, keptName := range append(olderSnapshotNames[1:], "kept-by-hand", "-"+firstBundle.commit[:7]) {
		if !strings.Contains(remainingNames, keptName) {
			t.Errorf("snapshot %s should have been kept, remaining: %s", keptName, remainingNames)
		}
	}
}

func TestFirstServerInstallOverMovedDataTakesASnapshotAndUsesTheDataAsItIs(t *testing.T) {
	fixture := newInstallFixture(t)
	fixture.putDatabaseInPocketBaseData()
	dataBeforeInstall := fingerprintEveryPathUnder(t, fixture.inRoot(pocketBaseDataFolder))
	bundle := fixture.buildServerBundle("first", "1", "1")

	_, calls := fixture.mustRunScript(bundle.scriptPath, everyServerAddressOption()...)

	if snapshotNames := fixture.snapshotNames(); len(snapshotNames) != 1 {
		t.Errorf("expected a snapshot of the moved data, got %v", snapshotNames)
	}
	assertNothingChanged(t, dataBeforeInstall, fingerprintEveryPathUnder(t, fixture.inRoot(pocketBaseDataFolder)))
	assertNoCallStartsWith(t, calls, "chown _skillsurvey:_skillsurvey "+fixture.inRoot(pocketBaseDataFolder))
}

func TestSpecificListenAddressIsUsedByBothServicesAndTheirChecks(t *testing.T) {
	fixture := newInstallFixture(t)
	bundle := fixture.buildServerBundle("first", "1", "1")

	_, calls := fixture.mustRunScript(bundle.scriptPath,
		"--listen-address", documentationOnlyListenAddress,
		"--site-address", exampleSiteAddress,
		"--pocketbase-address", examplePocketBaseAddress,
	)

	if !strings.Contains(fixture.readFileInRoot("etc/httpd.conf"), "listen on "+documentationOnlyListenAddress+" port 80") {
		t.Error("httpd.conf does not listen on the given address")
	}
	if flags := fixture.serviceFlags("pocketbaseserver"); flags != expectedPocketBaseFlags(documentationOnlyListenAddress, exampleSiteAddress) {
		t.Errorf("pocketbaseserver flags are %q", flags)
	}
	for _, checkedAddress := range []string{"http://" + documentationOnlyListenAddress + ":8090/api/health", "http://" + documentationOnlyListenAddress + "/"} {
		if len(callsStartingWith(calls, "ftp "+checkedAddress)) == 0 {
			t.Errorf("expected a check of %s, got %v", checkedAddress, callsStartingWith(calls, "ftp "))
		}
	}
}

func TestServerRunStoppedPartWayIsFinishedByTheNextRun(t *testing.T) {
	fixture := newInstallFixture(t)
	bundle := fixture.buildServerBundle("first", "1", "1")
	fixture.setStubState("fail-start-pocketbaseserver", "")
	output, _ := fixture.mustFailScript(bundle.scriptPath, everyServerAddressOption()...)
	assertOutputContains(t, output, "failed while starting pocketbaseserver")
	fixture.clearStubState("fail-start-pocketbaseserver")

	fixture.mustRunScript(bundle.scriptPath)
	output, calls := fixture.mustRunScript(bundle.scriptPath)

	if !fixture.serviceIsRunning("pocketbaseserver") || !fixture.serviceIsRunning("httpd") {
		t.Error("the second run did not finish starting both services")
	}
	if fixture.readFileInRoot(serverInstallRecord) != bundleInformation(bundle.commit) {
		t.Error("the second run did not record the install")
	}
	assertNoCallStartsWith(t, calls, "rcctl ")
	assertOutputContains(t, output, "0 changes")
}

func TestServerRefusalsChangeNothing(t *testing.T) {
	refusalCases := []struct {
		name                     string
		prepareFixture           func(fixture *installFixture)
		installFirstBundleBefore bool
		bundleToRun              func(fixture *installFixture) serverBundle
		arguments                []string
		expectedReason           string
	}{
		{
			name:           "not root",
			prepareFixture: func(fixture *installFixture) { fixture.setStubState("uid", "1000\n") },
			arguments:      everyServerAddressOption(),
			expectedReason: "run this as root",
		},
		{
			name:           "another OpenBSD release",
			prepareFixture: func(fixture *installFixture) { fixture.setStubState("uname-r", "7.9\n") },
			arguments:      everyServerAddressOption(),
			expectedReason: "7.9",
		},
		{
			name:           "another architecture",
			prepareFixture: func(fixture *installFixture) { fixture.setStubState("uname-m", "arm64\n") },
			arguments:      everyServerAddressOption(),
			expectedReason: "arm64",
		},
		{
			name:           "another system",
			prepareFixture: func(fixture *installFixture) { fixture.setStubState("uname-s", "Linux\n") },
			arguments:      everyServerAddressOption(),
			expectedReason: "Linux",
		},
		{
			name: "a file that does not match SHA256",
			bundleToRun: func(fixture *installFixture) serverBundle {
				bundle := fixture.buildServerBundle("first", "1", "1")
				writeFileCreatingFolders(fixture.t, filepath.Join(bundle.directory, "pocketbaseserver"), "altered after the build\n", 0o755)
				return bundle
			},
			arguments:      everyServerAddressOption(),
			expectedReason: "pocketbaseserver",
		},
		{
			name: "an account with another login class",
			prepareFixture: func(fixture *installFixture) {
				writeFileCreatingFolders(fixture.t, fixture.inRoot("etc/master.passwd"),
					"root:*:0:0:daemon:0:0:Charlie &:/root:/bin/ksh\n_skillsurvey:*:498:498:default:0:0::/var/skillsurvey:/sbin/nologin\n", 0o600)
			},
			arguments:      everyServerAddressOption(),
			expectedReason: "login class default",
		},
		{
			name: "an account in wheel",
			prepareFixture: func(fixture *installFixture) {
				writeFileCreatingFolders(fixture.t, fixture.inRoot("etc/master.passwd"),
					"root:*:0:0:daemon:0:0:Charlie &:/root:/bin/ksh\n_skillsurvey:*:498:498:daemon:0:0::/var/skillsurvey:/sbin/nologin\n", 0o600)
				writeFileCreatingFolders(fixture.t, fixture.inRoot("etc/group"), "wheel:*:0:root,_skillsurvey\n", 0o644)
			},
			arguments:      everyServerAddressOption(),
			expectedReason: "wheel",
		},
		{
			name:           "a first install without the PocketBase address",
			arguments:      []string{"--listen-address", "0.0.0.0", "--site-address", exampleSiteAddress},
			expectedReason: "--pocketbase-address",
		},
		{
			name: "an httpd.conf this script did not write",
			prepareFixture: func(fixture *installFixture) {
				writeFileCreatingFolders(fixture.t, fixture.inRoot("etc/httpd.conf"), "server \"mine\" {\n\tlisten on * port 80\n}\n", 0o644)
			},
			arguments:      everyServerAddressOption(),
			expectedReason: "httpd.conf",
		},
		{
			name:           "an httpd.conf that httpd rejects",
			prepareFixture: func(fixture *installFixture) { fixture.setStubState("fail-httpd-check", "") },
			arguments:      everyServerAddressOption(),
			expectedReason: "httpd -n",
		},
		{
			name: "not enough free space for a snapshot",
			prepareFixture: func(fixture *installFixture) {
				fixture.putDatabaseInPocketBaseData()
				fixture.setStubState("df-available-kilobytes", "1\n")
			},
			installFirstBundleBefore: true,
			bundleToRun:              func(fixture *installFixture) serverBundle { return fixture.buildServerBundle("second", "2", "2") },
			expectedReason:           "free",
		},
		{
			name:           "a listen address that is not an IPv4 address",
			arguments:      []string{"--listen-address", "0.0.0.0 --dev", "--site-address", exampleSiteAddress, "--pocketbase-address", examplePocketBaseAddress},
			expectedReason: "--listen-address",
		},
		{
			name:           "a site address holding a quote",
			arguments:      []string{"--listen-address", "0.0.0.0", "--site-address", exampleSiteAddress + `"`, "--pocketbase-address", examplePocketBaseAddress},
			expectedReason: "--site-address",
		},
		{
			name:           "a PocketBase address holding a second line",
			arguments:      []string{"--listen-address", "0.0.0.0", "--site-address", exampleSiteAddress, "--pocketbase-address", examplePocketBaseAddress + "\nhttp://other.example.test"},
			expectedReason: "--pocketbase-address",
		},
		{
			name:           "an option this script does not know",
			arguments:      append(everyServerAddressOption(), "--dev"),
			expectedReason: "usage",
		},
		{
			name:           "an option without its value",
			arguments:      []string{"--listen-address"},
			expectedReason: "--listen-address needs a value",
		},
	}

	for _, refusalCase := range refusalCases {
		t.Run(refusalCase.name, func(t *testing.T) {
			fixture := newInstallFixture(t)
			if refusalCase.installFirstBundleBefore {
				fixture.mustRunScript(fixture.buildServerBundle("first", "1", "1").scriptPath, everyServerAddressOption()...)
			}
			if refusalCase.prepareFixture != nil {
				refusalCase.prepareFixture(fixture)
			}
			bundle := fixture.buildServerBundle("first", "1", "1")
			if refusalCase.bundleToRun != nil {
				bundle = refusalCase.bundleToRun(fixture)
			}
			before := fixture.fingerprintRoot()

			output, calls := fixture.mustFailScript(bundle.scriptPath, refusalCase.arguments...)

			assertNothingChanged(t, before, fixture.fingerprintRoot())
			assertNoCallStartsWith(t, calls, "rcctl ", "useradd", "chown", "pkg_add")
			assertOutputContains(t, output, refusalCase.expectedReason)
		})
	}
}
