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

// rootCmd is the base command when called without any subcommands.
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

// Execute adds all child commands to the root command and sets flags appropriately.
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
	dim := lipgloss.NewStyle().Foreground(tui.ColorDimC)
	text := lipgloss.NewStyle().Foreground(tui.ColorText)
	amber := lipgloss.NewStyle().Foreground(tui.ColorAmber)
	green := lipgloss.NewStyle().Foreground(tui.ColorGreen)

	var b strings.Builder

	// 1. IDENTITY BOX
	logo := cyanBold.Render("◈ SNIPPET-SNAP")
	version := lipgloss.NewStyle().
		Background(tui.ColorBG3).
		Foreground(tui.ColorMuted).
		Padding(0, 1).
		Render("v1.0.0")
	titleRow := logo + "  " + version
	tagline := muted.Render("Fast CLI tool for saving, searching, and reusing code snippets")

	identityBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ColorCyan).
		Width(56).
		Padding(0, 1).
		Render(titleRow + "\n" + tagline)
	b.WriteString(identityBox + "\n\n")

	// 2. COMMANDS section
	sectionHdr := func(name string) string {
		hdr := dim.Bold(true).Render("  " + strings.ToUpper(name))
		line := dim.Render(strings.Repeat("─", 56-len(name)-4))
		return hdr + " " + line
	}

	b.WriteString(sectionHdr("COMMANDS") + "\n")

	type cmdEntry struct {
		name string
		desc string
		star bool
	}
	commands := []cmdEntry{
		{"add", "Save a new snippet interactively", false},
		{"list", "Print all saved snippets as cards or table", false},
		{"find", "Launch fuzzy-search TUI", true},
		{"copy", "Copy snippet to clipboard — fills {{VARS}} interactively", false},
		{"edit", "Open snippet in $EDITOR", false},
		{"rm", "Remove a snippet by ID", false},
	}

	for _, c := range commands {
		name := cyan.Bold(true).Width(10).PaddingLeft(4).Render(c.name)
		desc := muted.Render(c.desc)
		line := name + desc
		if c.star {
			line += amber.Render(" ★ recommended")
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")

	// 3. FLAGS section
	b.WriteString(sectionHdr("FLAGS") + "\n")

	type flagEntry struct {
		flag string
		typ  string
		desc string
	}
	flags := []flagEntry{
		{"--db", "string", "Path to the SQLite database file"},
		{"--config", "string", "Config file (default ~/.config/snippet-snap)"},
		{"-h, --help", "", "Show this help message"},
	}

	for _, f := range flags {
		flag := text.Width(16).PaddingLeft(4).Render(f.flag)
		typ := dim.Width(8).Render(f.typ)
		desc := muted.Render(f.desc)
		b.WriteString(flag + typ + desc + "\n")
	}
	b.WriteString("\n")

	// 4. EXAMPLES section
	b.WriteString(sectionHdr("EXAMPLES") + "\n")

	type exLine struct {
		parts []struct {
			text  string
			style lipgloss.Style
		}
	}

	writeEx := func(segments ...string) {
		// segments alternate: style-key, text, style-key, text...
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
				line += text.Render(val)
			}
		}
		b.WriteString(line + "\n")
	}

	writeEx("c", "snap add", "f", " --name ", "s", "docker-clean", "f", " --lang ", "a", "bash", "f", " --tags ", "s", `"docker,ops"`)
	writeEx("c", "snap find")
	writeEx("c", "snap copy", "a", " 3")
	writeEx("c", "snap list", "f", " --short")
	writeEx("c", "snap rm", "a", " 5")
	b.WriteString("\n")

	// 5. TIPS section
	b.WriteString(sectionHdr("TIPS") + "\n")

	tips := []struct {
		icon string
		text string
	}{
		{"★", cyanBold.Render("snap find") + muted.Render(" is the fastest way to browse and copy snippets")},
		{"◈", muted.Render("Use ") + cyanBold.Render("{{VAR}}") + muted.Render(" placeholders — ") + cyanBold.Render("snap copy") + muted.Render(" prompts you to fill them")},
		{"◈", muted.Render("Run ") + cyanBold.Render("snap [command] --help") + muted.Render(" for flags on any command")},
	}

	for _, t := range tips {
		icon := amber.Render(t.icon)
		b.WriteString("    " + icon + "  " + t.text + "\n")
	}
	b.WriteString("\n")

	fmt.Print(b.String())
}
