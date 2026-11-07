package deploy_test

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The scripts run against a throwaway root through DESTDIR. Stubs stand in for the commands that need real root,
// or that would reach the network, and log what they were asked to do, so these tests prove the scripts' logic.
// The first install on a real host proves the rest.

const stubThatReportsRoot = `#!/bin/sh
cat "$DEPLOY_TEST_STUB_STATE/uid" 2>/dev/null || echo 0
`

const stubThatReportsThisHostAsOpenBsd80OnAmd64 = `#!/bin/sh
case $1 in
-s) cat "$DEPLOY_TEST_STUB_STATE/uname-s" 2>/dev/null || echo OpenBSD ;;
-r) cat "$DEPLOY_TEST_STUB_STATE/uname-r" 2>/dev/null || echo 8.0 ;;
-m) cat "$DEPLOY_TEST_STUB_STATE/uname-m" 2>/dev/null || echo amd64 ;;
*) echo "uname stub: unexpected arguments $*" >&2; exit 2 ;;
esac
`

const stubThatLogsChown = `#!/bin/sh
echo "chown $*" >> "$DEPLOY_TEST_CALL_LOG"
`

const stubThatAddsTheAccountToTheThrowawayRoot = `#!/bin/sh
echo "useradd $*" >> "$DEPLOY_TEST_CALL_LOG"
while getopts u:g:L:d:s:c: option; do
	case $option in
	u) user_identifier=$OPTARG ;;
	L) login_class=$OPTARG ;;
	d) home_directory=$OPTARG ;;
	s) login_shell=$OPTARG ;;
	c) full_name=$OPTARG ;;
	*) ;;
	esac
done
shift $((OPTIND - 1))
echo "$1:*:$user_identifier:$user_identifier:$login_class:0:0:$full_name:$home_directory:$login_shell" >> "$DESTDIR/etc/master.passwd"
echo "$1:*:$user_identifier:" >> "$DESTDIR/etc/group"
`

// The rcctl stub keeps each service's state in files and logs every action that changes something, with what the
// service would see at that moment, so a test can prove the order of stop, snapshot, replace, set flags and start.
const stubThatKeepsServiceStateLikeRcctl = `#!/bin/sh
state=$DEPLOY_TEST_STUB_STATE/rcctl
mkdir -p "$state"
action=$1
service=${2:-}
pocketbaseserver_program=$DESTDIR/usr/local/libexec/skillsurvey/pocketbaseserver
snapshots_directory=$DESTDIR/var/skillsurvey/snapshots

describe_what_the_service_sees() {
	running=no
	[ -f "$state/$service.running" ] && running=yes
	program=none
	[ -f "$pocketbaseserver_program" ] && program=$(head -2 "$pocketbaseserver_program" | tail -1)
	snapshot_count=$(ls "$snapshots_directory" 2>/dev/null | wc -l | tr -d ' ')
	printf 'running=%s program=%s snapshots=%s flags=%s' "$running" "$program" "$snapshot_count" "$(cat "$state/$service.flags" 2>/dev/null)"
}

case $action in
get)
	case $3 in
	status) [ -f "$state/$service.enabled" ] ;;
	flags)
		if [ -f "$state/$service.enabled" ]; then
			cat "$state/$service.flags" 2>/dev/null || echo
		else
			echo NO
		fi
		;;
	*) echo "rcctl stub: unexpected get $*" >&2; exit 2 ;;
	esac
	;;
check)
	[ -f "$state/$service.running" ]
	;;
enable)
	echo "rcctl enable $service" >> "$DEPLOY_TEST_CALL_LOG"
	touch "$state/$service.enabled"
	;;
set)
	[ "$3" = flags ] || { echo "rcctl stub: unexpected set $*" >&2; exit 2; }
	echo "rcctl set $service flags ($(describe_what_the_service_sees))" >> "$DEPLOY_TEST_CALL_LOG"
	shift 3
	echo "$*" > "$state/$service.flags"
	;;
start | stop | reload)
	echo "rcctl $action $service ($(describe_what_the_service_sees))" >> "$DEPLOY_TEST_CALL_LOG"
	if [ -f "$DEPLOY_TEST_STUB_STATE/fail-$action-$service" ]; then
		exit 1
	fi
	case $action in
	start) touch "$state/$service.running" ;;
	stop) rm -f "$state/$service.running" ;;
	esac
	;;
*)
	echo "rcctl stub: unexpected arguments $*" >&2
	exit 2
	;;
esac
`

