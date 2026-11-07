package deploy_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const installedRuntaskProgram = "usr/local/libexec/runtask/runtask"
const installedScheduledJobRunner = "usr/local/libexec/runtask/run-scheduled-job"
const runtaskSettingsFolder = "etc/runtask"
const runtaskInstallRecord = "var/db/runtask/BUNDLE-INFO"
const systemCrontab = "etc/crontab"

var everyRuntaskSettingsFileName = []string{"runtask.json", "www.seek.com.au.json", "au.jora.com.json"}

var xLibrariesChromiumNeeds = []string{"libX11", "libxcb", "libXext", "libXrandr", "libfontconfig", "libfreetype", "libdrm", "libgbm"}

const expectedScheduleForSevenAndSeventeen = `# BEGIN runtask schedule. install-runtask.sh replaces everything down to the END line.
MAILTO=root
0 7,17 * * * _runtask -n -s /usr/local/libexec/runtask/run-scheduled-job scrape
0 1 1 * * _runtask -n -s /usr/local/libexec/runtask/run-scheduled-job report
0 19 * * 5 _runtask -n -s /usr/local/libexec/runtask/run-scheduled-job housekeeping cleanfs
0 19 * * 5 _runtask -n -s /usr/local/libexec/runtask/run-scheduled-job housekeeping sendlog
# END runtask schedule
`

type runtaskBundle struct {
	scriptPath     string
	directory      string
	commit         string
	programContent string
}

func (fixture *installFixture) buildRuntaskBundle(commitLabel, programVersion string) runtaskBundle {
	fixture.t.Helper()
	commit := commitNamed(commitLabel)
	bundleDirectory := filepath.Join(fixture.t.TempDir(), "skillsurvey-runtask-"+commit[:7])
	programContent := "fake runtask\nruntask-version-" + programVersion + "\n"

	writeFileCreatingFolders(fixture.t, filepath.Join(bundleDirectory, "BUNDLE-INFO"), bundleInformation(commit), 0o644)
	copyFileFromDeploy(fixture.t, "install-runtask.sh", filepath.Join(bundleDirectory, "install-runtask.sh"), 0o644)
	copyFileFromDeploy(fixture.t, "install-common.sh", filepath.Join(bundleDirectory, "install-common.sh"), 0o644)
	copyFileFromDeploy(fixture.t, "run-scheduled-job", filepath.Join(bundleDirectory, "run-scheduled-job"), 0o755)
	for _, settingsFileName := range everyRuntaskSettingsFileName {
		copyFileFromDeploy(fixture.t, filepath.Join("examples", settingsFileName), filepath.Join(bundleDirectory, "examples", settingsFileName), 0o644)
	}
	writeFileCreatingFolders(fixture.t, filepath.Join(bundleDirectory, "runtask"), programContent, 0o755)
	writeChecksumFile(fixture.t, bundleDirectory)

	return runtaskBundle{
		scriptPath:     filepath.Join(bundleDirectory, "install-runtask.sh"),
		directory:      bundleDirectory,
		commit:         commit,
		programContent: programContent,
	}
}

func (fixture *installFixture) installXLibrariesAndTimeZone() {
	fixture.t.Helper()
	for _, libraryName := range xLibrariesChromiumNeeds {
		writeFileCreatingFolders(fixture.t, fixture.inRoot("usr/X11R6/lib/"+libraryName+".so.1.0"), "", 0o644)
	}
	if err := os.Symlink("/usr/share/zoneinfo/Australia/Melbourne", fixture.inRoot("etc/localtime")); err != nil {
		fixture.t.Fatal(err)
	}
}

func newRuntaskHostFixture(t *testing.T) *installFixture {
	t.Helper()
	fixture := newInstallFixture(t)
	fixture.installXLibrariesAndTimeZone()
	return fixture
}

func (fixture *installFixture) replaceEveryPlaceholderInTheSettings() {
	fixture.t.Helper()
	for _, settingsFileName := range everyRuntaskSettingsFileName {
		settingsPath := fixture.inRoot(runtaskSettingsFolder + "/" + settingsFileName)
		settings := strings.ReplaceAll(fixture.readFileInRoot(runtaskSettingsFolder+"/"+settingsFileName), "REPLACE_ME", "real-value-"+settingsFileName)
		writeFileCreatingFolders(fixture.t, settingsPath, settings, 0o640)
	}
}

