package exception

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/readonlysettings"
)

const errorLogFile = "error.log"

var ErrorLogger *log.Logger

func init() {
	configSettings := config.LoadMainConfig()
	var err error
	var appDataFolder string
	appDataFolder, err = readonlysettings.GetAppDataFolder(configSettings.IsProduction)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	file, err := os.OpenFile(
		filepath.Join(appDataFolder, errorLogFile),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0666,
	)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func ReportErrorIfPanicThenPanicAgain(extraData map[string]interface{}) (err error) {
	panic(ReportErrorIfPanic(extraData))
}

func ReportErrorIfPanic(extraData map[string]interface{}) (err error) {
	if errInterface := recover(); errInterface != nil {
		ReportError(extraData)
		switch x := errInterface.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = fmt.Errorf("%v", x)
		}
	}
	return err
}

func ReportError(extraData map[string]interface{}) {
	errorMap := make(map[string]interface{})
	if extraData != nil {
		errorMap["Extra data"] = extraData
	}
	errorMap["Stacktrace"] = string(debug.Stack())
	ErrorLogger.Println(errorMap)
}
