package gbc

var log Logger = &nullLogger{}

// Logger used to enable debugging
type Logger interface {
	Debugf(message string, args ...interface{})
}

type nullLogger struct {
}

func (nullLogger) Debugf(message string, args ...interface{}) {}

// SetLogger assign a logger for debugging gbc
func SetLogger(l Logger) {
	log = l
}
