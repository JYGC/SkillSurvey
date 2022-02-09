package exception

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/JYGC/SkillSurvey/internal/config"
)

var ErrorLogger *log.Logger

func init() {
	configSettings := config.LoadMainConfig()
	file, err := os.OpenFile(
		filepath.Join(configSettings.AppDataFolder, configSettings.ErrorLogFile),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0666,
	)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
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
