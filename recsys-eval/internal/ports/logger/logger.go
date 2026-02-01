package logger

// Logger is a minimal logging interface.
type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}
