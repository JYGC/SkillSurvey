package main

import (
	"log"
	"net/smtp"
	"os"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/environment"
	"github.com/JYGC/SkillSurvey/internal/exception"
)

func main() {
	createAndSendAdminReport()
	if err := exception.ClearLogs(); err != nil {
		panic(err)
	}
}

const mailAdminConfigFileName = "./mailadmin.json"

func createAndSendAdminReport() {
	mailAdminConfig, err := getMailAdminConfig()
	if err != nil {
		panic(err)
	}
	allLogs, err := exception.GetAllLogs()
	if err != nil {
		panic(err)
	}
	errorLogsReport := "*** Errorlogs ***\n" +
		allLogs + "\n"
	body := errorLogsReport
	sendEmailToAdmin(body, mailAdminConfig)
}

func getMailAdminConfig() (mailAdminConfig MailAdminConfig, err error) {
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
		log.Printf("smtp error: %s", err)
		return
	}
	log.Print("sent, visit http://foobarbazz.mailinator.com")
}

type MailAdminConfig struct {
	SenderEmail         string
	AdminEmail          string
	AppName             string
	SenderEmailPassword string
}
