package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/O-Aditya/snippet-snap/config"
	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/models"
	"github.com/O-Aditya/snippet-snap/internal/tui"
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
			tui.PrintError("--name is required")
			return nil
		}

		content, err := getContent()
		if err != nil {
			tui.PrintError(err.Error())
			return nil
		}

		snippet := &models.Snippet{
			Alias:    name,
			Content:  content,
			Language: lang,
			Tags:     tags,
		}

		if err := snippet.Validate(); err != nil {
			tui.PrintError(err.Error())
			return nil
		}

		id, err := db.InsertSnippet(getDB(), snippet)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
				tui.PrintError("Alias " +
					tui.BrightStyle.Bold(true).Render("\""+name+"\"") +
					" already exists")
				fmt.Println(tui.DimStyle.Render("  Try a different name or use ") +
					tui.AccentStyle.Render("snap edit "+name))
				return nil
			}
			tui.PrintError(err.Error())
			return nil
		}

		// Success — the ONE justified bordered element
		fmt.Println(tui.RenderConfirmBox(name, id, lang, tags))
		fmt.Println()
		fmt.Println(tui.DimStyle.Render("  Run ") +
			tui.AccentStyle.Render("snap copy "+strconv.FormatInt(id, 10)) +
			tui.DimStyle.Render(" to use it"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("name", "n", "", "unique alias for the snippet (required)")
	addCmd.Flags().StringP("lang", "l", "", "language for syntax highlighting")
	addCmd.Flags().StringP("tags", "t", "", "comma-separated tags")
}

func getContent() (string, error) {
	if stat, _ := os.Stdin.Stat(); stat.Mode()&os.ModeCharDevice == 0 {
		return readStdin()
	}

	editor := config.Editor()

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
			tui.PrintWarn("Editor failed, reading from stdin instead.")
			fmt.Println(tui.DimStyle.Render("  Type content then press Ctrl+D (or Ctrl+Z on Windows):"))
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

	tui.PrintInfo("Enter snippet content (Ctrl+D / Ctrl+Z to finish):")
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
