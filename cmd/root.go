package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/O-Aditya/snippet-snap/config"
	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/tui"
	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	dbPath   string
	database *sql.DB
)

var rootCmd = &cobra.Command{
	Use:   "snap",
	Short: "Snippet-Snap — manage your code snippets from the terminal",
	Long: `Snippet-Snap is a fast CLI tool for saving, searching, and copying
code snippets. Use 'snap add' to save, 'snap find' to search with
fuzzy matching, and 'snap copy' to paste with variable injection.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/snippet-snap/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "database file path")

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printStyledHelp()
	})
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		printStyledHelp()
		return nil
	})
}

// getDB returns the database connection.
func getDB() *sql.DB {
	if database == nil {
		fmt.Fprintln(os.Stderr, "error: database not initialized")
		os.Exit(1)
	}
	return database
}

// ──────────────────────────────────────────────────
// Custom styled help
// ──────────────────────────────────────────────────

func printStyledHelp() {
	cyan := lipgloss.NewStyle().Foreground(tui.ColorCyan)
	cyanBold := lipgloss.NewStyle().Foreground(tui.ColorCyan).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	dim := lipgloss.NewStyle().Foreground(tui.ColorDim)
	bright := lipgloss.NewStyle().Foreground(tui.ColorBright)
	amber := lipgloss.NewStyle().Foreground(tui.ColorAmber)
	green := lipgloss.NewStyle().Foreground(tui.ColorGreen)

	var b strings.Builder

	// 1. ASCII BANNER
	banner := `   ___  _  _ _  _ ___ ___  ___  ___  ___
  / __|| \| | || | _ \ _ \/ _ \|  _||__ \
  \__ \| .` + "`" + ` | || |  _/  _/  __/|  _|  /_/
  |___/|_|\_|\__/|_|  |_|  \___||___|  (_)`

	b.WriteString(cyan.Render(banner) + "\n\n")

	// Version pill + tagline
	versionPill := lipgloss.NewStyle().
		Background(tui.ColorBG3).
		Foreground(tui.ColorMuted).
		Padding(0, 1).
		Render("v1.0.0")
	tagline := muted.Render("Fast CLI for saving, searching, and reusing code snippets")
	b.WriteString("  " + versionPill + "  " + tagline + "\n\n")

	// Section header helper
	sectionHdr := func(name string) string {
		hdr := dim.Bold(true).Render("  " + strings.ToUpper(name))
		line := lipgloss.NewStyle().Foreground(tui.ColorBorder2).Render(strings.Repeat("─", 52))
		return hdr + "\n  " + line
	}

	// 2. COMMANDS
	b.WriteString(sectionHdr("COMMANDS") + "\n")

	type cmdEntry struct {
		name string
		desc string
		star bool
	}
	commands := []cmdEntry{
		{"add", "Save a new snippet interactively", false},
		{"list", "Print all saved snippets as a flat table", false},
		{"find", "Launch fuzzy-search TUI", true},
		{"copy", "Copy snippet to clipboard — fills {{VARS}}", false},
		{"edit", "Open snippet in $EDITOR", false},
		{"rm", "Remove a snippet by ID", false},
	}

	for _, c := range commands {
		name := cyanBold.Width(10).PaddingLeft(4).Render(c.name)
		desc := muted.Render(c.desc)
		line := name + desc
		if c.star {
			line += amber.Render(" ★ recommended")
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")

	// 3. FLAGS
	b.WriteString(sectionHdr("FLAGS") + "\n")

	type flagEntry struct {
		flag string
		typ  string
		desc string
	}
	flags := []flagEntry{
		{"--db", "string", "Path to SQLite database file"},
		{"--config", "string", "Config file (~/.config/snippet-snap)"},
		{"-h", "bool", "Show this help"},
	}

	for _, f := range flags {
		flag := bright.Width(18).PaddingLeft(4).Render(f.flag)
		typ := dim.Width(8).Render(f.typ)
		desc := muted.Render(f.desc)
		b.WriteString(flag + typ + desc + "\n")
	}
	b.WriteString("\n")

	// 4. EXAMPLES
	b.WriteString(sectionHdr("EXAMPLES") + "\n")

	writeEx := func(segments ...string) {
		line := "    " + dim.Render("$ ")
		for i := 0; i+1 < len(segments); i += 2 {
			key := segments[i]
			val := segments[i+1]
			switch key {
			case "c":
				line += cyan.Render(val)
			case "f":
				line += amber.Render(val)
			case "s":
				line += green.Render(val)
			case "a":
				line += bright.Render(val)
			}
		}
		b.WriteString(line + "\n")
	}

	writeEx("c", "snap add", "f", " --name ", "s", "docker-clean", "f", " --lang ", "a", "bash", "f", " --tags ", "s", `"docker,ops"`)
	writeEx("c", "snap find")
	writeEx("c", "snap copy", "a", " 3")
	writeEx("c", "snap list", "f", " --lang ", "a", "bash")
	writeEx("c", "snap rm", "a", " 5")
	b.WriteString("\n")

	// 5. TIPS
	b.WriteString(sectionHdr("TIPS") + "\n")

	tips := []struct {
		icon string
		text string
	}{
		{"★", cyanBold.Render("snap find") + muted.Render("  is the fastest way — no ID memorization needed")},
		{"◈", muted.Render("Use ") + cyanBold.Render("{{VAR}}") + muted.Render(" in content, ") + cyanBold.Render("snap copy") + muted.Render(" will prompt for each value")},
		{"◈", cyanBold.Render("snap [command] --help") + muted.Render("  for per-command flag details")},
	}

	for _, t := range tips {
		icon := amber.Render(t.icon + "  ")
		b.WriteString("    " + icon + t.text + "\n")
	}
	b.WriteString("\n")

	fmt.Print(b.String())
}
