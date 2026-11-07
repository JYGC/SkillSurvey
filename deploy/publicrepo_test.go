package deploy_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The repository is public, so nothing under deploy/ may name a host on the home network or hold a private address.
// The patterns are written escaped so this file does not match itself.
var privateHomeNetworkHostnamePattern = regexp.MustCompile(`(?i)\.home\.arpa`)
var privateNetworkAddressPattern = regexp.MustCompile(`(^|[^0-9.])(10\.[0-9]{1,3}|192\.168|172\.(1[6-9]|2[0-9]|3[01]))\.[0-9]{1,3}\.[0-9]{1,3}($|[^0-9])`)

const placeholderForEveryValueOnlyTheHostKnows = "REPLACE_ME"

func findFilesHoldingPrivateHostnamesOrAddresses(t *testing.T, rootDirectory string) []string {
	t.Helper()
	filesFound := []string{}
	err := filepath.WalkDir(rootDirectory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(rootDirectory, path)
		if err != nil {
			return err
		}
		searchableText := relativePath + "\n" + string(content)
		if privateHomeNetworkHostnamePattern.MatchString(searchableText) || privateNetworkAddressPattern.MatchString(searchableText) {
			filesFound = append(filesFound, relativePath)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return filesFound
}

func TestNothingUnderDeployNamesAPrivateHostOrAddress(t *testing.T) {
	if filesFound := findFilesHoldingPrivateHostnamesOrAddresses(t, "."); len(filesFound) > 0 {
		t.Errorf("these files name a private host or address, and the repository is public: %v", filesFound)
	}
}

func TestPrivateHostOrAddressFinderCatchesPlantedValues(t *testing.T) {
	plantedValueByFileName := map[string]string{
		"hostname.txt":         "scp bundle.tgz someone@build" + ".home" + ".arpa:",
		"uppercase.txt":        "BUILD" + ".HOME" + ".ARPA",
		"class-c-address.txt":  "--listen-address " + "192" + ".168" + ".0.99",
		"class-a-address.txt":  "http://" + "10" + ".1.2.3:8090",
		"class-b-address.txt":  "172" + ".20" + ".0.5",
		"address-at-start.txt": "192" + ".168" + ".1.1 is the router",
	}
	valueThatMustNotBeFound := map[string]string{
		"loopback.txt":       "http://127.0.0.1:8090",
		"wildcard.txt":       "--listen-address 0.0.0.0",
		"public-address.txt": "172" + ".32" + ".0.1",
		"version-number.txt": "go1.10.3.4",
	}
	plantedDirectory := t.TempDir()
	for fileName, content := range plantedValueByFileName {
		if err := os.WriteFile(filepath.Join(plantedDirectory, fileName), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for fileName, content := range valueThatMustNotBeFound {
		if err := os.WriteFile(filepath.Join(plantedDirectory, fileName), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	filesFound := map[string]bool{}
	for _, relativePath := range findFilesHoldingPrivateHostnamesOrAddresses(t, plantedDirectory) {
		filesFound[relativePath] = true
	}

	for fileName := range plantedValueByFileName {
		if !filesFound[fileName] {
			t.Errorf("planted private value in %s was not found", fileName)
		}
	}
	for fileName := range valueThatMustNotBeFound {
		if filesFound[fileName] {
			t.Errorf("%s holds no private value, but was reported", fileName)
		}
	}
}

func readJsonExample(t *testing.T, exampleFileName string, decodedExample any) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("examples", exampleFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, decodedExample); err != nil {
		t.Fatalf("examples/%s is not valid JSON: %v", exampleFileName, err)
	}
}

func TestRuntaskSettingsExampleHoldsPlaceholdersForEveryValueOnlyTheHostKnows(t *testing.T) {
	var runtaskSettings map[string]any
	readJsonExample(t, "runtask.json", &runtaskSettings)

	valueEveryHostShares := map[string]any{
		"ServiceAccountEmail":                    "runtask@skillsurvey.com",
		"SeekConfigFile":                         "/etc/runtask/www.seek.com.au.json",
		"JoraConfigFile":                         "/etc/runtask/au.jora.com.json",
		"ErrorLogFile":                           "/var/log/runtask/error.log",
		"SmtpPort":                               float64(587),
		"DynamicContentExtractionTimeoutSeconds": float64(60),
	}
	fieldsOnlyTheHostKnows := []string{
		"PocketBaseUrl", "ServiceAccountPassword", "SmtpDomain", "SenderEmail", "SenderEmailPassword", "EmailRecipient",
	}

	if len(runtaskSettings) != len(valueEveryHostShares)+len(fieldsOnlyTheHostKnows) {
		t.Errorf("examples/runtask.json has %d fields, want exactly the %d this test knows about", len(runtaskSettings), len(valueEveryHostShares)+len(fieldsOnlyTheHostKnows))
	}
	for fieldName, sharedValue := range valueEveryHostShares {
		if runtaskSettings[fieldName] != sharedValue {
			t.Errorf("examples/runtask.json %s is %v, want %v", fieldName, runtaskSettings[fieldName], sharedValue)
		}
	}
	for _, fieldName := range fieldsOnlyTheHostKnows {
		value, isText := runtaskSettings[fieldName].(string)
		if !isText || !strings.Contains(value, placeholderForEveryValueOnlyTheHostKnows) {
			t.Errorf("examples/runtask.json %s is %v, want a value holding %s", fieldName, runtaskSettings[fieldName], placeholderForEveryValueOnlyTheHostKnows)
		}
	}
}

func TestSeekSettingsExampleHoldsPlaceholdersForEverySearchParameter(t *testing.T) {
	var seekSettings struct {
		Pages         int
		ApiParameters []map[string]any
	}
	readJsonExample(t, "www.seek.com.au.json", &seekSettings)

	if len(seekSettings.ApiParameters) == 0 {
		t.Fatal("examples/www.seek.com.au.json has no ApiParameters, so it shows nothing to replace")
	}
	for parameterSetIndex, parameterSet := range seekSettings.ApiParameters {
		for parameterName, value := range parameterSet {
			if _, isNumber := value.(float64); isNumber {
				continue
			}
			if value != placeholderForEveryValueOnlyTheHostKnows {
				t.Errorf("examples/www.seek.com.au.json ApiParameters[%d].%s is %v, want %s", parameterSetIndex, parameterName, value, placeholderForEveryValueOnlyTheHostKnows)
			}
		}
	}
}

func TestJoraSettingsExampleHoldsPlaceholdersForEverySearch(t *testing.T) {
	var joraSettings struct {
		SearchCriterias []struct {
			Name string
			Url  string
		}
	}
	readJsonExample(t, "au.jora.com.json", &joraSettings)

	if len(joraSettings.SearchCriterias) == 0 {
		t.Fatal("examples/au.jora.com.json has no SearchCriterias, so it shows nothing to replace")
	}
	for searchIndex, search := range joraSettings.SearchCriterias {
		if !strings.Contains(search.Name, placeholderForEveryValueOnlyTheHostKnows) || !strings.Contains(search.Url, placeholderForEveryValueOnlyTheHostKnows) {
			t.Errorf("examples/au.jora.com.json SearchCriterias[%d] is %+v, want both fields holding %s", searchIndex, search, placeholderForEveryValueOnlyTheHostKnows)
		}
	}
}
