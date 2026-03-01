package models

import (
	"fmt"
	"strings"
	"time"
)

// Snippet represents a saved code snippet with metadata.
type Snippet struct {
	ID        int64
	Alias     string
	Content   string
	Language  string
	Tags      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks that the snippet has required fields populated.
func (s *Snippet) Validate() error {
	s.Alias = strings.TrimSpace(s.Alias)
	s.Content = strings.TrimSpace(s.Content)

	if s.Alias == "" {
		return fmt.Errorf("snippet alias cannot be empty")
	}
	if s.Content == "" {
		return fmt.Errorf("snippet content cannot be empty")
	}
	return nil
}

// TagList returns the tags split into a slice.
func (s *Snippet) TagList() []string {
	if s.Tags == "" {
		return nil
	}
	parts := strings.Split(s.Tags, ",")
	tags := make([]string, 0, len(parts))
	for _, t := range parts {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
