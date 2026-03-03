package cmd

import (
	"fmt"
	"strconv"

	"github.com/O-Aditya/snippet-snap/internal/clipboard"
	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/inject"
	"github.com/O-Aditya/snippet-snap/internal/tui"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy <id>",
	Short: "Copy a snippet to the clipboard",
	Long: `Copy a snippet's content to the system clipboard. If the snippet
contains {{VAR}} placeholders, you will be prompted for values.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id %q: %w", args[0], err)
		}

		snippet, err := db.GetSnippetByID(getDB(), id)
		if err != nil {
			return fmt.Errorf("copy: snippet not found: %w", err)
		}

		// ResolveVars handles both cases:
		// — no vars: returns content unchanged immediately
		// — has vars: prompts user on stderr, returns resolved
		resolved, err := inject.ResolveVars(snippet.Content)
		if err != nil {
			tui.PrintInfo("Aborted.")
			return nil
		}

		if err := clipboard.Copy(resolved); err != nil {
			return fmt.Errorf("copy: clipboard error: %w", err)
		}

		vars := inject.FindVars(snippet.Content)
		if len(vars) > 0 {
			tui.PrintSuccess(fmt.Sprintf("Copied %s (%d var(s) resolved)", snippet.Alias, len(vars)))
		} else {
			tui.PrintSuccess("Copied " + snippet.Alias)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
