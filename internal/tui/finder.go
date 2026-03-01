package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/O-Aditya/snippet-snap/internal/clipboard"
	"github.com/O-Aditya/snippet-snap/internal/highlight"
	"github.com/O-Aditya/snippet-snap/internal/inject"
	"github.com/O-Aditya/snippet-snap/internal/models"
)

// Finder is the Bubble Tea model for the `snap find` TUI.
type Finder struct {
	allSnippets []models.Snippet
	filtered    []models.Snippet
	cursor      int
	searchInput textinput.Model
	preview     viewport.Model
	keys        KeyMap
	width       int
	height      int
	showPreview bool
	statusMsg   string
	quitting    bool
}

// NewFinder creates a new Finder model with the given snippets.
func NewFinder(snippets []models.Snippet) Finder {
	ti := textinput.New()
	ti.Placeholder = "Type to search snippets..."
	ti.Focus()
	ti.PromptStyle = SearchPromptStyle
	ti.Prompt = "  "
	ti.CharLimit = 256

	vp := viewport.New(80, 10)

	return Finder{
		allSnippets: snippets,
		filtered:    snippets,
		searchInput: ti,
		preview:     vp,
		keys:        DefaultKeyMap(),
		showPreview: true,
		width:       80,
		height:      24,
	}
}

// Init implements bubbletea.Model.
func (f Finder) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements bubbletea.Model.
func (f Finder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		f.height = msg.Height
		leftW := msg.Width * 37 / 100
		rightW := msg.Width - leftW - 1
		f.preview.Width = rightW - 4
		if f.preview.Width < 10 {
			f.preview.Width = 10
		}
		f.preview.Height = msg.Height - 8
		if f.preview.Height < 3 {
			f.preview.Height = 3
		}
		f.updatePreview()
		return f, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keys.Quit):
			f.quitting = true
			return f, tea.Quit

		case key.Matches(msg, f.keys.Up):
			if f.cursor > 0 {
				f.cursor--
				f.updatePreview()
			}
			return f, nil

		case key.Matches(msg, f.keys.Down):
			if f.cursor < len(f.filtered)-1 {
				f.cursor++
				f.updatePreview()
			}
			return f, nil

		case key.Matches(msg, f.keys.Tab):
			f.showPreview = !f.showPreview
			return f, nil

		case key.Matches(msg, f.keys.Enter):
			if len(f.filtered) > 0 {
				selected := f.filtered[f.cursor]
				content := selected.Content

				vars := inject.FindVars(content)
				if len(vars) > 0 {
					f.statusMsg = fmt.Sprintf("✓ Copied %s (%d vars — use 'snap copy %d' to fill)",
						selected.Alias, len(vars), selected.ID)
				} else {
					f.statusMsg = fmt.Sprintf("✓ Copied %s", selected.Alias)
				}

				if err := clipboard.Copy(content); err != nil {
					f.statusMsg = fmt.Sprintf("✗ Copy failed: %v", err)
				}
			}
			return f, nil

		case key.Matches(msg, f.keys.PageUp):
			f.preview.HalfViewUp()
			return f, nil

		case key.Matches(msg, f.keys.PageDown):
			f.preview.HalfViewDown()
			return f, nil
		}
	}

	// Update search input
	var inputCmd tea.Cmd
	f.searchInput, inputCmd = f.searchInput.Update(msg)
	cmds = append(cmds, inputCmd)

	// Re-filter on search change
	f.filterSnippets()
	f.updatePreview()

	return f, tea.Batch(cmds...)
}

