package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/O-Aditya/snippet-snap/config"
	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a snippet in your $EDITOR",
	Long:  `Open a snippet's content in your configured editor. Saves changes on exit.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id %q: %w", args[0], err)
		}

		snippet, err := db.GetSnippetByID(getDB(), id)
		if err != nil {
			return fmt.Errorf("get snippet: %w", err)
		}

		// Write current content to a temp file
		ext := ".txt"
		if snippet.Language != "" {
			ext = "." + snippet.Language
		}
		tmpFile, err := os.CreateTemp("", "snap-edit-*"+ext)
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := tmpFile.WriteString(snippet.Content); err != nil {
			tmpFile.Close()
			return fmt.Errorf("write temp file: %w", err)
		}
		tmpFile.Close()

		// Open editor
		editor := config.Editor()
		editorCmd := exec.Command(editor, tmpPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("editor: %w", err)
		}

		// Read back the edited content
		data, err := os.ReadFile(tmpPath)
		if err != nil {
			return fmt.Errorf("read temp file: %w", err)
		}

		newContent := strings.TrimSpace(string(data))
		if newContent == "" {
			return fmt.Errorf("empty content, edit cancelled")
		}

		if newContent == snippet.Content {
			fmt.Println("No changes detected.")
			return nil
		}

		snippet.Content = newContent
		if err := db.UpdateSnippet(getDB(), snippet); err != nil {
			return fmt.Errorf("update: %w", err)
		}

		fmt.Printf("✓ Snippet %d (%s) updated.\n", snippet.ID, snippet.Alias)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
