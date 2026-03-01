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
	ti.Prompt = "🔍 "
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
		f.preview.Width = msg.Width - 4
		previewHeight := msg.Height / 3
		if previewHeight < 5 {
			previewHeight = 5
		}
		f.preview.Height = previewHeight
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

				// Resolve vars if present
				vars := inject.FindVars(content)
				if len(vars) > 0 {
					// In TUI mode, copy raw content (vars unresolved)
					// The user can use `snap copy <id>` for interactive resolution
					f.statusMsg = fmt.Sprintf("Copied %s (has %d vars — use 'snap copy %d' to fill)", selected.Alias, len(vars), selected.ID)
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

	var b strings.Builder

	// Title bar
	title := TitleStyle.Render(" 📋 Snippet-Snap Finder ")
	b.WriteString(title + "\n\n")

	// Search input
	b.WriteString(f.searchInput.View() + "\n\n")

	// Snippet list
	listHeight := f.height - 8
	if f.showPreview {
		listHeight = f.height/2 - 4
	}
	if listHeight < 3 {
		listHeight = 3
	}

	if len(f.filtered) == 0 {
		b.WriteString(DimStyle.Render("  No matching snippets.\n"))
	} else {
		start := 0
		if f.cursor >= listHeight {
			start = f.cursor - listHeight + 1
		}
		end := start + listHeight
		if end > len(f.filtered) {
			end = len(f.filtered)
		}

		for i := start; i < end; i++ {
			s := f.filtered[i]
			line := fmt.Sprintf("[%d] %s", s.ID, s.Alias)
			if s.Language != "" {
				line += "  " + DimStyle.Render(s.Language)
			}
			if s.Tags != "" {
				line += "  " + DimStyle.Render("["+s.Tags+"]")
			}

			if i == f.cursor {
				b.WriteString(SelectedItemStyle.Render("▸ "+line) + "\n")
			} else {
				b.WriteString(NormalItemStyle.Render("  "+line) + "\n")
			}
		}
	}

	// Preview pane
	if f.showPreview && len(f.filtered) > 0 {
		b.WriteString("\n")
		previewTitle := DimStyle.Render("─── Preview ───")
		b.WriteString(previewTitle + "\n")
		previewContent := PreviewBorderStyle.
			Width(f.width - 4).
			Render(f.preview.View())
		b.WriteString(previewContent + "\n")
	}

	// Status bar
	status := f.keys.ShortHelp()
	if f.statusMsg != "" {
		status = f.statusMsg + "  │  " + status
	}
	countInfo := fmt.Sprintf("%d/%d snippets", len(f.filtered), len(f.allSnippets))
	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		StatusBarStyle.Render(countInfo),
		StatusBarStyle.Width(f.width-lipgloss.Width(countInfo)-2).Render(status),
	)
	b.WriteString("\n" + bar)

	return b.String()
}

// filterSnippets applies fuzzy matching on the search query against aliases, content, and tags.
func (f *Finder) filterSnippets() {
	query := strings.ToLower(strings.TrimSpace(f.searchInput.Value()))
	if query == "" {
		f.filtered = f.allSnippets
		if f.cursor >= len(f.filtered) {
			f.cursor = max(0, len(f.filtered)-1)
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
		f.cursor = max(0, len(f.filtered)-1)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
