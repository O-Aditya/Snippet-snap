package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/tui"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Search snippets with a fuzzy-search TUI",
	Long: `Launch an interactive terminal UI to search, preview, and copy
code snippets. Supports fuzzy matching, syntax highlighting, and
direct clipboard copy.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		snippets, err := db.ListSnippets(getDB(), "", "")
		if err != nil {
			return fmt.Errorf("list snippets: %w", err)
		}

		if len(snippets) == 0 {
			fmt.Println("No snippets yet. Use 'snap add' to create one.")
			return nil
		}

		finder := tui.NewFinder(snippets)
		p := tea.NewProgram(finder, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("tui: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(findCmd)
}
