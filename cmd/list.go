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

		return printShortList(snippets, w)
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

// printEmptyState shows a centered empty box.
func printEmptyState(w int) {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ColorDim).
		Padding(1, 4).
		Render("✦  No snippets yet\n\n" +
			lipgloss.NewStyle().Foreground(tui.ColorDim).Render("Run ") +
			lipgloss.NewStyle().Foreground(tui.ColorCyan).Bold(true).Render("snap add") +
			lipgloss.NewStyle().Foreground(tui.ColorDim).Render(" to save your first snippet"))
	fmt.Println(lipgloss.Place(w, 8, lipgloss.Center, lipgloss.Center, box))
}

// printHeaderBar renders the top brand bar.
func printHeaderBar(count int, w int) {
	snapLogo := lipgloss.NewStyle().
		Background(tui.ColorCyan).
		Foreground(tui.ColorBG).
		Bold(true).
		Padding(0, 1).
		Render("◈ snap")

	leftStr := snapLogo + "  " + lipgloss.NewStyle().Foreground(tui.ColorMuted).Render("All Snippets")
	rightStr := lipgloss.NewStyle().Foreground(tui.ColorCyan).Render(strconv.Itoa(count)) +
		lipgloss.NewStyle().Foreground(tui.ColorMuted).Render(" total")

	fill := w - lipgloss.Width(leftStr) - lipgloss.Width(rightStr) - 4
	if fill < 0 {
		fill = 0
	}
	barInner := " " + leftStr + strings.Repeat(" ", fill) + rightStr + " "
	headerBar := lipgloss.NewStyle().
		Background(tui.ColorBG2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(tui.ColorBorder).
		Padding(0, 1).
		Width(w).
		Render(barInner)
	fmt.Println(headerBar)
}

func printShortList(snippets []models.Snippet, w int) error {
	printHeaderBar(len(snippets), w)

	// Column widths
	colID := 5
	colAlias := 22
	colLang := 10
	colSaved := 10
	colTags := w - colID - colAlias - colLang - colSaved - 14
	if colTags < 10 {
		colTags = 10
	}

	// Column headers
	dimHdr := lipgloss.NewStyle().Foreground(tui.ColorDim)
	headerRow := " " +
		dimHdr.Width(colID).Align(lipgloss.Right).Render("#") + "  " +
		dimHdr.Width(colAlias).Render("ALIAS") + "  " +
		dimHdr.Width(colLang).Render("LANG") + "  " +
		dimHdr.Width(colTags).Render("TAGS") + "  " +
		dimHdr.Width(colSaved).Align(lipgloss.Right).Render("SAVED")
	colHdrBar := lipgloss.NewStyle().
		Background(tui.ColorBG2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(tui.ColorBorder).
		Padding(0, 1).
		Width(w).
		Render(headerRow)
	fmt.Println(colHdrBar)

	// Data rows
	for idx, s := range snippets {
		var rowBg lipgloss.Style
		if idx%2 == 1 {
			rowBg = lipgloss.NewStyle().Background(tui.ColorBG4)
		} else {
			rowBg = lipgloss.NewStyle().Background(tui.ColorBG)
		}

		id := lipgloss.NewStyle().Foreground(tui.ColorDim).Width(colID).Align(lipgloss.Right).
			Render(strconv.Itoa(int(s.ID)))

		alias := s.Alias
		if len(alias) > 20 {
			alias = alias[:19] + "…"
		}
		aliasStr := lipgloss.NewStyle().Foreground(tui.ColorBright).Bold(true).Width(colAlias).Render(alias)

		langStr := lipgloss.NewStyle().Width(colLang).Render(tui.RenderLangBadge(s.Language))

		tagsRendered := truncateTags(s.Tags, colTags)
		tagsStr := lipgloss.NewStyle().Width(colTags).Render(tagsRendered)

		saved := lipgloss.NewStyle().Foreground(tui.ColorDim).Width(colSaved).Align(lipgloss.Right).
			Render(tui.RelativeTime(s.UpdatedAt))

		row := id + "  " + aliasStr + "  " + langStr + "  " + tagsStr + "  " + saved

		rowLine := rowBg.
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(tui.ColorBorder).
			Padding(0, 1).
			Width(w).
			Render(" " + row)

		fmt.Println(rowLine)
	}

	// Footer
	codeStyle := lipgloss.NewStyle().
		Background(tui.ColorBG3).
		Foreground(tui.ColorCyan).
		Padding(0, 1)

	footerContent := "  " +
		codeStyle.Render("snap find") + " · " +
		codeStyle.Render("snap copy <id>") + " · " +
		codeStyle.Render("snap add")

	footer := lipgloss.NewStyle().
		Background(tui.ColorBG2).
		Foreground(tui.ColorDim).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(tui.ColorBorder).
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
				more := lipgloss.NewStyle().Foreground(tui.ColorDim).Render(fmt.Sprintf("+%d", remaining))
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
