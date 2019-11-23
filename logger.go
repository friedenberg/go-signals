package signals

import (
	"fmt"
	"os"
)

type Logger interface {
	Info(v ...interface{})
	Error(v ...interface{})
}

type StandardLogger struct{}

func (l *StandardLogger) Info(v ...interface{}) {
	fmt.Fprintln(os.Stdout, v...)
}

func (l *StandardLogger) Error(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}
