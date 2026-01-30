package exposure

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Logger records exposure events.
type Logger interface {
	Log(ctx context.Context, event Event) error
	Close() error
}

// FileLogger writes exposure events as JSON lines.
type FileLogger struct {
	mu            sync.Mutex
	file          *os.File
	path          string
	dir           string
	format        string
	fsync         bool
	clock         func() time.Time
	retentionDays int
	currentDate   string
	closed        bool
}

// FileLoggerOptions configures the file logger.
type FileLoggerOptions struct {
	Path          string
	Format        string
	Fsync         bool
	RetentionDays int
	Clock         func() time.Time
}

// NewFileLogger constructs a file-backed exposure logger.
func NewFileLogger(opts FileLoggerOptions) (*FileLogger, error) {
	path := strings.TrimSpace(opts.Path)
	if path == "" {
		return nil, errors.New("exposure log path is required")
	}
	clock := opts.Clock
	if clock == nil {
		clock = time.Now
	}
	format := normalizeLogFormat(opts.Format)
	if format == "" {
		if strings.TrimSpace(opts.Format) != "" {
			return nil, errors.New("unsupported exposure log format: " + opts.Format)
		}
		format = LogFormatServiceV1
	}
	switch format {
	case LogFormatServiceV1, LogFormatEvalV1:
	default:
		return nil, errors.New("unsupported exposure log format: " + opts.Format)
	}
	logger := &FileLogger{
		path:          path,
		format:        format,
		fsync:         opts.Fsync,
		clock:         clock,
		retentionDays: opts.RetentionDays,
	}
	if isDir(path) {
		logger.dir = path
	} else if strings.HasSuffix(path, string(os.PathSeparator)) {
		logger.dir = path
	}
	if logger.dir != "" {
		if err := os.MkdirAll(logger.dir, 0o700); err != nil {
			return nil, err
		}
	}
	return logger, nil
}

// Log appends an exposure event as a JSON line.
func (l *FileLogger) Log(ctx context.Context, event Event) error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errors.New("exposure logger closed")
	}
	now := l.clock().UTC()
	if err := l.ensureFileLocked(now); err != nil {
		return err
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = now
	}
	event.Normalize()
	if err := event.Validate(); err != nil {
		return err
	}
	payload, err := l.encodeEvent(event)
	if err != nil {
		return err
	}
	if _, err := l.file.Write(append(payload, '\n')); err != nil {
		return err
	}
	if l.fsync {
		return l.file.Sync()
	}
	return nil
}

func (l *FileLogger) encodeEvent(event Event) ([]byte, error) {
	if l == nil || l.format == LogFormatServiceV1 {
		return json.Marshal(event)
	}
	if l.format == LogFormatEvalV1 {
		exp, err := buildEvalExposure(event)
		if err != nil {
			return nil, err
		}
		return json.Marshal(exp)
	}
	return nil, errors.New("unsupported exposure log format")
}

// Close closes the underlying file handle.
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
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *FileLogger) ensureFileLocked(now time.Time) error {
	if l.dir == "" {
		if l.file == nil {
			f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
			if err != nil {
				return err
			}
			l.file = f
		}
		return nil
	}
	date := now.Format("2006-01-02")
	if l.file == nil || l.currentDate != date {
		if l.file != nil {
			_ = l.file.Close()
		}
		filename := filepath.Join(l.dir, "exposure-"+date+".jsonl")
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		l.file = f
		l.currentDate = date
		if l.retentionDays > 0 {
			l.cleanupLocked(now)
		}
	}
	return nil
}

func (l *FileLogger) cleanupLocked(now time.Time) {
	cutoff := now.AddDate(0, 0, -l.retentionDays)
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "exposure-") || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		datePart := strings.TrimSuffix(strings.TrimPrefix(name, "exposure-"), ".jsonl")
		ts, err := time.Parse("2006-01-02", datePart)
		if err != nil {
			continue
		}
		if ts.Before(cutoff) {
			_ = os.Remove(filepath.Join(l.dir, name))
		}
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
