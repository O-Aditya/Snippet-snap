package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove a snippet by ID",
	Long:  `Permanently delete a snippet by its numeric ID. Prompts for confirmation.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id %q: %w", args[0], err)
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			// Show the snippet to be deleted
			snippet, err := db.GetSnippetByID(getDB(), id)
			if err != nil {
				return fmt.Errorf("get snippet: %w", err)
			}
			fmt.Printf("Delete snippet %d (%s)? [y/N]: ", snippet.ID, snippet.Alias)

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := db.DeleteSnippet(getDB(), id); err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		fmt.Printf("✓ Snippet %d deleted.\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}
