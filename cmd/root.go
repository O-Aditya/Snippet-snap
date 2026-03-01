package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/O-Aditya/snippet-snap/config"
	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dbPath  string
	database *sql.DB
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "snap",
	Short: "Snippet-Snap — manage your code snippets from the terminal",
	Long: `Snippet-Snap is a fast CLI tool for saving, searching, and copying
code snippets. Use 'snap add' to save, 'snap find' to search with
fuzzy matching, and 'snap copy' to paste with variable injection.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip DB init for help/completion commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		if err := config.Load(cfgFile); err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		path := dbPath
		if path == "" {
			path = config.DBPath()
		}

		var err error
		database, err = db.Open(path)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		if err := db.RunMigrations(database); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if database != nil {
			database.Close()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/snippet-snap/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "database file path")
}

// getDB returns the database connection, falling back to opening it if unset.
// This is used by subcommands.
func getDB() *sql.DB {
	if database == nil {
		fmt.Fprintln(os.Stderr, "error: database not initialized")
		os.Exit(1)
	}
	return database
}
