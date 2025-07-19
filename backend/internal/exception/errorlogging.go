package exception

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/JYGC/SkillSurvey/internal/environment"
)

const errorLogFileName = "./error.log"

var ErrorLogger *log.Logger

func init() {
	var err error
	file, err := os.OpenFile(
		environment.AttachToExecutableDir(errorLogFileName),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0666,
	)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func ReportErrorIfPanic(extraData map[string]any) (err error) {
	if errInterface := recover(); errInterface != nil {
		LogExtraData(extraData)
		switch x := errInterface.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = fmt.Errorf("%v", x)
		}
		ErrorLogger.Println(err)
	}
	return err
}

func LogExtraData(extraData map[string]any) {
	errorMap := make(map[string]any)
	if extraData != nil {
		errorMap["Extra data"] = extraData
	}
	errorMap["Stacktrace"] = string(debug.Stack())
	ErrorLogger.Println(errorMap)
}

func LogErrorWithLabel(label string, err error) {
	ErrorLogger.Printf("%s: %v\n", label, err)
}

func GetAllLogs() (allLogs string, err error) {
	logContentBytes, err := os.ReadFile(
		environment.AttachToExecutableDir(errorLogFileName),
	)
	if err != nil {
		return "", err
	}
	allLogs = string(logContentBytes)
	return allLogs, nil
}

func ClearLogs() error {
	return os.Truncate(
		environment.AttachToExecutableDir(errorLogFileName),
		0,
	)
}
