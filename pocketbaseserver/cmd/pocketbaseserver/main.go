package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/spf13/cobra"

	"keybook/pocketbaseserver/internal/altmigrate"
	_ "keybook/pocketbaseserver/migrations"
)

// isProbablyGoRun reports whether the binary was started with `go run`.
// `go run` compiles to a temp directory, so the executable path contains the OS temp dir.
func isProbablyGoRun() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.Contains(execPath, os.TempDir())
}

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// Auto-creates migration files when collections change via the Dashboard,
		// but only when running via `go run` (not the compiled binary).
		Automigrate: isProbablyGoRun(),
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))
		return se.Next()
	})

	altMigrateCmd := &cobra.Command{
		Use:   "alt-migrate",
		Short: "Migrate jobPosts from legacy SQLite directly into PocketBase",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, _ := cmd.Flags().GetString("db")
			if dbPath == "" {
				return errors.New("--db flag is required")
			}
			if err := app.Bootstrap(); err != nil {
				return err
			}
			summary, err := altmigrate.Run(app, dbPath)
			fmt.Printf("jobPosts:   attempted=%d  written=%d  skipped=%d  failed=%d\n",
				summary.Attempted, summary.Written, summary.Skipped, summary.Failed)
			if err != nil {
				return err
			}
			if summary.Failed > 0 {
				return errors.New("migration completed with failures — see log for details")
			}
			return nil
		},
	}
	altMigrateCmd.Flags().String("db", "", "Path to legacy SkillSurvey.db SQLite file")
	app.RootCmd.AddCommand(altMigrateCmd)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
