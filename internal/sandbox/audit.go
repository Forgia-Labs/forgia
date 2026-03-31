package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditEntry records a single command execution in the sandbox.
type AuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Command     string    `json:"command"`
	ExitCode    int       `json:"exit_code"`
	DurationMs  int64     `json:"duration_ms"`
	OutputBytes int       `json:"output_bytes"`
}

// AuditLogger writes command audit entries to a JSONL file.
type AuditLogger struct {
	file *os.File
	mu   sync.Mutex
}

// NewAuditLogger creates a logger writing to the given path.
func NewAuditLogger(path string) (*AuditLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open audit log: %w", err)
	}
	return &AuditLogger{file: f}, nil
}

// Log writes an audit entry.
func (a *AuditLogger) Log(entry AuditEntry) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}
	data = append(data, '\n')

	_, err = a.file.Write(data)
	return err
}

// Close flushes and closes the audit log.
func (a *AuditLogger) Close() error {
	return a.file.Close()
}
