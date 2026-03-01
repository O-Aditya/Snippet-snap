package cmd

import (
	"fmt"
	"strconv"

	"github.com/O-Aditya/snippet-snap/internal/clipboard"
	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/inject"
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
			return fmt.Errorf("get snippet: %w", err)
		}

		// Resolve any {{VAR}} placeholders
		vars := inject.FindVars(snippet.Content)
		resolved := snippet.Content
		if len(vars) > 0 {
			fmt.Println("This snippet has placeholders. Enter values:")
			var err error
			resolved, err = inject.ResolveVars(snippet.Content)
			if err != nil {
				return fmt.Errorf("resolve vars: %w", err)
			}
		}

		if err := clipboard.Copy(resolved); err != nil {
			return fmt.Errorf("copy: %w", err)
		}

		fmt.Printf("✓ Copied to clipboard (%s)\n", snippet.Alias)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
