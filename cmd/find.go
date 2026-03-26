package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/O-Aditya/snippet-snap/internal/clipboard"
	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/inject"
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
			fmt.Println("No snippets yet. Use 'snip add' to create one.")
			return nil
		}

		// Save terminal state BEFORE Bubble Tea takes over.
		// On Windows, the TUI's cancelreader can leave the console
		// in a semi-raw state after exit. This guarantees we can
		// restore cooked mode for var prompts later.
		fd := int(os.Stdin.Fd())
		savedState, stateErr := term.GetState(fd)

		finder := tui.NewFinder(snippets)
		p := tea.NewProgram(finder, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("tui: %w", err)
		}

		// If the user selected a snippet with {{VAR}} placeholders,
		// resolve vars now that the TUI has exited.
		f := finalModel.(tui.Finder)
		if f.SelectedSnippet != nil {
			// Force-restore terminal to cooked mode so stdin reads work.
			if stateErr == nil {
				_ = term.Restore(fd, savedState)
			}

			selected := f.SelectedSnippet
			resolved, err := inject.ResolveVars(selected.Content)
			if err != nil {
				tui.PrintInfo("Aborted.")
				return nil
			}

			if err := clipboard.Copy(resolved); err != nil {
				return fmt.Errorf("copy: clipboard error: %w", err)
			}

			vars := inject.FindVars(selected.Content)
			tui.PrintSuccess(fmt.Sprintf("Copied %s (%d var(s) resolved)", selected.Alias, len(vars)))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(findCmd)
}
