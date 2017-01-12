package tools

import (
	"testing"

	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

// Logger logs messages in a structured format in prod and pretty colours in local.
type Logger struct {
	log    *logrus.Logger
	fields logrus.Fields
}

// Info should be used to log key application events.
func (l *Logger) Info(msg ...interface{}) {
	l.log.WithFields(l.fields).Info(msg)
}

// Error should be used to log events that need to be actioned on immediately.
func (l *Logger) Error(msg ...interface{}) {
	l.log.WithFields(l.fields).Error(msg)
}

// Debug can be used to log events for local development.
func (l *Logger) Debug(msg ...interface{}) {
	l.log.WithFields(l.fields).Debug(msg)
}

// Warn is for when something bad happened but doesnt need instant action.
func (l *Logger) Warn(msg ...interface{}) {
	l.log.WithFields(l.fields).Warn(msg)
}

// NewLogger returns a new structured logger.
func NewLogger(isLocal bool) *Logger {
	logger := logrus.New()

	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		logger.Level = logrus.DebugLevel
	}

	if !isLocal {
		logger.Formatter = &logrus.JSONFormatter{}
	}

	return &Logger{
		log: logger,
		fields: logrus.Fields{
			"component": getComponentName(),
			"env":       getEnv(),
		},
	}
}

func getComponentName() string {
	if name := os.Getenv("COMPONENT_NAME"); len(name) > 0 {
		return name
	}
	return "a-service-has-no-name"
}

func getEnv() string {
	if env := os.Getenv("ENV_NAME"); len(env) > 0 {
		return env
	}
	return "local"
}

// TestLogger accepts the testing package so you wont be bombarded with logs
// when your tests pass but if they fail you will see what's going on.
type TestLogger struct {
	T *testing.T
}

// Info logs info to the test logger.
func (testLogger TestLogger) Info(msg ...interface{}) {
	testLogger.T.Log("[Info]", msg)
}

// Debug logs debug to the test logger.
func (testLogger TestLogger) Debug(msg ...interface{}) {
	testLogger.T.Log("[Debug]", msg)
}

// Error logs error to the test logger.
func (testLogger TestLogger) Error(msg ...interface{}) {
	testLogger.T.Log("[Error]", msg)
}

// Warn logs warn to the test logger.
func (testLogger TestLogger) Warn(msg ...interface{}) {
	testLogger.T.Log("[Warn]", msg)
}
