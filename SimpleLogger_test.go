package runamqp

import (
	"bytes"
	"testing"
)

const testMsg = "Hello, world"

func TestSimpleLoggerDebug(t *testing.T) {
	var b bytes.Buffer
	logger := SimpleLogger{&b}

	t.Run("Debug", func(t *testing.T) {
		b.Reset()
		logger.Debug(testMsg)

		if b.String() != "DEBUG [Hello, world]\n" {
			t.Error("Unexpected debug output", b.String())
		}
	})

	t.Run("Info", func(t *testing.T) {
		b.Reset()
		logger.Info(testMsg)

		if b.String() != "INFO [Hello, world]\n" {
			t.Error("Unexpected debug output", b.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		b.Reset()
		logger.Error(testMsg)

		if b.String() != "ERROR [Hello, world]\n" {
			t.Error("Unexpected debug output", b.String())
		}
	})

}
