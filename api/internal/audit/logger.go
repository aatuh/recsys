package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

// Event captures an auditable action.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
	ActorID   string    `json:"actor_id,omitempty"`
	TenantID  string    `json:"tenant_id,omitempty"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Action    string    `json:"action"`
	Status    int       `json:"status"`
	PrevHash  string    `json:"prev_hash,omitempty"`
	Hash      string    `json:"hash,omitempty"`
}

// Logger records audit events.
type Logger interface {
	Log(ctx context.Context, event Event) error
	Close() error
}

// FileLogger writes append-only audit events with hash chaining.
type FileLogger struct {
	mu     sync.Mutex
	file   *os.File
	fsync  bool
	prev   string
	clock  func() time.Time
	closed bool
}

// FileLoggerOptions configures the file logger.
type FileLoggerOptions struct {
	Path  string
	Fsync bool
	Clock func() time.Time
}

// NewFileLogger constructs a file-backed audit logger.
func NewFileLogger(opts FileLoggerOptions) (*FileLogger, error) {
	if strings.TrimSpace(opts.Path) == "" {
		return nil, errors.New("audit log path is required")
	}
	f, err := os.OpenFile(opts.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	clock := opts.Clock
	if clock == nil {
		clock = time.Now
	}
	return &FileLogger{file: f, fsync: opts.Fsync, clock: clock}, nil
}

// Log appends an audit event and updates the hash chain.
func (l *FileLogger) Log(ctx context.Context, event Event) error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errors.New("audit logger closed")
	}
	event.Timestamp = l.clock().UTC()
	event.PrevHash = l.prev

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(append([]byte(l.prev), payload...))
	event.Hash = hex.EncodeToString(hash[:])

	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := l.file.Write(append(line, '\n')); err != nil {
		return err
	}
	if l.fsync {
		if err := l.file.Sync(); err != nil {
			return err
		}
	}
	l.prev = event.Hash
	return nil
}

// Close closes the audit log file.
func (l *FileLogger) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	return l.file.Close()
}
