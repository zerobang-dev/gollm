package logger

import (
	"time"
)

// Query represents a logged query
type Query struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Prompt      string    `json:"prompt"`
	Model       string    `json:"model"`
	Response    string    `json:"response"`
	Duration    int64     `json:"duration_ms"`
	Temperature float64   `json:"temperature"`
}
