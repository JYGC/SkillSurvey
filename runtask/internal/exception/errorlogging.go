package exception

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

var ErrorLogger *log.Logger

// Init initialises the error logger using the provided file path.
// Called once at startup from main.go using cfg.ErrorLogFile.
func Init(errorLogFilePath string) error {
	file, err := os.OpenFile(
		errorLogFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0666,
	)
	if err != nil {
		return fmt.Errorf("open error log %s: %w", errorLogFilePath, err)
	}
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
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

func GetAllLogs(errorLogFilePath string) (string, error) {
	logContentBytes, err := os.ReadFile(errorLogFilePath)
	if err != nil {
		return "", err
	}
	return string(logContentBytes), nil
}

func ClearLogs(errorLogFilePath string) error {
	return os.Truncate(errorLogFilePath, 0)
}
