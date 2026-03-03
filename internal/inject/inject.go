package inject

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// varRegex matches {{VAR_NAME}} placeholders where name starts with
// an uppercase letter followed by uppercase letters, digits, or underscores.
var varRegex = regexp.MustCompile(`\{\{([A-Z][A-Z0-9_]*)\}\}`)

// Local styles — no tui import to avoid circular dependency.
var (
	promptDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("#636E7B"))
	promptAccent = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA135")).Bold(true)
	promptErr    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149")).Bold(true)
)

// FindVars returns a deduplicated, ordered list of variable names found in content.
// Returns nil if no vars are found. Preserves first-occurrence order.
func FindVars(content string) []string {
	matches := varRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var vars []string
	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			vars = append(vars, name)
		}
	}
	return vars
}

// ResolveVars scans content for {{VAR}} placeholders, prompts the user for
// each unique variable on stderr/stdin, and returns the content with all
// placeholders replaced. If no vars are found, returns content unchanged.
func ResolveVars(content string) (string, error) {
	vars := FindVars(content)
	if vars == nil {
		return content, nil
	}

	fmt.Fprintln(os.Stderr,
		promptDim.Render("  resolving ")+
			promptAccent.Render(fmt.Sprintf("%d", len(vars)))+
			promptDim.Render(" variable(s)"))

	reader := bufio.NewReader(os.Stdin)
	values := make(map[string]string, len(vars))

	for _, name := range vars {
		fmt.Fprint(os.Stderr, promptAccent.Render("  "+name+" › "))

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("input cancelled")
		}
		value := strings.TrimSpace(line)

		// Re-prompt once on empty input
		if value == "" {
			fmt.Fprint(os.Stderr, promptErr.Render("  (required) ")+promptAccent.Render(name+" › "))
			line, err = reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("input cancelled")
			}
			value = strings.TrimSpace(line)
			if value == "" {
				return "", fmt.Errorf("aborted: %s is required", name)
			}
		}

		values[name] = value
	}

	resolved := content
	for name, value := range values {
		resolved = strings.ReplaceAll(resolved, "{{"+name+"}}", value)
	}

	return resolved, nil
}