func scheduleArguments(scrapeHours string) []string {
	return []string{"--start-schedule", "--scrape-hours", scrapeHours}
}

func TestFirstRuntaskInstallSetsUpTheAccountChromiumProgramAndSettingsButNoSchedule(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	bundle := fixture.buildRuntaskBundle("first", "1")

	output, calls := fixture.mustRunScript(bundle.scriptPath)

	useraddCalls := callsStartingWith(calls, "useradd ")
	if len(useraddCalls) != 1 {
		t.Fatalf("expected one useradd call, got %v", useraddCalls)
	}
	for _, expectedArgument := range []string{"-u 497", "-L runtask", "-d /var/runtask", "-s /sbin/nologin", "_runtask"} {
		if !strings.Contains(useraddCalls[0], expectedArgument) {
			t.Errorf("useradd call %q lacks %q", useraddCalls[0], expectedArgument)
		}
	}
	if loginClass := fixture.readFileInRoot("etc/login.conf.d/runtask"); loginClass != "runtask:datasize=4096M:coredumpsize=0:maxproc=256:tc=default:\n" {
		t.Errorf("login class is %q", loginClass)
	}
	if pkgAddCalls := callsStartingWith(calls, "pkg_add "); len(pkgAddCalls) != 1 || !strings.Contains(pkgAddCalls[0], "chromium") {
		t.Errorf("expected one pkg_add of chromium, got %v", pkgAddCalls)
	}
	if fixture.readFileInRoot(installedRuntaskProgram) != bundle.programContent {
		t.Error("installed runtask differs from the bundle's")
	}
	assertMode(t, fixture.inRoot(installedRuntaskProgram), 0o755)
	jobRunner, _ := os.ReadFile("run-scheduled-job")
	if fixture.readFileInRoot(installedScheduledJobRunner) != string(jobRunner) {
		t.Error("installed run-scheduled-job differs from deploy/run-scheduled-job")
	}
	assertMode(t, fixture.inRoot(installedScheduledJobRunner), 0o755)
	if linkTarget, err := os.Readlink(fixture.inRoot("usr/local/libexec/runtask/runtask.json")); err != nil || linkTarget != "/etc/runtask/runtask.json" {
		t.Errorf("runtask.json beside the program points at %q (%v), want /etc/runtask/runtask.json", linkTarget, err)
	}
	for _, settingsFileName := range everyRuntaskSettingsFileName {
		example, _ := os.ReadFile(filepath.Join("examples", settingsFileName))
		if fixture.readFileInRoot(runtaskSettingsFolder+"/"+settingsFileName) != string(example) {
			t.Errorf("%s was not created from its example", settingsFileName)
		}
		assertMode(t, fixture.inRoot(runtaskSettingsFolder+"/"+settingsFileName), 0o640)
		if len(callsStartingWith(calls, "chown root:_runtask "+fixture.inRoot(runtaskSettingsFolder)+"/")) == 0 {
			t.Errorf("no settings file was given to root:_runtask, got %v", callsStartingWith(calls, "chown "))
		}
	}
	assertMode(t, fixture.inRoot(runtaskSettingsFolder), 0o750)
	assertMode(t, fixture.inRoot("var/runtask"), 0o700)
	assertMode(t, fixture.inRoot("var/log/runtask"), 0o750)
	if fixture.readFileInRoot(runtaskInstallRecord) != bundleInformation(bundle.commit) {
		t.Error("the install record differs from the bundle's BUNDLE-INFO")
	}
	if _, err := os.Stat(fixture.inRoot(systemCrontab)); !os.IsNotExist(err) {
		t.Error("a first install without --start-schedule wrote /etc/crontab")
	}
	assertNoCallStartsWith(t, calls, "su ")
	assertOutputContains(t, output, "/etc/runtask/runtask.json", "--start-schedule --scrape-hours")
}

