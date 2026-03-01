package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/O-Aditya/snippet-snap/config"
	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/models"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new code snippet",
	Long:  `Add a new code snippet interactively. Content is read from $EDITOR or stdin.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		lang, _ := cmd.Flags().GetString("lang")
		tags, _ := cmd.Flags().GetString("tags")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		content, err := getContent()
		if err != nil {
			return fmt.Errorf("get content: %w", err)
		}

		snippet := &models.Snippet{
			Alias:    name,
			Content:  content,
			Language: lang,
			Tags:     tags,
		}

		if err := snippet.Validate(); err != nil {
			return fmt.Errorf("validate: %w", err)
		}

		id, err := db.InsertSnippet(getDB(), snippet)
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}

		fmt.Printf("✓ Snippet saved (id: %d, alias: %s)\n", id, name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("name", "n", "", "unique alias for the snippet (required)")
	addCmd.Flags().StringP("lang", "l", "", "language for syntax highlighting")
	addCmd.Flags().StringP("tags", "t", "", "comma-separated tags")
}

// getContent reads snippet content. If stdin is piped, it reads directly.
// Otherwise it opens $EDITOR, falling back to interactive stdin.
func getContent() (string, error) {
	// If stdin is piped (not a terminal), read from it directly
	if stat, _ := os.Stdin.Stat(); stat.Mode()&os.ModeCharDevice == 0 {
		return readStdin()
	}

	editor := config.Editor()

	// Try using the editor
	if editor != "" {
		tmpFile, err := os.CreateTemp("", "snap-*.txt")
		if err != nil {
			return "", fmt.Errorf("create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		editorCmd := exec.Command(editor, tmpPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			// If editor fails, fall through to stdin
			fmt.Fprintln(os.Stderr, "Editor failed, reading from stdin instead. Type content then press Ctrl+D (or Ctrl+Z on Windows):")
			return readStdin()
		}

		data, err := os.ReadFile(tmpPath)
		if err != nil {
			return "", fmt.Errorf("read temp file: %w", err)
		}

		content := strings.TrimSpace(string(data))
		if content == "" {
			return "", fmt.Errorf("empty content from editor")
		}
		return content, nil
	}

	fmt.Fprintln(os.Stderr, "Enter snippet content (Ctrl+D / Ctrl+Z to finish):")
	return readStdin()
}

func readStdin() (string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	return strings.Join(lines, "\n"), nil
}
