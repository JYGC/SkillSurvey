package main

import (
	"fmt"
	"log"
	"os"

	pocketbaseclient "github.com/r--w/pocketbase"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"keybook/migrate/internal/config"
	"keybook/migrate/internal/migrator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Open legacy SQLite database read-only.
	db, err := gorm.Open(sqlite.Open(cfg.LegacyDbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open legacy db: %v", err)
	}

	// Authenticate with PocketBase.
	pb := pocketbaseclient.NewClient(cfg.PocketBaseUrl,
		pocketbaseclient.WithUserEmailPassword(cfg.ServiceAccountEmail, cfg.ServiceAccountPassword))

	summaries, err := migrator.New(db, pb).Run()
	if err != nil {
		log.Printf("migration error: %v", err)
	}

	// Print summary table.
	fmt.Printf("%-25s %10s %10s\n", "Collection", "Attempted", "Written")
	fmt.Printf("%-25s %10s %10s\n", "----------", "---------", "-------")
	anyFailed := false
	for _, s := range summaries {
		fmt.Printf("%-25s %10d %10d\n", s.Collection, s.Attempted, s.Written)
		if s.Attempted != s.Written {
			anyFailed = true
		}
	}

	if anyFailed {
		os.Exit(1)
	}
}