func TestSecondRuntaskRunWithTheSameBundleChangesNothing(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	bundle := fixture.buildRuntaskBundle("first", "1")
	fixture.mustRunScript(bundle.scriptPath)
	before := fixture.fingerprintRoot()

	output, calls := fixture.mustRunScript(bundle.scriptPath)

	assertNothingChanged(t, before, fixture.fingerprintRoot())
	assertNoCallStartsWith(t, calls, "pkg_add", "useradd", "chown", "su ")
	assertOutputContains(t, output, "0 changes")
}

func TestRuntaskStartsTheScheduleOnlyAfterCheckSetupPassesAndPrintsTheTimeZone(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	bundle := fixture.buildRuntaskBundle("first", "1")
	fixture.mustRunScript(bundle.scriptPath)
	fixture.replaceEveryPlaceholderInTheSettings()

	output, calls := fixture.mustRunScript(bundle.scriptPath, scheduleArguments("7,17")...)

	if crontab := fixture.readFileInRoot(systemCrontab); crontab != expectedScheduleForSevenAndSeventeen {
		t.Errorf("crontab is:\n%s\nwant:\n%s", crontab, expectedScheduleForSevenAndSeventeen)
	}
	assertMode(t, fixture.inRoot(systemCrontab), 0o644)
	suCalls := callsStartingWith(calls, "su ")
	if len(suCalls) != 1 || !strings.Contains(suCalls[0], "-c runtask") || !strings.Contains(suCalls[0], "_runtask") || !strings.Contains(suCalls[0], "/usr/local/libexec/runtask/run-scheduled-job check-setup") {
		t.Errorf("expected check-setup as _runtask under its login class, got %v", suCalls)
	}
	assertOutputContains(t, output, "Australia/Melbourne")
}

func TestRuntaskScheduleWithTheSameHoursChangesNothing(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	bundle := fixture.buildRuntaskBundle("first", "1")
	fixture.mustRunScript(bundle.scriptPath)
	fixture.replaceEveryPlaceholderInTheSettings()
	fixture.mustRunScript(bundle.scriptPath, scheduleArguments("7,17")...)
	before := fixture.fingerprintRoot()

	output, _ := fixture.mustRunScript(bundle.scriptPath, scheduleArguments("7,17")...)

	assertNothingChanged(t, before, fixture.fingerprintRoot())
	assertOutputContains(t, output, "0 changes")
}

func TestRuntaskScheduleWithNewHoursReplacesTheBlockAndLeavesOtherCrontabLinesAlone(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	linesBeforeTheBlock := "SHELL=/bin/sh\n0 4 * * * root /usr/libexec/someone-elses-job\n"
	writeFileCreatingFolders(t, fixture.inRoot(systemCrontab), linesBeforeTheBlock, 0o644)
	bundle := fixture.buildRuntaskBundle("first", "1")
	fixture.mustRunScript(bundle.scriptPath)
	fixture.replaceEveryPlaceholderInTheSettings()
	fixture.mustRunScript(bundle.scriptPath, scheduleArguments("7,17")...)
	lineAfterTheBlock := "30 5 * * * root /usr/libexec/a-job-added-later\n"
	crontabWithLaterLine := fixture.readFileInRoot(systemCrontab) + lineAfterTheBlock
	writeFileCreatingFolders(t, fixture.inRoot(systemCrontab), crontabWithLaterLine, 0o644)

	fixture.mustRunScript(bundle.scriptPath, scheduleArguments("3,12")...)

	expectedSchedule := strings.Replace(expectedScheduleForSevenAndSeventeen, "0 7,17 * * *", "0 3,12 * * *", 1)
	expectedCrontab := linesBeforeTheBlock + expectedSchedule + lineAfterTheBlock
	if crontab := fixture.readFileInRoot(systemCrontab); crontab != expectedCrontab {
		t.Errorf("crontab is:\n%s\nwant:\n%s", crontab, expectedCrontab)
	}
}

