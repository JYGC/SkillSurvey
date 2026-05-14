package housekeeping_test

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"keybook/runtask/internal/config"
	"keybook/runtask/internal/housekeeping"
)

func TestCleanFSRemovesChromiumTempDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories matching the Chromium temp patterns.
	chromiumDir := filepath.Join(tmpDir, ".org.chromium.Chromium.abcdef")
	chromedpDir := filepath.Join(tmpDir, "chromedp-runner123")
	unrelatedDir := filepath.Join(tmpDir, "unrelated-dir")

	for _, d := range []string{chromiumDir, chromedpDir, unrelatedDir} {
		if err := os.Mkdir(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	if err := housekeeping.CleanFS(tmpDir); err != nil {
		t.Fatalf("CleanFS: %v", err)
	}

	// Chromium dirs must be gone.
	for _, d := range []string{chromiumDir, chromedpDir} {
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed, still exists", d)
		}
	}
	// Unrelated dir must still exist.
	if _, err := os.Stat(unrelatedDir); os.IsNotExist(err) {
		t.Error("unrelated-dir was incorrectly removed by CleanFS")
	}
}

// startSMTPStub listens on a random TCP port and accepts one SMTP conversation.
// It sends the DATA payload (everything between DATA and the terminating dot) to the
// returned channel.
func startSMTPStub(t *testing.T) (port int, received <-chan string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start smtp stub: %v", err)
	}
	port = ln.Addr().(*net.TCPAddr).Port
	ch := make(chan string, 1)

	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		r := bufio.NewReader(conn)
		write := func(s string) { fmt.Fprintf(conn, "%s\r\n", s) }

		write("220 localhost SMTP stub")
		var body strings.Builder
		inData := false

		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\r\n")

			if inData {
				if line == "." {
					ch <- body.String()
					write("250 OK")
					inData = false
					// Allow QUIT to follow.
					continue
				}
				body.WriteString(line)
				body.WriteByte('\n')
				continue
			}

			switch {
			case strings.HasPrefix(line, "EHLO"):
				// Advertise AUTH PLAIN so smtp.PlainAuth is accepted.
				fmt.Fprintf(conn, "250-OK\r\n250 AUTH PLAIN\r\n")
			case strings.HasPrefix(line, "HELO"):
				write("250 OK")
			case strings.HasPrefix(line, "AUTH"):
				write("235 Authentication successful")
			case strings.HasPrefix(line, "MAIL FROM"):
				write("250 OK")
			case strings.HasPrefix(line, "RCPT TO"):
				write("250 OK")
			case line == "DATA":
				write("354 Start input")
				inData = true
				body.Reset()
			case line == "QUIT":
				write("221 Bye")
				return
			default:
				write("500 Unknown")
			}
		}
	}()

	t.Cleanup(func() { ln.Close() })
	return port, ch
}

func TestSendLogEmailsContentsAndTruncates(t *testing.T) {
	logDir := t.TempDir()
	logFile := filepath.Join(logDir, "error.log")
	const logContent = "ERROR: something went wrong\n"
	if err := os.WriteFile(logFile, []byte(logContent), 0644); err != nil {
		t.Fatalf("write error.log: %v", err)
	}

	port, received := startSMTPStub(t)

	cfg := config.Config{
		ErrorLogFile:        logFile,
		SmtpDomain:          "127.0.0.1",
		SmtpPort:            port,
		SenderEmail:         "sender@example.com",
		SenderEmailPassword: "testpassword",
		EmailRecipient:      "admin@example.com",
	}

	if err := housekeeping.SendLog(cfg); err != nil {
		t.Fatalf("SendLog: %v", err)
	}

	// Verify email received with log content.
	select {
	case msg := <-received:
		if !strings.Contains(msg, logContent) {
			t.Errorf("email body missing log content\ngot:\n%s", msg)
		}
	case <-time.After(3 * time.Second):
		t.Error("SMTP stub did not receive a message within 3s")
	}

	// Verify error.log is now zero bytes.
	fi, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("stat error.log: %v", err)
	}
	if fi.Size() != 0 {
		t.Errorf("expected error.log size=0 after SendLog, got %d", fi.Size())
	}
}
