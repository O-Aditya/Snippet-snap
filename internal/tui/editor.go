package tui

import (
	"database/sql"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/O-Aditya/snippet-snap/internal/db"
	"github.com/O-Aditya/snippet-snap/internal/models"
)

// EditorModel is a standalone Bubble Tea model for adding snippets interactively.
type EditorModel struct {
	aliasInput  textinput.Model
	langInput   textinput.Model
	tagsInput   textinput.Model
	contentArea textarea.Model

	focusIndex int // 0=alias 1=lang 2=tags 3=content
	err        string
	saved      bool
	confirming bool // true when showing Esc discard confirmation
	database   *sql.DB

	// Exported result for caller
	SavedID    int64
	SavedAlias string
	SavedLang  string
	SavedTags  string
}

// NewEditorModel creates a new EditorModel ready for use.
func NewEditorModel(database *sql.DB) EditorModel {
	alias := textinput.New()
	alias.Placeholder = "my-snippet"
	alias.Focus()
	alias.CharLimit = 64
	alias.PromptStyle = AccentStyle
	alias.Prompt = "  "

	lang := textinput.New()
	lang.Placeholder = "bash, go, python, sql..."
	lang.CharLimit = 32
	lang.PromptStyle = DimStyle
	lang.Prompt = "  "

	tags := textinput.New()
	tags.Placeholder = "docker, cleanup, devops"
	tags.CharLimit = 256
	tags.PromptStyle = DimStyle
	tags.Prompt = "  "

	content := textarea.New()
	content.Placeholder = "Paste or type your snippet here..."
	content.CharLimit = 0 // unlimited
	content.ShowLineNumbers = true
	content.SetWidth(60)
	content.SetHeight(8)

	return EditorModel{
		aliasInput:  alias,
		langInput:   lang,
		tagsInput:   tags,
		contentArea: content,
		focusIndex:  0,
		database:    database,
	}
}

func (m EditorModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear previous error on any keystroke
		if m.err != "" && msg.String() != "ctrl+s" {
			m.err = ""
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.confirming {
				// Second Esc — discard and quit
				return m, tea.Quit
			}
			// First Esc — show confirmation
			m.confirming = true
			return m, nil

		case "ctrl+s":
			m.confirming = false
			return m.save()

		case "tab":
			m.confirming = false
			return m.cycleFocus(1)

		case "shift+tab":
			m.confirming = false
			return m.cycleFocus(-1)
		}

		// Any other key cancels discard confirmation
		if m.confirming {
			m.confirming = false
		}

	case tea.WindowSizeMsg:
		w := msg.Width - 8
		if w < 40 {
			w = 40
		}
		if w > 80 {
			w = 80
		}
		m.contentArea.SetWidth(w)
		h := msg.Height - 18
		if h < 4 {
			h = 4
		}
		if h > 20 {
			h = 20
		}
		m.contentArea.SetHeight(h)
	}

	// Route input to focused field
	return m.updateFocusedInput(msg)
}

func (m EditorModel) cycleFocus(dir int) (tea.Model, tea.Cmd) {
	// Blur all
	m.aliasInput.Blur()
	m.langInput.Blur()
	m.tagsInput.Blur()
	m.contentArea.Blur()

	m.focusIndex = (m.focusIndex + dir + 4) % 4

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		cmd = m.aliasInput.Focus()
	case 1:
		cmd = m.langInput.Focus()
	case 2:
		cmd = m.tagsInput.Focus()
	case 3:
		cmd = m.contentArea.Focus()
	}
	return m, cmd
}

func (m EditorModel) updateFocusedInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.aliasInput, cmd = m.aliasInput.Update(msg)
	case 1:
		m.langInput, cmd = m.langInput.Update(msg)
	case 2:
		m.tagsInput, cmd = m.tagsInput.Update(msg)
	case 3:
		m.contentArea, cmd = m.contentArea.Update(msg)
	}
	return m, cmd
}