func TestRuntaskUpgradeWithoutTheScheduleOptionLeavesTheScheduleAndSettingsAsTheyWere(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	firstBundle := fixture.buildRuntaskBundle("first", "1")
	fixture.mustRunScript(firstBundle.scriptPath)
	fixture.replaceEveryPlaceholderInTheSettings()
	fixture.mustRunScript(firstBundle.scriptPath, scheduleArguments("7,17")...)
	crontabBefore := fixture.readFileInRoot(systemCrontab)
	settingsBefore := fingerprintEveryPathUnder(t, fixture.inRoot(runtaskSettingsFolder))
	secondBundle := fixture.buildRuntaskBundle("second", "2")

	_, calls := fixture.mustRunScript(secondBundle.scriptPath)

	if fixture.readFileInRoot(systemCrontab) != crontabBefore {
		t.Error("an upgrade without --start-schedule changed the schedule")
	}
	assertNothingChanged(t, settingsBefore, fingerprintEveryPathUnder(t, fixture.inRoot(runtaskSettingsFolder)))
	if fixture.readFileInRoot(installedRuntaskProgram) != secondBundle.programContent {
		t.Error("runtask was not replaced")
	}
	if fixture.readFileInRoot(runtaskInstallRecord) != bundleInformation(secondBundle.commit) {
		t.Error("the install record still names the first bundle")
	}
	assertNoCallStartsWith(t, calls, "su ", "pkg_add")
}

func TestRuntaskKeepsSettingsThatAlreadyExistOnEveryRun(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	settingsWrittenBeforeTheFirstInstall := "{\"written\": \"by the operator before the first install\"}\n"
	writeFileCreatingFolders(t, fixture.inRoot(runtaskSettingsFolder+"/runtask.json"), settingsWrittenBeforeTheFirstInstall, 0o640)

	fixture.mustRunScript(fixture.buildRuntaskBundle("first", "1").scriptPath)
	fixture.mustRunScript(fixture.buildRuntaskBundle("second", "2").scriptPath)

	if fixture.readFileInRoot(runtaskSettingsFolder+"/runtask.json") != settingsWrittenBeforeTheFirstInstall {
		t.Error("runtask.json was overwritten")
	}
}

func TestRuntaskRefusesTheScheduleWhileAnySettingsFileHoldsAPlaceholder(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	bundle := fixture.buildRuntaskBundle("first", "1")
	fixture.mustRunScript(bundle.scriptPath)
	fixture.replaceEveryPlaceholderInTheSettings()
	writeFileCreatingFolders(t, fixture.inRoot(runtaskSettingsFolder+"/au.jora.com.json"), "{\"Name\": \"REPLACE_ME\"}\n", 0o640)

	output, calls := fixture.mustFailScript(bundle.scriptPath, scheduleArguments("7,17")...)

	assertOutputContains(t, output, "/etc/runtask/au.jora.com.json")
	assertNoCallStartsWith(t, calls, "su ")
	if _, err := os.Stat(fixture.inRoot(systemCrontab)); !os.IsNotExist(err) {
		t.Error("a schedule was written although a placeholder was left")
	}
}

func TestRuntaskRefusesTheScheduleWhenCheckSetupFailsAndLeavesTheExistingScheduleAlone(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	bundle := fixture.buildRuntaskBundle("first", "1")
	fixture.mustRunScript(bundle.scriptPath)
	fixture.replaceEveryPlaceholderInTheSettings()
	fixture.mustRunScript(bundle.scriptPath, scheduleArguments("7,17")...)
	crontabBefore := fixture.readFileInRoot(systemCrontab)
	fixture.setStubState("check-setup-exit-code", "1\n")

	output, _ := fixture.mustFailScript(bundle.scriptPath, scheduleArguments("3,12")...)

	assertOutputContains(t, output, "check-setup", "pocketbase sign-in")
	if fixture.readFileInRoot(systemCrontab) != crontabBefore {
		t.Error("the schedule changed although check-setup failed")
	}
}

