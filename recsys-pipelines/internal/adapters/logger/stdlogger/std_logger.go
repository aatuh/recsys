package stdlogger

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
)

type StdLogger struct {
	mu  sync.Mutex
	l   *log.Logger
	lvl level
}

type level int

const (
	levelDebug level = iota
	levelInfo
	levelWarn
	levelError
)

type Option func(*StdLogger)

func WithLevelDebug() Option { return func(s *StdLogger) { s.lvl = levelDebug } }
func WithLevelInfo() Option  { return func(s *StdLogger) { s.lvl = levelInfo } }

func New(opts ...Option) *StdLogger {
	s := &StdLogger{
		l:   log.New(os.Stdout, "", log.LstdFlags),
		lvl: levelInfo,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *StdLogger) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	s.log(levelDebug, "DEBUG", msg, fields...)
}

func (s *StdLogger) Info(ctx context.Context, msg string, fields ...logger.Field) {
	s.log(levelInfo, "INFO", msg, fields...)
}

func (s *StdLogger) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	s.log(levelWarn, "WARN", msg, fields...)
}

func (s *StdLogger) Error(ctx context.Context, msg string, fields ...logger.Field) {
	s.log(levelError, "ERROR", msg, fields...)
}

func (s *StdLogger) log(lvl level, prefix, msg string, fields ...logger.Field) {
	if lvl < s.lvl {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(fields) == 0 {
		s.l.Printf("%s %s", prefix, msg)
		return
	}
	kv := ""
	for i, f := range fields {
		if i > 0 {
			kv += " "
		}
		kv += fmt.Sprintf("%s=%v", f.Key, f.Value)
	}
	s.l.Printf("%s %s %s", prefix, msg, kv)
}
