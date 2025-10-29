package main

import (
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/environment"
	"github.com/JYGC/SkillSurvey/internal/exception"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("need at leat one argument")
		return
	}
	for _, arg := range os.Args[1:] {
		switch arg {
		case "cleanfs":
			cleanupFilesystem()
		case "sendlog":
			sendLogToAdmin()
		case "sendcleanlog":
			sendLogToAdminAndClearLog()
		case "alertsendlog":
			sendLogToAdminIfErrorsOverLimit()
		default:
			fmt.Println("unknown command")
		}
	}
}

func sendLogToAdminAndClearLog() {
	if err := sendLogToAdmin(); err != nil {
		fmt.Printf("sendLogToAdmin error: %s\n", err)
		return
	}
	if err := exception.ClearLogs(); err != nil {
		fmt.Printf("exception.ClearLogs error: %s\n", err)
	}
}

func sendLogToAdminIfErrorsOverLimit() {
	const numberOfErrorLimit = 1000
	errorPattern := "(?m)^ERROR:"
	re := regexp.MustCompile(errorPattern)
	allLogs, err := exception.GetAllLogs()
	if err != nil {
		fmt.Printf("exception.GetAllLogs error: %s\n", err)
		return
	}
	matches := re.FindAllString(allLogs, -1)
	numberOfErrors := len(matches)
	fmt.Printf("number of errors: %d\n", numberOfErrors)
	if numberOfErrors >= numberOfErrorLimit {
		sendLogToAdmin()
	}
}

func sendLogToAdmin() error {
	mailAdminConfig, err := getMailAdminConfig()
	if err != nil {
		fmt.Printf("getMailAdminConfig error: %s\n", err)
		return err
	}
	allLogs, err := exception.GetAllLogs()
	if err != nil {
		fmt.Printf("exception.GetAllLogs error: %s\n", err)
		return err
	}
	errorLogsReport := "*** Errorlogs ***\n" +
		allLogs + "\n"
	body := errorLogsReport
	sendEmailToAdmin(body, mailAdminConfig)
	return nil
}

func getMailAdminConfig() (mailAdminConfig MailAdminConfig, err error) {
	const mailAdminConfigFileName = "./mailadmin.json"
	mailAdminConfigPath :=
		environment.AttachToExecutableDir(mailAdminConfigFileName)
	if _, err := os.Stat(mailAdminConfigPath); err != nil {
		return MailAdminConfig{}, err
	}
	config.JsonToConfig(&mailAdminConfig, mailAdminConfigPath)
	return mailAdminConfig, nil
}

func sendEmailToAdmin(body string, mailAdminConfig MailAdminConfig) {
	msg := "From: " + mailAdminConfig.SenderEmail + "\n" +
		"To: " + mailAdminConfig.AdminEmail + "\n" +
		"Subject: SkillSurvey Reports\n\n" +
		body

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth(
			"",
			mailAdminConfig.SenderEmail,
			mailAdminConfig.SenderEmailPassword,
			"smtp.gmail.com",
		),
		mailAdminConfig.SenderEmail,
		[]string{mailAdminConfig.AdminEmail},
		[]byte(msg),
	)

	if err != nil {
		fmt.Printf("smtp error: %s\n", err)
		return
	}
	fmt.Print("sent, visit http://foobarbazz.mailinator.com")
}

type MailAdminConfig struct {
	SenderEmail         string
	AdminEmail          string
	AppName             string
	SenderEmailPassword string
}

func cleanupFilesystem() {
	filepathPatterns := []string{
		"/tmp/.org.chromium.Chromium.[a-l][a-lm-qr-vw-zA-Z0-9]*",
		"/tmp/.org.chromium.Chromium.[m-q][a-lm-qr-vw-zA-Z0-9]*",
		"/tmp/.org.chromium.Chromium.[r-v][a-lm-qr-vw-zA-Z0-9]*",
		"/tmp/.org.chromium.Chromium.[w-z][a-lm-qr-vw-zA-Z0-9]*",
		"/tmp/.org.chromium.Chromium.[A-Z][a-lm-qr-vw-zA-Z0-9]*",
		"/tmp/.org.chromium.Chromium.[0-9][a-lm-qr-vw-zA-Z0-9]*",
		"/tmp/chromedp-runner*",
	}

	for _, pattern := range filepathPatterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Bad glob pattern %q: %v\n", pattern, err)
			continue
		}

		for _, path := range matches {
			fmt.Printf("Removing: %s\n", path)
			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("Failed to remove %s: %v\n", path, err)
			}
		}
	}

	fmt.Println("Cleanup complete.")
}
