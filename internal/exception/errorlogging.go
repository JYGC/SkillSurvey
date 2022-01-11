package exception

import (
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

func ReportErrorIfPanic(extraData map[string]interface{}) {
	if err := recover(); err != nil {
		ReportError(extraData)
	}
}

func ReportError(extraData map[string]interface{}) {
	errorMap := make(map[string]interface{})
	if extraData != nil {
		errorMap["Extra data"] = extraData
	}
	errorMap["Stacktrace"] = string(debug.Stack())
	ErrorLogger.Println(errorMap)
}
