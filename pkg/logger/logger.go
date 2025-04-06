package logger

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Logger handles query logging
type Logger struct {
	db *sql.DB
}

// NewLogger creates a new query logger
func NewLogger(configDir string) (*Logger, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	dbPath := filepath.Join(configDir, "queries.db")

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS queries (
			id TEXT PRIMARY KEY,
			timestamp TEXT NOT NULL,
			prompt TEXT NOT NULL,
			model TEXT NOT NULL,
			response TEXT,
			duration_ms INTEGER,
			temperature REAL
		)
	`)
	if err != nil {
		if err := db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing database: %v\n", err)
		}
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Logger{db: db}, nil
}

// Close closes the database connection
func (l *Logger) Close() error {
	if l.db != nil {
		return l.db.Close()
	}
	return nil
}

// LogQuery logs a query to the database
func (l *Logger) LogQuery(prompt, model, response string, duration time.Duration, temperature float64) error {
	// Generate a unique ID
	id := uuid.New().String()

	// Insert query record
	_, err := l.db.Exec(
		"INSERT INTO queries (id, timestamp, prompt, model, response, duration_ms, temperature) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, time.Now().Format(time.RFC3339), prompt, model, response, duration.Milliseconds(), temperature,
	)

	if err != nil {
		return fmt.Errorf("failed to log query: %w", err)
	}

	return nil
}

// GetRecentQueries retrieves recent queries
func (l *Logger) GetRecentQueries(limit int) ([]Query, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := l.db.Query(
		"SELECT id, timestamp, prompt, model, response, duration_ms, temperature FROM queries ORDER BY timestamp DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch queries: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing rows: %v\n", err)
		}
	}()

	var queries []Query
	for rows.Next() {
		var q Query
		var timestamp string

		err := rows.Scan(&q.ID, &timestamp, &q.Prompt, &q.Model, &q.Response, &q.Duration, &q.Temperature)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse timestamp
		q.Timestamp, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		queries = append(queries, q)
	}

	return queries, nil
}

// SearchQueries searches for queries containing the given text
func (l *Logger) SearchQueries(text string, limit int) ([]Query, error) {
	if limit <= 0 {
		limit = 10
	}

	// Create search pattern for SQLite LIKE operator
	pattern := "%" + text + "%"

	// Query with search criteria
	rows, err := l.db.Query(
		`SELECT id, timestamp, prompt, model, response, duration_ms, temperature 
		FROM queries 
		WHERE prompt LIKE ? OR response LIKE ?
		ORDER BY timestamp DESC LIMIT ?`,
		pattern, pattern, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search queries: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing rows: %v\n", err)
		}
	}()

	var queries []Query
	for rows.Next() {
		var q Query
		var timestamp string

		err := rows.Scan(&q.ID, &timestamp, &q.Prompt, &q.Model, &q.Response, &q.Duration, &q.Temperature)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse timestamp
		q.Timestamp, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		queries = append(queries, q)
	}

	return queries, nil
}