// View implements bubbletea.Model.
func (f Finder) View() string {
	if f.quitting {
		return ""
	}

	leftW := f.width * 37 / 100
	if leftW < 20 {
		leftW = 20
	}
	rightW := f.width - leftW - 1
	if rightW < 20 {
		rightW = 20
	}
	bodyH := f.height - 6
	if bodyH < 3 {
		bodyH = 3
	}

	var sections []string

	// ─── LINE 1: HEADER BAR ────────────────────────
	wordmark := lipgloss.NewStyle().
		Background(ColorCyan).
		Foreground(ColorBG).
		Bold(true).
		Padding(0, 2).
		Render("◈ SNIPPET-SNAP")

	countStr := lipgloss.NewStyle().Foreground(ColorMuted).
		Render(fmt.Sprintf("%d/%d snippets", len(f.filtered), len(f.allSnippets)))

	headerFill := f.width - lipgloss.Width(wordmark) - lipgloss.Width(countStr) - 2
	if headerFill < 0 {
		headerFill = 0
	}
	headerInner := wordmark + strings.Repeat(" ", headerFill) + countStr
	header := lipgloss.NewStyle().
		Background(ColorBG2).
		Width(f.width).
		Render(" " + headerInner + " ")
	sections = append(sections, header)

	// ─── LINE 2: SEARCH BAR ───────────────────────
	searchIcon := lipgloss.NewStyle().Foreground(ColorCyan).Render(" ⌕ ")

	searchBoxContent := f.searchInput.View()
	searchBox := lipgloss.NewStyle().
		Background(ColorBG3).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorCyanDim).
		Padding(0, 1).
		Render(searchBoxContent)

	searchInner := searchIcon + " " + searchBox
	searchRow := lipgloss.NewStyle().
		Background(ColorBG2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ColorBorder).
		Width(f.width).
		Render(searchInner)
	sections = append(sections, searchRow)

	// ─── BODY: LEFT PANE (list) + RIGHT PANE (preview) ──

	// LEFT PANE
	var leftLines []string
	if len(f.filtered) == 0 {
		emptyMsg := lipgloss.NewStyle().Foreground(ColorDimC).Render("no results")
		emptyBlock := lipgloss.Place(leftW-1, bodyH, lipgloss.Center, lipgloss.Center, emptyMsg)
		leftLines = append(leftLines, emptyBlock)
	} else {
		start := 0
		if f.cursor >= bodyH {
			start = f.cursor - bodyH + 1
		}
		end := start + bodyH
		if end > len(f.filtered) {
			end = len(f.filtered)
		}

		for i := start; i < end; i++ {
			s := f.filtered[i]

			aliasMaxW := leftW - 16
			if aliasMaxW < 8 {
				aliasMaxW = 8
			}
			alias := s.Alias
			if len(alias) > aliasMaxW {
				alias = alias[:aliasMaxW-1] + "…"
			}

			line := alias
			langBadge := RenderLangBadge(s.Language)
			if langBadge != "" {
				line += "  " + langBadge
			}

			if i == f.cursor {
				leftLines = append(leftLines, SelectedItemStyle.Width(leftW-1).Render("▸ "+line))
			} else {
				leftLines = append(leftLines, NormalItemStyle.Width(leftW-1).Render("  "+line))
			}
		}

		// Pad to full height
		for len(leftLines) < bodyH {
			leftLines = append(leftLines, strings.Repeat(" ", leftW-1))
		}
	}

	leftPane := DividerStyle.Width(leftW).Height(bodyH).Render(strings.Join(leftLines, "\n"))

	// RIGHT PANE
	var rightParts []string

	if len(f.filtered) > 0 && f.showPreview {
		selected := f.filtered[f.cursor]

		// Preview header
		previewIcon := lipgloss.NewStyle().Foreground(ColorCyan).Render("  ◈ ")
		previewAlias := lipgloss.NewStyle().Foreground(ColorText).Bold(true).Render(selected.Alias)
		langBadge := RenderLangBadge(selected.Language)
		previewTitle := previewIcon + previewAlias
		if langBadge != "" {
			previewTitle += "  " + langBadge
		}
		previewHdr := PreviewHeaderStyle.Width(rightW).Render(previewTitle)
		rightParts = append(rightParts, previewHdr)

		// Tags row
		extraHeaderLines := 0
		if selected.Tags != "" {
			tagsLabel := lipgloss.NewStyle().Foreground(ColorMuted).Render("  tags  ")
			tagsRow := lipgloss.NewStyle().
				Background(ColorBG2).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(ColorBorder).
				Width(rightW).
				Render(tagsLabel + RenderTagBadges(selected.Tags))
			rightParts = append(rightParts, tagsRow)
			extraHeaderLines = 2
		}

		// Divider
		divider := lipgloss.NewStyle().Foreground(ColorDimC).Render(strings.Repeat("─", rightW))
		rightParts = append(rightParts, divider)

		// Content
		contentH := bodyH - 3 - extraHeaderLines
		if contentH < 1 {
			contentH = 1
		}
		previewContent := lipgloss.NewStyle().
			Width(rightW).
			Height(contentH).
			Padding(0, 1).
			Render(f.preview.View())
		rightParts = append(rightParts, previewContent)

	} else if !f.showPreview {
		msg := lipgloss.NewStyle().Foreground(ColorDimC).Render("preview hidden — press Tab")
		rightParts = append(rightParts, lipgloss.Place(rightW, bodyH, lipgloss.Center, lipgloss.Center, msg))
	} else {
		noSel := lipgloss.NewStyle().Foreground(ColorDimC).Render("◈\n\nno snippet selected")
		rightParts = append(rightParts, lipgloss.Place(rightW, bodyH, lipgloss.Center, lipgloss.Center, noSel))
	}

	rightPane := strings.Join(rightParts, "\n")
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	sections = append(sections, body)

	// ─── STATUS BAR ────────────────────────────────
	hints := []string{
		KeyBadgeStyle.Render("↑↓") + " " + lipgloss.NewStyle().Foreground(ColorMuted).Render("nav"),
		KeyBadgeStyle.Render("enter") + " " + lipgloss.NewStyle().Foreground(ColorMuted).Render("copy"),
		KeyBadgeStyle.Render("e") + " " + lipgloss.NewStyle().Foreground(ColorMuted).Render("edit"),
		KeyBadgeStyle.Render("d") + " " + lipgloss.NewStyle().Foreground(ColorMuted).Render("delete"),
		KeyBadgeStyle.Render("tab") + " " + lipgloss.NewStyle().Foreground(ColorMuted).Render("preview"),
		KeyBadgeStyle.Render("esc") + " " + lipgloss.NewStyle().Foreground(ColorMuted).Render("quit"),
	}
	leftStr := strings.Join(hints, "  ")

	var statusRight string
	if f.statusMsg != "" {
		if strings.HasPrefix(f.statusMsg, "✓") {
			statusRight = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true).Render(f.statusMsg)
		} else if strings.HasPrefix(f.statusMsg, "✗") {
			statusRight = lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render(f.statusMsg)
		} else {
			statusRight = lipgloss.NewStyle().Foreground(ColorMuted).Render(f.statusMsg)
		}
	}

	statusFill := f.width - lipgloss.Width(leftStr) - lipgloss.Width(statusRight) - 4
	if statusFill < 0 {
		statusFill = 0
	}
	barContent := " " + leftStr + strings.Repeat(" ", statusFill) + statusRight + " "
	statusBar := lipgloss.NewStyle().
		Background(ColorBG2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorBorder).
		Width(f.width).
		Render(barContent)
	sections = append(sections, statusBar)

	return strings.Join(sections, "\n")
}

