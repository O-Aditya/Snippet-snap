package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/O-Aditya/snippet-snap/internal/models"
)

// InsertSnippet creates a new snippet and returns its ID.
func InsertSnippet(database *sql.DB, s *models.Snippet) (int64, error) {
	result, err := database.Exec(
		`INSERT INTO snippets (alias, content, language, tags) VALUES (?, ?, ?, ?)`,
		s.Alias, s.Content, s.Language, s.Tags,
	)
	if err != nil {
		return 0, fmt.Errorf("insert snippet: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

// GetSnippetByID retrieves a single snippet by its primary key.
func GetSnippetByID(database *sql.DB, id int64) (*models.Snippet, error) {
	row := database.QueryRow(
		`SELECT id, alias, content, language, tags, created_at, updated_at
		 FROM snippets WHERE id = ?`, id,
	)
	return scanSnippet(row)
}

// GetSnippetByAlias retrieves a single snippet by its alias.
func GetSnippetByAlias(database *sql.DB, alias string) (*models.Snippet, error) {
	row := database.QueryRow(
		`SELECT id, alias, content, language, tags, created_at, updated_at
		 FROM snippets WHERE alias = ?`, alias,
	)
	return scanSnippet(row)
}

// ListSnippets returns all snippets, optionally filtered by language and/or tag.
func ListSnippets(database *sql.DB, lang, tag string) ([]models.Snippet, error) {
	query := `SELECT id, alias, content, language, tags, created_at, updated_at FROM snippets WHERE 1=1`
	args := []interface{}{}

	if lang != "" {
		query += ` AND language = ?`
		args = append(args, lang)
	}
	if tag != "" {
		query += ` AND (',' || tags || ',') LIKE '%,' || ? || ',%'`
		args = append(args, tag)
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list snippets: %w", err)
	}
	defer rows.Close()

	return scanSnippets(rows)
}

// SearchSnippets performs a full-text search using the FTS5 index.
func SearchSnippets(database *sql.DB, term string) ([]models.Snippet, error) {
	rows, err := database.Query(
		`SELECT s.id, s.alias, s.content, s.language, s.tags, s.created_at, s.updated_at
		 FROM snippets_fts fts
		 JOIN snippets s ON s.id = fts.rowid
		 WHERE snippets_fts MATCH ?
		 ORDER BY rank`, term,
	)
	if err != nil {
		return nil, fmt.Errorf("search snippets: %w", err)
	}
	defer rows.Close()

	return scanSnippets(rows)
}

// UpdateSnippet updates a snippet's content, language, and tags.
func UpdateSnippet(database *sql.DB, s *models.Snippet) error {
	_, err := database.Exec(
		`UPDATE snippets SET content = ?, language = ?, tags = ?, updated_at = ?
		 WHERE id = ?`,
		s.Content, s.Language, s.Tags, time.Now(), s.ID,
	)
	if err != nil {
		return fmt.Errorf("update snippet: %w", err)
	}
	return nil
}

// DeleteSnippet removes a snippet by ID.
func DeleteSnippet(database *sql.DB, id int64) error {
	result, err := database.Exec(`DELETE FROM snippets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete snippet: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("snippet with id %d not found", id)
	}
	return nil
}

// scanSnippet reads a single snippet from a *sql.Row.
func scanSnippet(row *sql.Row) (*models.Snippet, error) {
	var s models.Snippet
	err := row.Scan(&s.ID, &s.Alias, &s.Content, &s.Language, &s.Tags, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan snippet: %w", err)
	}
	return &s, nil
}

// scanSnippets reads multiple snippets from *sql.Rows.
func scanSnippets(rows *sql.Rows) ([]models.Snippet, error) {
	var snippets []models.Snippet
	for rows.Next() {
		var s models.Snippet
		if err := rows.Scan(&s.ID, &s.Alias, &s.Content, &s.Language, &s.Tags, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan snippets: %w", err)
		}
		snippets = append(snippets, s)
	}
	return snippets, rows.Err()
}