const stubThatChecksHttpdConfiguration = `#!/bin/sh
echo "httpd $*" >> "$DEPLOY_TEST_CALL_LOG"
while [ $# -gt 0 ]; do
	[ "$1" = -f ] && cp "$2" "$DEPLOY_TEST_STUB_STATE/last-checked-httpd.conf"
	shift
done
[ ! -f "$DEPLOY_TEST_STUB_STATE/fail-httpd-check" ]
`

// The ftp stub answers like PocketBase for /api/health and like httpd for everything else, serving the file under
// the live release or falling back to index.html, as httpd.conf's "location not found" rule does.
const stubThatAnswersLikePocketBaseAndHttpd = `#!/bin/sh
output_file=-
while [ $# -gt 1 ]; do
	case $1 in
	-o) output_file=$2; shift 2 ;;
	-w) shift 2 ;;
	*) shift ;;
	esac
done
address=$1
echo "ftp $address" >> "$DEPLOY_TEST_CALL_LOG"
path=/${address#http://*/}
[ "$path" = "/$address" ] && path=/

write_answer() {
	if [ "$output_file" = - ]; then cat; else cat > "$output_file"; fi
}

live_release=$DESTDIR/var/www/htdocs/skillsurvey/current
case $path in
/api/health)
	[ -f "$DEPLOY_TEST_STUB_STATE/fail-health" ] && exit 1
	echo '{"message":"API is healthy.","code":200,"data":{}}' | write_answer
	;;
*)
	[ -f "$DEPLOY_TEST_STUB_STATE/fail-site" ] && exit 1
	if [ "$path" != / ] && [ -f "$live_release$path" ]; then
		write_answer < "$live_release$path"
	else
		write_answer < "$live_release/index.html"
	fi
	;;
esac
`

const stubThatReportsFreeSpaceOnVar = `#!/bin/sh
available_kilobytes=$(cat "$DEPLOY_TEST_STUB_STATE/df-available-kilobytes" 2>/dev/null || echo 104857600)
echo "Filesystem 1024-blocks Used Available Capacity Mounted on"
echo "/dev/sd0d 209715200 1024 $available_kilobytes 1% /var"
`

const stubThatReturnsAtOnce = `#!/bin/sh
exit 0
`

const stubThatInstallsPackagesInTheStubState = `#!/bin/sh
echo "pkg_add $*" >> "$DEPLOY_TEST_CALL_LOG"
[ -f "$DEPLOY_TEST_STUB_STATE/fail-pkg_add" ] && exit 1
for package_name in "$@"; do
	case $package_name in -*) continue ;; esac
	touch "$DEPLOY_TEST_STUB_STATE/package-$package_name"
done
`

const stubThatReportsPackagesFromTheStubState = `#!/bin/sh
[ "$1" = -e ] || { echo "pkg_info stub: unexpected arguments $*" >&2; exit 2; }
package_name=${2%%-\**}
[ -f "$DEPLOY_TEST_STUB_STATE/package-$package_name" ]
`

const stubThatRunsCheckSetupWithAnExitCodeTheTestChooses = `#!/bin/sh
echo "su $*" >> "$DEPLOY_TEST_CALL_LOG"
exit_code=$(cat "$DEPLOY_TEST_STUB_STATE/check-setup-exit-code" 2>/dev/null || echo 0)
if [ "$exit_code" -eq 0 ]; then
	echo "ok      seek settings"
else
	echo "FAILED  pocketbase sign-in: stub failure"
fi
exit "$exit_code"
`

func skipUnlessRunningOnOpenBsd(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "openbsd" {
		t.Skip("the install scripts use OpenBSD's sha256, so their tests run on the OpenBSD CI host")
	}
}

type installFixture struct {
	t                  *testing.T
	root               string
	stubDirectory      string
	stubStateDirectory string
	callLog            string
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFileCreatingFolders(t *testing.T, path, content string, permissions os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), permissions); err != nil {
		t.Fatal(err)
	}
}

