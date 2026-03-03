package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/models"
	"github.com/O-Aditya/snippet-snap/internal/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved snippets",
	Long:  `List all snippets with optional language and tag filters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lang, _ := cmd.Flags().GetString("lang")
		tag, _ := cmd.Flags().GetString("tag")

		snippets, err := db.ListSnippets(getDB(), lang, tag)
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}

		w := termWidth()

		if len(snippets) == 0 {
			printEmptyState(w)
			return nil
		}

		return printFlatTable(snippets, w)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("lang", "l", "", "filter by language")
	listCmd.Flags().StringP("tag", "t", "", "filter by tag")
}

// termWidth returns the current terminal width, defaulting to 100.
func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 100
	}
	return w
}

// printEmptyState — no background box, just centered text.
func printEmptyState(w int) {
	content := tui.DimStyle.Render("no snippets yet") + "\n" +
		tui.DimStyle.Render("run ") +
		lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true).Render("snap add") +
		tui.DimStyle.Render(" to save your first one")
	fmt.Println(lipgloss.Place(w, 6, lipgloss.Center, lipgloss.Center, content))
}

func printFlatTable(snippets []models.Snippet, w int) error {
	// ── HEADER — no background on line, only wordmark pill has bg ──
	wordmark := tui.RenderWordmark()
	subtitle := tui.DimStyle.Render("All Snippets")
	rightStr := lipgloss.NewStyle().Foreground(tui.ColorAccent).Render(strconv.Itoa(len(snippets))) +
		tui.DimStyle.Render(" total")
	headerFill := w - lipgloss.Width(wordmark) - lipgloss.Width(subtitle) - lipgloss.Width(rightStr) - 6
	if headerFill < 0 {
		headerFill = 0
	}
	headerLine := wordmark + "  " + subtitle + strings.Repeat(" ", headerFill) + rightStr
	fmt.Println(headerLine)
	fmt.Println(tui.BorderStyle.Render(strings.Repeat("─", w)))

	// ── COLUMN HEADERS — no background ──
	colID := 5
	colAlias := 22
	colLang := 10
	colSaved := 10
	colTags := w - colID - colAlias - colLang - colSaved - 14
	if colTags < 10 {
		colTags = 10
	}

	hdr := " " +
		tui.DimStyle.Width(colID).Align(lipgloss.Right).Render("#") + "  " +
		tui.DimStyle.Width(colAlias).Render("ALIAS") + "  " +
		tui.DimStyle.Width(colLang).Render("LANG") + "  " +
		tui.DimStyle.Width(colTags).Render("TAGS") + "  " +
		tui.DimStyle.Width(colSaved).Align(lipgloss.Right).Render("SAVED")
	fmt.Println(hdr)
	fmt.Println(tui.BorderStyle.Render(strings.Repeat("─", w)))

	// ── ROWS — no background, terminal shows through ──
	for _, s := range snippets {
		id := tui.DimStyle.Width(colID).Align(lipgloss.Right).
			Render(strconv.Itoa(int(s.ID)))

		alias := s.Alias
		if len(alias) > 20 {
			alias = alias[:19] + "…"
		}
		aliasStr := tui.BrightStyle.Bold(true).Width(colAlias).Render(alias)

		langStr := lipgloss.NewStyle().Width(colLang).Render(tui.RenderLangBadge(s.Language))

		tagsRendered := truncateTags(s.Tags, colTags)
		tagsStr := lipgloss.NewStyle().Width(colTags).Render(tagsRendered)

		saved := tui.DimStyle.Width(colSaved).Align(lipgloss.Right).
			Render(tui.RelativeTime(s.UpdatedAt))

		row := id + "  " + aliasStr + "  " + langStr + "  " + tagsStr + "  " + saved
		fmt.Println(lipgloss.NewStyle().Padding(0, 1).Render(" " + row))
		fmt.Println(tui.BorderStyle.Render(strings.Repeat("─", w)))
	}

	// ── FOOTER — BgStatusBar anchors the bottom ──
	footerContent := "  " +
		tui.RenderKey("snap find") + " · " +
		tui.RenderKey("snap copy <id>") + " · " +
		tui.RenderKey("snap add")
	footer := lipgloss.NewStyle().
		Background(tui.BgStatusBar).
		Foreground(tui.ColorDim).
		Padding(0, 1).
		Width(w).
		Render(footerContent)
	fmt.Println(footer)
	return nil
}

// truncateTags renders tag badges, truncating with "+N" if they don't fit.
func truncateTags(tags string, maxW int) string {
	if tags == "" {
		return ""
	}
	parts := strings.Split(tags, ",")
	var result []string
	width := 0
	shown := 0
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		badge := tui.RenderTagBadge(p)
		bw := lipgloss.Width(badge)
		if width > 0 && width+bw+1 > maxW-8 {
			remaining := len(parts) - shown
			if remaining > 0 {
				more := tui.DimStyle.Render(fmt.Sprintf("+%d", remaining))
				result = append(result, more)
			}
			break
		}
		result = append(result, badge)
		width += bw + 1
		shown++
	}
	return strings.Join(result, " ")
}
