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
		leftW := msg.Width * 36 / 100
		rightW := msg.Width - leftW - 1
		f.preview.Width = rightW - 6
		if f.preview.Width < 10 {
			f.preview.Width = 10
		}
		f.preview.Height = msg.Height - 10
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
			if len(f.filtered) == 0 {
				return f, nil
			}
			selected := f.filtered[f.cursor]

			// ResolveVars handles both cases:
			// — no vars: returns content unchanged immediately
			// — has vars: prompts user on stderr, returns resolved
			resolved, err := inject.ResolveVars(selected.Content)
			if err != nil {
				f.statusMsg = "✗ Aborted"
				return f, nil
			}

			if err := clipboard.Copy(resolved); err != nil {
				f.statusMsg = fmt.Sprintf("✗ Copy failed: %v", err)
				return f, nil
			}

			vars := inject.FindVars(selected.Content)
			if len(vars) > 0 {
				f.statusMsg = fmt.Sprintf("✓ Copied %s (%d var(s) resolved)", selected.Alias, len(vars))
			} else {
				f.statusMsg = "✓ Copied " + selected.Alias
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

	var inputCmd tea.Cmd
	f.searchInput, inputCmd = f.searchInput.Update(msg)
	cmds = append(cmds, inputCmd)

	f.filterSnippets()
	f.updatePreview()

	return f, tea.Batch(cmds...)
}

