package prometheus

import "log"

// Logger represents internal logger
type Logger interface {
	Error(err string)
}

// DefaultLogger represents default logger instance
type DefaultLogger struct{}

func (DefaultLogger) Error(err string) {
	log.Print("[PROM]" + err)
}
