package inject

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// varPattern matches {{VAR_NAME}} placeholders where VAR_NAME starts with
// an uppercase letter followed by uppercase letters, digits, or underscores.
var varPattern = regexp.MustCompile(`\{\{([A-Z][A-Z0-9_]*)\}\}`)

// FindVars returns a deduplicated, ordered list of variable names found in content.
func FindVars(content string) []string {
	matches := varPattern.FindAllStringSubmatch(content, -1)
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
// each unique variable, and returns the content with all placeholders replaced.
func ResolveVars(content string) (string, error) {
	vars := FindVars(content)
	if len(vars) == 0 {
		return content, nil
	}

	scanner := bufio.NewScanner(os.Stdin)
	resolved := content

	for _, name := range vars {
		fmt.Printf("  %s: ", name)
		if !scanner.Scan() {
			return "", fmt.Errorf("input cancelled")
		}
		value := strings.TrimSpace(scanner.Text())
		placeholder := "{{" + name + "}}"
		resolved = strings.ReplaceAll(resolved, placeholder, value)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}

	return resolved, nil
}

// ResolveVarsWithMap replaces all {{VAR}} placeholders using the provided map.
// This is useful for non-interactive injection (e.g., from the TUI).
func ResolveVarsWithMap(content string, values map[string]string) string {
	for name, value := range values {
		placeholder := "{{" + name + "}}"
		content = strings.ReplaceAll(content, placeholder, value)
	}
	return content
}
