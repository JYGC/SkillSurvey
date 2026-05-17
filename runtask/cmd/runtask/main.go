package main

import (
	"fmt"
	"log"
	"os"

	"keybook/runtask/internal/config"
	"keybook/runtask/internal/exception"
	"keybook/runtask/internal/housekeeping"
	"keybook/runtask/internal/pbclient"
	"keybook/runtask/internal/report"
	"keybook/runtask/internal/scrape"
)

func main() {
	defer func() {
		if err := exception.ReportErrorIfPanic(nil); err != nil {
			log.Printf("panic recovered: %v", err)
			os.Exit(1)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := exception.Init(cfg.ErrorLogFile); err != nil {
		log.Fatalf("init error logger: %v", err)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	// housekeeping sub-commands do not require PocketBase.
	if cmd == "housekeeping" {
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "cleanfs":
			if err := housekeeping.CleanFS("/tmp"); err != nil {
				log.Fatalf("housekeeping cleanfs: %v", err)
			}
		case "sendlog":
			if err := housekeeping.SendLog(cfg); err != nil {
				log.Fatalf("housekeeping sendlog: %v", err)
			}
		default:
			printUsage()
			os.Exit(1)
		}
		return
	}

	// All other commands require PocketBase authentication.
	pb, err := pbclient.New(cfg.PocketBaseUrl, cfg.ServiceAccountEmail, cfg.ServiceAccountPassword)
	if err != nil {
		log.Fatalf("pbclient: %v", err)
	}

	switch cmd {
	case "scrape":
		if err := scrape.Run(cfg, pb); err != nil {
			exception.LogErrorWithLabel("scrape", err)
			log.Fatalf("scrape: %v", err)
		}
	case "report":
		if err := report.Run(cfg, pb); err != nil {
			exception.LogErrorWithLabel("report", err)
			log.Fatalf("report: %v", err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: runtask <command>")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  scrape")
	fmt.Fprintln(os.Stderr, "  report")
	fmt.Fprintln(os.Stderr, "  housekeeping cleanfs")
	fmt.Fprintln(os.Stderr, "  housekeeping sendlog")
}
