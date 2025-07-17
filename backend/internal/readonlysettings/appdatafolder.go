package readonlysettings

import (
	"os"
	"path/filepath"
)

const (
	appDataFolder    = "SkillSurvey"
	devAppDataFolder = "SkillSurveyDev"
)

func GetAppDataFolder(isProduction bool) (userConfigDir string, err error) {
	var chosenAppDataFolder string
	if isProduction {
		chosenAppDataFolder = appDataFolder
	} else {
		chosenAppDataFolder = devAppDataFolder
	}
	userConfigDir, err = os.UserConfigDir()
	return filepath.Join(userConfigDir, chosenAppDataFolder), err
}
