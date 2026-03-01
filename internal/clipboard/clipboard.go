package clipboard

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// Copy writes the given text to the system clipboard.
func Copy(text string) error {
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("clipboard copy: %w", err)
	}
	return nil
}

// Read returns the current contents of the system clipboard.
func Read() (string, error) {
	text, err := clipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("clipboard read: %w", err)
	}
	return text, nil
}
