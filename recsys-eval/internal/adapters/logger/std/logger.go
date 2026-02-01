package std

import "log"

// Logger logs to stdout/stderr using the standard logger.
type Logger struct{}

func (Logger) Infof(format string, args ...any) {
	log.Printf("INFO: "+format, args...)
}

func (Logger) Errorf(format string, args ...any) {
	log.Printf("ERROR: "+format, args...)
}