func newInstallFixture(t *testing.T) *installFixture {
	t.Helper()
	skipUnlessRunningOnOpenBsd(t)
	fixture := &installFixture{
		t:                  t,
		root:               t.TempDir(),
		stubDirectory:      t.TempDir(),
		stubStateDirectory: t.TempDir(),
	}
	fixture.callLog = filepath.Join(fixture.stubStateDirectory, "calls.log")

	writeFileCreatingFolders(t, fixture.inRoot("etc/master.passwd"), "root:*:0:0:daemon:0:0:Charlie &:/root:/bin/ksh\n", 0o600)
	writeFileCreatingFolders(t, fixture.inRoot("etc/group"), "wheel:*:0:root,someone\n", 0o644)
	for _, folderEveryOpenBsdHostHas := range []string{"etc/rc.d", "var/www/htdocs", "var/log", "var/db", "usr/local/libexec"} {
		if err := os.MkdirAll(fixture.inRoot(folderEveryOpenBsdHostHas), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	stubContentByCommandName := map[string]string{
		"id":       stubThatReportsRoot,
		"uname":    stubThatReportsThisHostAsOpenBsd80OnAmd64,
		"chown":    stubThatLogsChown,
		"useradd":  stubThatAddsTheAccountToTheThrowawayRoot,
		"rcctl":    stubThatKeepsServiceStateLikeRcctl,
		"httpd":    stubThatChecksHttpdConfiguration,
		"ftp":      stubThatAnswersLikePocketBaseAndHttpd,
		"df":       stubThatReportsFreeSpaceOnVar,
		"sleep":    stubThatReturnsAtOnce,
		"pkg_add":  stubThatInstallsPackagesInTheStubState,
		"pkg_info": stubThatReportsPackagesFromTheStubState,
		"su":       stubThatRunsCheckSetupWithAnExitCodeTheTestChooses,
	}
	for commandName, stubContent := range stubContentByCommandName {
		writeExecutable(t, filepath.Join(fixture.stubDirectory, commandName), stubContent)
	}
	return fixture
}

func (fixture *installFixture) inRoot(relativePath string) string {
	return filepath.Join(fixture.root, relativePath)
}

func (fixture *installFixture) setStubState(stateFileName, content string) {
	fixture.t.Helper()
	writeFileCreatingFolders(fixture.t, filepath.Join(fixture.stubStateDirectory, stateFileName), content, 0o644)
}

func (fixture *installFixture) clearStubState(stateFileName string) {
	fixture.t.Helper()
	if err := os.Remove(filepath.Join(fixture.stubStateDirectory, stateFileName)); err != nil && !os.IsNotExist(err) {
		fixture.t.Fatal(err)
	}
}

func (fixture *installFixture) serviceIsRunning(serviceName string) bool {
	_, err := os.Stat(filepath.Join(fixture.stubStateDirectory, "rcctl", serviceName+".running"))
	return err == nil
}

func (fixture *installFixture) serviceIsEnabled(serviceName string) bool {
	_, err := os.Stat(filepath.Join(fixture.stubStateDirectory, "rcctl", serviceName+".enabled"))
	return err == nil
}

func (fixture *installFixture) serviceFlags(serviceName string) string {
	content, _ := os.ReadFile(filepath.Join(fixture.stubStateDirectory, "rcctl", serviceName+".flags"))
	return strings.TrimSpace(string(content))
}

// runScript runs a script from a bundle and returns its output and only the calls made during this run.
func (fixture *installFixture) runScript(scriptPath string, arguments ...string) (output string, callsDuringThisRun []string, err error) {
	fixture.t.Helper()
	if err := os.WriteFile(fixture.callLog, nil, 0o644); err != nil {
		fixture.t.Fatal(err)
	}
	command := exec.Command("sh", append([]string{scriptPath}, arguments...)...)
	command.Env = append(os.Environ(),
		"PATH="+fixture.stubDirectory+string(os.PathListSeparator)+os.Getenv("PATH"),
		"DESTDIR="+fixture.root,
		"DEPLOY_TEST_CALL_LOG="+fixture.callLog,
		"DEPLOY_TEST_STUB_STATE="+fixture.stubStateDirectory,
	)
	outputBytes, err := command.CombinedOutput()
	callLogContent, readErr := os.ReadFile(fixture.callLog)
	if readErr != nil {
		fixture.t.Fatal(readErr)
	}
	for _, line := range strings.Split(string(callLogContent), "\n") {
		if line != "" {
			callsDuringThisRun = append(callsDuringThisRun, line)
		}
	}
	return string(outputBytes), callsDuringThisRun, err
}

func (fixture *installFixture) mustRunScript(scriptPath string, arguments ...string) (string, []string) {
	fixture.t.Helper()
	output, calls, err := fixture.runScript(scriptPath, arguments...)
	if err != nil {
		fixture.t.Fatalf("%s failed: %v\n%s", filepath.Base(scriptPath), err, output)
	}
	return output, calls
}

func (fixture *installFixture) mustFailScript(scriptPath string, arguments ...string) (string, []string) {
	fixture.t.Helper()
	output, calls, err := fixture.runScript(scriptPath, arguments...)
	if err == nil {
		fixture.t.Fatalf("%s succeeded, but should have failed:\n%s", filepath.Base(scriptPath), output)
	}
	return output, calls
}

func (fixture *installFixture) readFileInRoot(relativePath string) string {
	fixture.t.Helper()
	content, err := os.ReadFile(fixture.inRoot(relativePath))
	if err != nil {
		fixture.t.Fatal(err)
	}
	return string(content)
}

func fingerprintEveryPathUnder(t *testing.T, directory string) map[string]string {
	t.Helper()
	fingerprintByPath := map[string]string{}
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		information, err := entry.Info()
		if err != nil {
			return err
		}
		fingerprint := fmt.Sprintf("%s %s", information.Mode(), information.ModTime())
		switch {
		case information.Mode()&fs.ModeSymlink != 0:
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			fingerprint += " -> " + linkTarget
		case information.Mode().IsRegular():
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fingerprint += fmt.Sprintf(" %x", sha256.Sum256(content))
		}
		fingerprintByPath[path] = fingerprint
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fingerprintByPath
}

func (fixture *installFixture) fingerprintRoot() map[string]string {
	fixture.t.Helper()
	return fingerprintEveryPathUnder(fixture.t, fixture.root)
}

func assertNothingChanged(t *testing.T, before, after map[string]string) {
	t.Helper()
	for path, fingerprint := range after {
		if before[path] != fingerprint {
			t.Errorf("%s changed:\n before %s\n after  %s", path, before[path], fingerprint)
		}
	}
	for path := range before {
		if _, stillThere := after[path]; !stillThere {
			t.Errorf("%s was removed", path)
		}
	}
}

func callsStartingWith(calls []string, prefix string) []string {
	matchingCalls := []string{}
	for _, call := range calls {
		if strings.HasPrefix(call, prefix) {
			matchingCalls = append(matchingCalls, call)
		}
	}
	return matchingCalls
}

func assertNoCallStartsWith(t *testing.T, calls []string, prefixes ...string) {
	t.Helper()
	for _, prefix := range prefixes {
		if matchingCalls := callsStartingWith(calls, prefix); len(matchingCalls) > 0 {
			t.Errorf("expected no call starting with %q, got %v", prefix, matchingCalls)
		}
	}
}

func assertOutputContains(t *testing.T, output string, expectedTexts ...string) {
	t.Helper()
	for _, expectedText := range expectedTexts {
		if !strings.Contains(output, expectedText) {
			t.Errorf("expected output to contain %q, got:\n%s", expectedText, output)
		}
	}
}

func assertMode(t *testing.T, path string, expectedPermissions os.FileMode) {
	t.Helper()
	information, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if information.Mode().Perm() != expectedPermissions {
		t.Errorf("%s has mode %o, want %o", path, information.Mode().Perm(), expectedPermissions)
	}
}

// writeChecksumFile writes SHA256 the way build-bundles.sh does, with OpenBSD's sha256 over every other file.
func writeChecksumFile(t *testing.T, bundleDirectory string) {
	t.Helper()
	command := exec.Command("sh", "-c", `find . -type f ! -name SHA256 | sed 's|^\./||' | sort | xargs sha256 > SHA256`)
	command.Dir = bundleDirectory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("write SHA256: %v\n%s", err, output)
	}
}

func copyFileFromDeploy(t *testing.T, fileName, destinationPath string, permissions os.FileMode) {
	t.Helper()
	content, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	writeFileCreatingFolders(t, destinationPath, string(content), permissions)
}

func commitNamed(label string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(label)))[:40]
}

func bundleInformation(commit string) string {
	return "commit=" + commit + "\nbuilt_on_release=8.0\nbuilt_on_architecture=amd64\n"
}
