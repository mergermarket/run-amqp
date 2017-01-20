package runamqp

import (
	"fmt"
	"io"
)

// SimpleLogger is a simple implementation of Logger which sends messages to a file
type SimpleLogger struct {
	Out io.Writer
}

// Info writes INFO and messages to stdout
func (s *SimpleLogger) Info(x ...interface{}) {
	fmt.Fprintln(s.Out, "INFO  ", x)
}

// Error writes ERROR and messages to stdout
func (s *SimpleLogger) Error(x ...interface{}) {
	fmt.Fprintln(s.Out, "ERROR ", x)
}

// Debug writes DEBUG and messages to stdout
func (s *SimpleLogger) Debug(x ...interface{}) {
	fmt.Fprintln(s.Out, "DEBUG ", x)
}