// filterSnippets applies fuzzy matching on the search query.
func (f *Finder) filterSnippets() {
	query := strings.ToLower(strings.TrimSpace(f.searchInput.Value()))
	if query == "" {
		f.filtered = f.allSnippets
		if f.cursor >= len(f.filtered) {
			f.cursor = maxInt(0, len(f.filtered)-1)
		}
		return
	}

	var result []models.Snippet
	for _, s := range f.allSnippets {
		target := strings.ToLower(s.Alias + " " + s.Content + " " + s.Tags)
		if fuzzyMatch(query, target) {
			result = append(result, s)
		}
	}

	f.filtered = result
	if f.cursor >= len(f.filtered) {
		f.cursor = maxInt(0, len(f.filtered)-1)
	}
}

// updatePreview renders the currently selected snippet into the preview viewport.
func (f *Finder) updatePreview() {
	if len(f.filtered) == 0 {
		f.preview.SetContent("(no snippet selected)")
		return
	}
	s := f.filtered[f.cursor]
	rendered, err := highlight.Render(s.Content, s.Language)
	if err != nil {
		rendered = s.Content
	}
	f.preview.SetContent(rendered)
	f.preview.GotoTop()
}

// fuzzyMatch checks if all characters in the query appear in order within the target.
func fuzzyMatch(query, target string) bool {
	qi := 0
	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
