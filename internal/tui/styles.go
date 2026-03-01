package tui

import "github.com/charmbracelet/lipgloss"

// Color palette — solarized-inspired for readability on dark terminals.
var (
	ColorAccent    = lipgloss.Color("#7C3AED") // violet
	ColorDim       = lipgloss.Color("#6B7280") // gray-500
	ColorHighlight = lipgloss.Color("#34D399") // emerald
	ColorWarning   = lipgloss.Color("#FBBF24") // amber
	ColorBg        = lipgloss.Color("#1F2937") // gray-800
	ColorFg        = lipgloss.Color("#F9FAFB") // gray-50
)

// Styles used throughout the TUI.
var (
	// TitleStyle is used for the header bar.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			Background(ColorAccent).
			Padding(0, 1)

	// SearchPromptStyle styles the search input label.
	SearchPromptStyle = lipgloss.NewStyle().
				Foreground(ColorAccent).
				Bold(true)

	// SelectedItemStyle highlights the currently selected list item.
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorHighlight).
				Bold(true).
				PaddingLeft(2)

	// NormalItemStyle is for non-selected list items.
	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			PaddingLeft(2)

	// DimStyle is for metadata and secondary text.
	DimStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// PreviewBorderStyle wraps the preview pane.
	PreviewBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorAccent).
				Padding(0, 1)

	// StatusBarStyle is for the bottom status bar.
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Padding(0, 1)

	// HelpStyle is for help text at the bottom.
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Italic(true)
)
