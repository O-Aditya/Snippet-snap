package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ──────────────────────────────────────────────────
// Core palette — Clean Terminal Noir
// ──────────────────────────────────────────────────

var (
	ColorBG      = lipgloss.Color("#0D1117")
	ColorBG2     = lipgloss.Color("#161B22")
	ColorBG3     = lipgloss.Color("#1C2333")
	ColorBG4     = lipgloss.Color("#0A0E13")
	ColorBorder  = lipgloss.Color("#21262D")
	ColorBorder2 = lipgloss.Color("#30363D")
	ColorDim     = lipgloss.Color("#484F58")
	ColorMuted   = lipgloss.Color("#7D8590")
	ColorText    = lipgloss.Color("#CDD9E5")
	ColorBright  = lipgloss.Color("#E6EDF3")
	ColorCyan    = lipgloss.Color("#39D0D8")
	ColorCyanDim = lipgloss.Color("#163940")
	ColorGreen   = lipgloss.Color("#3FB950")
	ColorGreenDm = lipgloss.Color("#0D2218")
	ColorRed     = lipgloss.Color("#F85149")
	ColorRedDim  = lipgloss.Color("#2A0F0E")
	ColorAmber   = lipgloss.Color("#CBA135")
	ColorAmberDm = lipgloss.Color("#261D08")
	ColorPurple  = lipgloss.Color("#B48EFF")
	ColorBlue    = lipgloss.Color("#58A6FF")
	ColorBlueDim = lipgloss.Color("#091D36")
)

// ──────────────────────────────────────────────────
// Composite styles
// ──────────────────────────────────────────────────

var (
	SelectedItemStyle = lipgloss.NewStyle().
				Background(ColorBG3).
				Foreground(ColorCyan).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorBright)

	PreviewHeaderStyle = lipgloss.NewStyle().
				Background(ColorBG2).
				Foreground(ColorBright).
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(ColorBorder)

	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorBG2).
			Foreground(ColorMuted).
			Padding(0, 2)

	KeyBadgeStyle = lipgloss.NewStyle().
			Background(ColorBG3).
			Foreground(ColorText).
			Bold(true).
			Padding(0, 1)

	DividerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(ColorBorder)

	DimStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleBold = lipgloss.NewStyle().Bold(true)

	SearchPromptStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).
				Bold(true)
)

// ──────────────────────────────────────────────────
// Language badge renderer
// ──────────────────────────────────────────────────

type langColor struct {
	bg lipgloss.Color
	fg lipgloss.Color
}

var langColors = map[string]langColor{
	"bash":       {lipgloss.Color("#0D2218"), lipgloss.Color("#3FB950")},
	"sh":         {lipgloss.Color("#0D2218"), lipgloss.Color("#3FB950")},
	"zsh":        {lipgloss.Color("#0D2218"), lipgloss.Color("#3FB950")},
	"python":     {lipgloss.Color("#161630"), lipgloss.Color("#7C8CF8")},
	"py":         {lipgloss.Color("#161630"), lipgloss.Color("#7C8CF8")},
	"go":         {lipgloss.Color("#001E2E"), lipgloss.Color("#00ACD7")},
	"golang":     {lipgloss.Color("#001E2E"), lipgloss.Color("#00ACD7")},
	"sql":        {lipgloss.Color("#261D08"), lipgloss.Color("#CBA135")},
	"postgres":   {lipgloss.Color("#261D08"), lipgloss.Color("#CBA135")},
	"yaml":       {lipgloss.Color("#1D1030"), lipgloss.Color("#B48EFF")},
	"toml":       {lipgloss.Color("#1D1030"), lipgloss.Color("#B48EFF")},
	"js":         {lipgloss.Color("#252200"), lipgloss.Color("#F7DF1E")},
	"ts":         {lipgloss.Color("#252200"), lipgloss.Color("#F7DF1E")},
	"javascript": {lipgloss.Color("#252200"), lipgloss.Color("#F7DF1E")},
	"typescript": {lipgloss.Color("#252200"), lipgloss.Color("#F7DF1E")},
}

// RenderLangBadge renders a per-language colored badge.
func RenderLangBadge(lang string) string {
	if lang == "" {
		return ""
	}
	lc, ok := langColors[strings.ToLower(lang)]
	if !ok {
		lc = langColor{bg: ColorBG3, fg: ColorMuted}
	}
	return lipgloss.NewStyle().
		Background(lc.bg).
		Foreground(lc.fg).
		Bold(true).
		Padding(0, 1).
		Render(strings.ToUpper(lang))
}

// ──────────────────────────────────────────────────
// Tag badge renderers
// ──────────────────────────────────────────────────

// RenderTagBadge renders a single styled tag badge.
func RenderTagBadge(tag string) string {
	return lipgloss.NewStyle().
		Background(ColorBlueDim).
		Foreground(ColorBlue).
		Padding(0, 1).
		Render(tag)
}

// RenderTagBadges splits comma-separated tags and renders each as a badge.
func RenderTagBadges(tags string) string {
	if tags == "" {
		return ""
	}
	parts := strings.Split(tags, ",")
	badges := make([]string, 0, len(parts))
	for _, t := range parts {
		t = strings.TrimSpace(t)
		if t != "" {
			badges = append(badges, RenderTagBadge(t))
		}
	}
	return strings.Join(badges, " ")
}

// ──────────────────────────────────────────────────
// Feedback helpers
// ──────────────────────────────────────────────────

// PrintSuccess prints a green success message.
func PrintSuccess(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(ColorGreen).Bold(true).Render("✓ " + msg))
}

// PrintError prints a red error message.
func PrintError(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render("✗ " + msg))
}

// PrintInfo prints a muted info message.
func PrintInfo(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(ColorMuted).Render("ℹ " + msg))
}

// PrintWarn prints an amber warning message.
func PrintWarn(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(ColorAmber).Render("⚠ " + msg))
}

// ──────────────────────────────────────────────────
// Relative time
// ──────────────────────────────────────────────────

// RelativeTime returns a human-friendly relative timestamp.
func RelativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/24/7))
	default:
		mo := int(d.Hours() / 24 / 30)
		if mo < 1 {
			mo = 1
		}
		return fmt.Sprintf("%dmo ago", mo)
	}
}

// ──────────────────────────────────────────────────
// Confirm box (used by snap add success)
// ──────────────────────────────────────────────────

// RenderConfirmBox renders a success card after snippet creation.
func RenderConfirmBox(alias string, id int64, lang, tags string) string {
	headerLine := lipgloss.NewStyle().
		Background(ColorGreenDm).
		Foreground(ColorGreen).
		Bold(true).
		Padding(0, 1).
		Width(40).
		Render("✓  Snippet saved")

	keyStyle := lipgloss.NewStyle().Foreground(ColorMuted).Width(8)
	valStyle := lipgloss.NewStyle().Foreground(ColorBright)
	cyanVal := lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)

	rows := []string{
		headerLine,
		"",
		keyStyle.Render("ID") + "  " + valStyle.Render(fmt.Sprintf("%d", id)),
		keyStyle.Render("Alias") + "  " + cyanVal.Render(alias),
	}

	if lang != "" {
		rows = append(rows, keyStyle.Render("Lang")+"  "+RenderLangBadge(lang))
	}
	if tags != "" {
		rows = append(rows, keyStyle.Render("Tags")+"  "+RenderTagBadges(tags))
	} else {
		rows = append(rows, keyStyle.Render("Tags")+"  "+lipgloss.NewStyle().Foreground(ColorDim).Render("—"))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGreen).
		Width(44).
		Padding(1, 1).
		Render(strings.Join(rows, "\n"))

	return box
}
