package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/highlight"
	"github.com/O-Aditya/snippet-snap/internal/models"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved snippets",
	Long:  `List all snippets with optional language and tag filters. Shows syntax-highlighted content.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lang, _ := cmd.Flags().GetString("lang")
		tag, _ := cmd.Flags().GetString("tag")
		short, _ := cmd.Flags().GetBool("short")

		snippets, err := db.ListSnippets(getDB(), lang, tag)
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}

		if len(snippets) == 0 {
			fmt.Println("No snippets found.")
			return nil
		}

		if short {
			return printShortList(snippets)
		}

		return printDetailedList(snippets)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("lang", "l", "", "filter by language")
	listCmd.Flags().StringP("tag", "t", "", "filter by tag")
	listCmd.Flags().BoolP("short", "s", false, "show compact table (no content)")
}

func printShortList(snippets []models.Snippet) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tALIAS\tLANG\tTAGS\tUPDATED")
	for _, s := range snippets {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			s.ID, s.Alias, s.Language, s.Tags, s.UpdatedAt.Format("2006-01-02 15:04"))
	}
	return w.Flush()
}

func printDetailedList(snippets []models.Snippet) error {
	for i, s := range snippets {
		if i > 0 {
			fmt.Println("─────────────────────────────────────────")
		}
		fmt.Printf("📌 [%d] %s", s.ID, s.Alias)
		if s.Language != "" {
			fmt.Printf("  (%s)", s.Language)
		}
		if s.Tags != "" {
			fmt.Printf("  [%s]", s.Tags)
		}
		fmt.Println()

		rendered, err := highlight.Render(s.Content, s.Language)
		if err != nil {
			// Fall back to plain text
			fmt.Println(s.Content)
		} else {
			fmt.Println(rendered)
		}
	}
	return nil
}
