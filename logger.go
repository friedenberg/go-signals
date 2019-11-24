package signals

import (
	"fmt"
	"os"
	"testing"
)

type Logger interface {
	Info(v ...interface{})
	Error(v ...interface{})
}

type NoOpLogger struct{}

func (l *NoOpLogger) Info(v ...interface{})  {}
func (l *NoOpLogger) Error(v ...interface{}) {}

type TestingLogger struct {
	t *testing.T
}

func (l *TestingLogger) Info(v ...interface{}) {
	l.t.Log(v...)
}

func (l *TestingLogger) Error(v ...interface{}) {
	l.t.Log(v...)
}

type StandardLogger struct{}

func (l *StandardLogger) Info(v ...interface{}) {
	fmt.Fprintln(os.Stdout, v...)
}

func (l *StandardLogger) Error(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}
