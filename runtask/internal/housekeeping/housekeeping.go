package housekeeping

import (
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"

	"keybook/runtask/internal/config"
	"keybook/runtask/internal/exception"
)

// CleanFS removes Chromium temporary directories under baseDir (pass "/tmp" in production).
// The glob patterns mirror those used in backend/cmd/housekeeping/main.go.
func CleanFS(baseDir string) error {
	patterns := []string{
		filepath.Join(baseDir, ".org.chromium.Chromium.[a-l][a-lm-qr-vw-zA-Z0-9]*"),
		filepath.Join(baseDir, ".org.chromium.Chromium.[m-q][a-lm-qr-vw-zA-Z0-9]*"),
		filepath.Join(baseDir, ".org.chromium.Chromium.[r-v][a-lm-qr-vw-zA-Z0-9]*"),
		filepath.Join(baseDir, ".org.chromium.Chromium.[w-z][a-lm-qr-vw-zA-Z0-9]*"),
		filepath.Join(baseDir, ".org.chromium.Chromium.[A-Z][a-lm-qr-vw-zA-Z0-9]*"),
		filepath.Join(baseDir, ".org.chromium.Chromium.[0-9][a-lm-qr-vw-zA-Z0-9]*"),
		filepath.Join(baseDir, "chromedp-runner*"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("bad glob pattern %q: %v\n", pattern, err)
			continue
		}
		for _, path := range matches {
			fmt.Printf("removing: %s\n", path)
			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("failed to remove %s: %v\n", path, err)
			}
		}
	}

	fmt.Println("cleanup complete")
	return nil
}

// SendLog reads ErrorLogFile, emails its contents to EmailRecipient, then truncates the file.
func SendLog(cfg config.Config) error {
	allLogs, err := exception.GetAllLogs(cfg.ErrorLogFile)
	if err != nil {
		return fmt.Errorf("read error log: %w", err)
	}

	msg := "From: runtask@localhost\r\n" +
		"To: " + cfg.EmailRecipient + "\r\n" +
		"Subject: SkillSurvey Error Log\r\n\r\n" +
		"*** Error logs ***\n" + allLogs + "\n"

	addr := fmt.Sprintf("%s:%d", cfg.SmtpHost, cfg.SmtpPort)
	if err := smtp.SendMail(addr, nil, "runtask@localhost", []string{cfg.EmailRecipient}, []byte(msg)); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	if err := exception.ClearLogs(cfg.ErrorLogFile); err != nil {
		return fmt.Errorf("clear logs: %w", err)
	}
	return nil
}
