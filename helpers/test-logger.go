package helpers

import "testing"

type TestLogger struct {
	t *testing.T
}

func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{
		t: t,
	}
}

func (t *TestLogger) Info(items ...interface{}) {
	t.t.Log("INFO", items)
}

func (t *TestLogger) Error(items ...interface{}) {
	t.t.Log("ERROR", items)
}

func (t *TestLogger) Debug(items ...interface{}) {
	t.t.Log("DEBUG", items)
}
