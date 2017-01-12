package runamqp

import "testing"

type testLogger struct {
	t *testing.T
}

func (t *testLogger) Info(items ...interface{}) {
	t.t.Log("INFO", items)
}

func (t *testLogger) Error(items ...interface{}) {
	t.t.Log("ERROR", items)
}

func (t *testLogger) Debug(items ...interface{}) {
	t.t.Log("DEBUG", items)
}