func (m EditorModel) save() (tea.Model, tea.Cmd) {
	alias := strings.TrimSpace(m.aliasInput.Value())
	lang := strings.TrimSpace(m.langInput.Value())
	tags := strings.TrimSpace(m.tagsInput.Value())
	content := strings.TrimSpace(m.contentArea.Value())

	// Validation
	if alias == "" {
		m.err = "alias is required"
		m.focusIndex = 0
		m.aliasInput.Focus()
		return m, nil
	}
	if lang == "" {
		m.err = "language is required"
		m.focusIndex = 1
		m.langInput.Focus()
		return m, nil
	}
	if content == "" {
		m.err = "content is required"
		m.focusIndex = 3
		m.contentArea.Focus()
		return m, nil
	}

	snippet := &models.Snippet{
		Alias:    alias,
		Content:  content,
		Language: lang,
		Tags:     tags,
	}

	id, err := db.InsertSnippet(m.database, snippet)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			m.err = "alias \"" + alias + "\" already exists"
			m.focusIndex = 0
			m.aliasInput.Focus()
			return m, nil
		}
		m.err = err.Error()
		return m, nil
	}

	m.saved = true
	m.SavedID = id
	m.SavedAlias = alias
	m.SavedLang = lang
	m.SavedTags = tags
	return m, tea.Quit
}

// IsSaved returns true if the snippet was successfully saved.
func (m EditorModel) IsSaved() bool {
	return m.saved
}

func (m EditorModel) View() string {
	if m.saved {
		return ""
	}

	var b strings.Builder

	// Title bar
	title := lipgloss.NewStyle().
		Background(BgWordmark).
		Foreground(lipgloss.Color("#0D1117")).
		Bold(true).
		Padding(0, 2).
		Render("◈  NEW SNIPPET")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Field labels and inputs
	labels := []string{"Alias", "Lang", "Tags", "Content"}
	fields := []string{
		m.aliasInput.View(),
		m.langInput.View(),
		m.tagsInput.View(),
		"", // content is rendered separately
	}

	labelStyle := DimStyle.Width(10)
	focusedLabel := lipgloss.NewStyle().Foreground(ColorAccent).Width(10)

	for i := 0; i < 3; i++ {
		ls := labelStyle
		if i == m.focusIndex {
			ls = focusedLabel
		}
		b.WriteString("  " + ls.Render(labels[i]) + " " + fields[i] + "\n")
	}

	// Content field with border
	b.WriteString("\n")
	contentLabel := labelStyle
	if m.focusIndex == 3 {
		contentLabel = focusedLabel
	}
	b.WriteString("  " + contentLabel.Render(labels[3]) + "\n")

	contentBorder := lipgloss.NormalBorder()
	borderColor := ColorBorder
	if m.focusIndex == 3 {
		borderColor = ColorAccent
	}
	contentBox := lipgloss.NewStyle().
		Border(contentBorder).
		BorderForeground(borderColor).
		Padding(0, 1).
		MarginLeft(2).
		Render(m.contentArea.View())
	b.WriteString(contentBox)
	b.WriteString("\n")

	// Error message
	if m.err != "" {
		errMsg := lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true).
			MarginLeft(2).
			Render("  ✗ " + m.err)
		b.WriteString("\n" + errMsg)
	}

	// Discard confirmation
	if m.confirming {
		warn := lipgloss.NewStyle().
			Foreground(ColorAmber).
			Bold(true).
			MarginLeft(2).
			Render("  ⚠ Press Esc again to discard and quit")
		b.WriteString("\n" + warn)
	}

	// Status bar
	b.WriteString("\n\n")
	hints := []string{
		RenderKey("Tab") + " " + DimStyle.Render("next"),
		RenderKey("S-Tab") + " " + DimStyle.Render("prev"),
		RenderKey("Ctrl+S") + " " + DimStyle.Render("save"),
		RenderKey("Esc") + " " + DimStyle.Render("discard"),
	}
	bar := lipgloss.NewStyle().
		Background(BgStatusBar).
		Padding(0, 2).
		Render(strings.Join(hints, "   "))
	b.WriteString(bar)

	return b.String()
}
