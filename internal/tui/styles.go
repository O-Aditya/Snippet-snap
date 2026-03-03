package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ──────────────────────────────────────────────────
// Foreground-only palette — no background assumptions
// The user's terminal background is sacred.
// ──────────────────────────────────────────────────

var (
	// Three levels of text brightness
	ColorBright = lipgloss.Color("#E6EDF3") // active, selected, important
	ColorNormal = lipgloss.Color("#9198A1") // body text, readable
	ColorDim    = lipgloss.Color("#636E7B") // metadata, hints, decorators

	// Structure
	ColorBorder = lipgloss.Color("#3D444D") // borders and dividers

	// Single accent
	ColorAccent = lipgloss.Color("#39D0D8") // cyan — THE ONLY bright color

	// Semantic — used in messages only, never decoration
	ColorGreen  = lipgloss.Color("#3FB950")
	ColorRed    = lipgloss.Color("#F85149")
	ColorAmber  = lipgloss.Color("#CBA135")
	ColorPurple = lipgloss.Color("#B48EFF")
	ColorBlue   = lipgloss.Color("#58A6FF")
)

// Backgrounds — used sparingly, only where justified.
var (
	BgSelected  = lipgloss.Color("#1F2937") // selected row only
	BgWordmark  = lipgloss.Color("#39D0D8") // ◈ SNIPPET-SNAP pill only
	BgStatusBar = lipgloss.Color("#161B22") // bottom status strip
	BgInput     = lipgloss.Color("#161B22") // search input box
	BgBadgeCyan = lipgloss.Color("#163940") // focused input border tint
	BgSuccess   = lipgloss.Color("#0D2218") // confirm box header tint
	BgError     = lipgloss.Color("#2A0F0E") // error box tint
	BgKeyBadge  = lipgloss.Color("#1C2128") // keyboard shortcut pill
)

// ──────────────────────────────────────────────────
// Lang badge color pairs — small pills, bg is acceptable
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

// ──────────────────────────────────────────────────
// Composite styles — minimal, foreground-first
// ──────────────────────────────────────────────────

var (
	BrightStyle = lipgloss.NewStyle().Foreground(ColorBright)
	NormalStyle = lipgloss.NewStyle().Foreground(ColorNormal)
	DimStyle    = lipgloss.NewStyle().Foreground(ColorDim)
	AccentStyle = lipgloss.NewStyle().Foreground(ColorAccent)
	BorderStyle = lipgloss.NewStyle().Foreground(ColorBorder)

	SearchPromptStyle = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
)

// ──────────────────────────────────────────────────
// Component renderers
// ──────────────────────────────────────────────────

// RenderWordmark renders the ◈  SNIPPET-SNAP pill — the one bg element.
func RenderWordmark() string {
	return lipgloss.NewStyle().
		Background(BgWordmark).
		Foreground(lipgloss.Color("#0D1117")).
		Bold(true).
		Padding(0, 2).
		Render("◈  SNIPPET-SNAP")
}

// RenderLangBadge renders a per-language colored pill.
func RenderLangBadge(lang string) string {
	if lang == "" {
		return ""
	}
	lc, ok := langColors[strings.ToLower(lang)]
	if !ok {
		lc = langColor{bg: lipgloss.Color("#1C2128"), fg: lipgloss.Color("#636E7B")}
	}
	return lipgloss.NewStyle().
		Background(lc.bg).
		Foreground(lc.fg).
		Bold(true).
		Padding(0, 1).
		Render(strings.ToUpper(lang))
}

// RenderTagBadge renders a single tag pill.
func RenderTagBadge(tag string) string {
	if tag == "" {
		return ""
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#091D36")).
		Foreground(ColorBlue).
		Padding(0, 1).
		Render(tag)
}

// RenderTagBadges splits comma-separated tags into pills.
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

// RenderKey renders a keyboard shortcut pill.
func RenderKey(k string) string {
	return lipgloss.NewStyle().
		Background(BgKeyBadge).
		Foreground(ColorBright).
		Bold(true).
		Padding(0, 1).
		Render(k)
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

// PrintInfo prints a dim info message.
func PrintInfo(msg string) {
	fmt.Println(DimStyle.Render("ℹ " + msg))
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
// Confirm box — the ONE justified bordered element
// ──────────────────────────────────────────────────

// RenderConfirmBox renders a success card after snippet creation.
func RenderConfirmBox(alias string, id int64, lang, tags string) string {
	headerLine := lipgloss.NewStyle().
		Background(BgSuccess).
		Foreground(ColorGreen).
		Bold(true).
		Padding(0, 1).
		Width(40).
		Render("✓  Snippet saved")

	keyStyle := DimStyle.Width(8)
	valStyle := BrightStyle

	rows := []string{
		headerLine,
		"",
		keyStyle.Render("ID") + "  " + valStyle.Render(fmt.Sprintf("%d", id)),
		keyStyle.Render("Alias") + "  " + lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render(alias),
	}

	if lang != "" {
		rows = append(rows, keyStyle.Render("Lang")+"  "+RenderLangBadge(lang))
	}
	if tags != "" {
		rows = append(rows, keyStyle.Render("Tags")+"  "+RenderTagBadges(tags))
	} else {
		rows = append(rows, keyStyle.Render("Tags")+"  "+DimStyle.Render("—"))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGreen).
		Width(44).
		Padding(1, 1).
		Render(strings.Join(rows, "\n"))

	return box
}