// View implements bubbletea.Model.
func (f Finder) View() string {
	if f.quitting {
		return ""
	}

	leftW := f.width * 36 / 100
	if leftW < 20 {
		leftW = 20
	}
	rightW := f.width - leftW - 1
	if rightW < 20 {
		rightW = 20
	}
	bodyH := f.height - 7
	if bodyH < 3 {
		bodyH = 3
	}

	var sections []string

	// ─── ROW 1: TITLE ROW — no background, terminal shows through ──
	wordmark := RenderWordmark()
	countStr := DimStyle.Render(fmt.Sprintf("%d / %d", len(f.filtered), len(f.allSnippets)))
	titleFill := f.width - lipgloss.Width(wordmark) - lipgloss.Width(countStr) - 2
	if titleFill < 0 {
		titleFill = 0
	}
	titleRow := lipgloss.NewStyle().Width(f.width).
		Render(wordmark + strings.Repeat(" ", titleFill) + countStr)
	sections = append(sections, titleRow)

	// ─── ROW 2: SEARCH ROW — no bg on outer, bg only on input box ──
	searchIcon := AccentStyle.Render("⌕  ")
	searchBox := lipgloss.NewStyle().
		Background(BgInput).
		Border(lipgloss.NormalBorder()).
		BorderForeground(BgBadgeCyan).
		Padding(0, 1).
		Render(f.searchInput.View())
	searchRow := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ColorBorder).
		Width(f.width).
		Render(searchIcon + searchBox)
	sections = append(sections, searchRow)

	// ─── BODY: TWO PANES ──────────────────────────

	// LEFT PANE — list items, no background on pane
	var leftLines []string
	if len(f.filtered) == 0 {
		emptyMsg := DimStyle.Render("no results")
		emptyBlock := lipgloss.Place(leftW-1, bodyH, lipgloss.Center, lipgloss.Center, emptyMsg)
		leftLines = append(leftLines, emptyBlock)
	} else {
		itemH := 3 // padding top + content + padding bottom
		visibleItems := bodyH / itemH
		if visibleItems < 1 {
			visibleItems = bodyH
		}
		start := 0
		if f.cursor >= start+visibleItems {
			start = f.cursor - visibleItems + 1
		}
		end := start + visibleItems
		if end > len(f.filtered) {
			end = len(f.filtered)
		}

		for i := start; i < end; i++ {
			s := f.filtered[i]

			aliasMaxW := leftW - 14
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
				// SELECTED — the ONE row with a background
				row := lipgloss.NewStyle().
					Background(BgSelected).
					Foreground(ColorAccent).
					Bold(true).
					Width(leftW-1).
					Padding(1, 1).
					Render("▸ " + line)
				leftLines = append(leftLines, row)
			} else {
				// NORMAL — no background, terminal shows through
				row := lipgloss.NewStyle().
					Foreground(ColorBright).
					Width(leftW-1).
					Padding(1, 2).
					Border(lipgloss.NormalBorder(), false, false, true, false).
					BorderForeground(ColorBorder).
					Render("  " + line)
				leftLines = append(leftLines, row)
			}
		}
	}

	leftContent := strings.Join(leftLines, "\n")
	// Pane divider: single │ character via right border, NO background
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(ColorBorder).
		Width(leftW).
		Height(bodyH).
		Render(leftContent)

	// RIGHT PANE — no background, terminal shows through
	var rightParts []string

	if len(f.filtered) > 0 && f.showPreview {
		selected := f.filtered[f.cursor]

		// Preview title — no background
		previewTitle := AccentStyle.Render("◈  ") +
			BrightStyle.Bold(true).Render(selected.Alias)
		langBadge := RenderLangBadge(selected.Language)
		if langBadge != "" {
			previewTitle += "  " + langBadge
		}
		titleLine := lipgloss.NewStyle().Padding(0, 2).Width(rightW).Render(previewTitle)
		rightParts = append(rightParts, titleLine)

		// Tags line — no background
		extraH := 0
		if selected.Tags != "" {
			tagsLine := lipgloss.NewStyle().Padding(0, 2).Width(rightW).
				Render(DimStyle.Render("tags  ") + RenderTagBadges(selected.Tags))
			rightParts = append(rightParts, tagsLine)
			extraH = 1
		}

		// Divider
		divider := BorderStyle.Render(strings.Repeat("─", rightW))
		rightParts = append(rightParts, divider)

		// Content — no background wrapper
		contentH := bodyH - 2 - extraH
		if contentH < 1 {
			contentH = 1
		}
		previewContent := lipgloss.NewStyle().
			Width(rightW).
			Height(contentH).
			Padding(1, 2).
			Render(f.preview.View())
		rightParts = append(rightParts, previewContent)

	} else if !f.showPreview {
		msg := DimStyle.Render("preview hidden — press Tab")
		rightParts = append(rightParts, lipgloss.Place(rightW, bodyH, lipgloss.Center, lipgloss.Center, msg))
	} else {
		noSel := DimStyle.Render("◈\n\nno snippet selected")
		rightParts = append(rightParts, lipgloss.Place(rightW, bodyH, lipgloss.Center, lipgloss.Center, noSel))
	}

	rightPane := strings.Join(rightParts, "\n")
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	sections = append(sections, body)

	// ─── ROW LAST: STATUS BAR — BgStatusBar is the ONE full-width bg ──
	hints := []string{
		RenderKey("↑↓") + " " + DimStyle.Render("navigate"),
		RenderKey("enter") + " " + DimStyle.Render("copy"),
		RenderKey("tab") + " " + DimStyle.Render("preview"),
		RenderKey("esc") + " " + DimStyle.Render("quit"),
	}
	leftStr := strings.Join(hints, "   ")

	var statusRight string
	if f.statusMsg != "" {
		if strings.HasPrefix(f.statusMsg, "✓") {
			statusRight = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true).Render(f.statusMsg)
		} else if strings.HasPrefix(f.statusMsg, "✗") {
			statusRight = lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render(f.statusMsg)
		}
	}

	statusFill := f.width - lipgloss.Width(leftStr) - lipgloss.Width(statusRight) - 6
	if statusFill < 0 {
		statusFill = 0
	}
	barContent := leftStr + strings.Repeat(" ", statusFill) + statusRight
	statusBar := lipgloss.NewStyle().
		Background(BgStatusBar).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorBorder).
		Padding(0, 2).
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