func TestRuntaskLeavesNoScheduleWhenChromiumCannotBeInstalledAndTheNextRunFinishes(t *testing.T) {
	fixture := newRuntaskHostFixture(t)
	bundle := fixture.buildRuntaskBundle("first", "1")
	fixture.setStubState("fail-pkg_add", "")

	output, calls := fixture.mustFailScript(bundle.scriptPath, scheduleArguments("7,17")...)

	assertOutputContains(t, output, "chromium")
	assertNoCallStartsWith(t, calls, "su ")
	if _, err := os.Stat(fixture.inRoot(systemCrontab)); !os.IsNotExist(err) {
		t.Error("a schedule was written although chromium could not be installed")
	}

	fixture.clearStubState("fail-pkg_add")
	fixture.mustRunScript(bundle.scriptPath)
	output, _ = fixture.mustRunScript(bundle.scriptPath)

	if fixture.readFileInRoot(runtaskInstallRecord) != bundleInformation(bundle.commit) {
		t.Error("the next run did not finish the install")
	}
	assertOutputContains(t, output, "0 changes")
}

func TestRuntaskRefusalsChangeNothing(t *testing.T) {
	refusalCases := []struct {
		name           string
		prepareFixture func(fixture *installFixture)
		bundleToRun    func(fixture *installFixture) runtaskBundle
		arguments      []string
		expectedReason string
	}{
		{
			name:           "not root",
			prepareFixture: func(fixture *installFixture) { fixture.setStubState("uid", "1000\n") },
			expectedReason: "run this as root",
		},
		{
			name:           "another OpenBSD release",
			prepareFixture: func(fixture *installFixture) { fixture.setStubState("uname-r", "7.9\n") },
			expectedReason: "7.9",
		},
		{
			name: "a file that does not match SHA256",
			bundleToRun: func(fixture *installFixture) runtaskBundle {
				bundle := fixture.buildRuntaskBundle("first", "1")
				writeFileCreatingFolders(fixture.t, filepath.Join(bundle.directory, "examples", "runtask.json"), "{}\n", 0o644)
				return bundle
			},
			expectedReason: "examples/runtask.json",
		},
		{
			name: "an account with another home",
			prepareFixture: func(fixture *installFixture) {
				writeFileCreatingFolders(fixture.t, fixture.inRoot("etc/master.passwd"),
					"root:*:0:0:daemon:0:0:Charlie &:/root:/bin/ksh\n_runtask:*:497:497:runtask:0:0::/home/runtask:/sbin/nologin\n", 0o600)
			},
			expectedReason: "home /home/runtask",
		},
		{
			name: "missing X libraries",
			prepareFixture: func(fixture *installFixture) {
				if err := os.Remove(fixture.inRoot("usr/X11R6/lib/libX11.so.1.0")); err != nil {
					fixture.t.Fatal(err)
				}
			},
			expectedReason: "libX11",
		},
		{
			name:           "a scrape hour that is not an hour",
			arguments:      scheduleArguments("7,25"),
			expectedReason: "--scrape-hours",
		},
		{
			name:           "scrape hours holding a second command",
			arguments:      scheduleArguments("7;17"),
			expectedReason: "--scrape-hours",
		},
		{
			name:           "the schedule option without hours",
			arguments:      []string{"--start-schedule"},
			expectedReason: "--scrape-hours",
		},
		{
			name:           "hours without the schedule option",
			arguments:      []string{"--scrape-hours", "7,17"},
			expectedReason: "--start-schedule",
		},
		{
			name:           "an option this script does not know",
			arguments:      []string{"--listen-address", "0.0.0.0"},
			expectedReason: "usage",
		},
	}

	for _, refusalCase := range refusalCases {
		t.Run(refusalCase.name, func(t *testing.T) {
			fixture := newRuntaskHostFixture(t)
			if refusalCase.prepareFixture != nil {
				refusalCase.prepareFixture(fixture)
			}
			bundle := fixture.buildRuntaskBundle("first", "1")
			if refusalCase.bundleToRun != nil {
				bundle = refusalCase.bundleToRun(fixture)
			}
			before := fixture.fingerprintRoot()

			output, calls := fixture.mustFailScript(bundle.scriptPath, refusalCase.arguments...)

			assertNothingChanged(t, before, fixture.fingerprintRoot())
			assertNoCallStartsWith(t, calls, "useradd", "chown", "pkg_add", "su ")
			assertOutputContains(t, output, refusalCase.expectedReason)
		})
	}
}
