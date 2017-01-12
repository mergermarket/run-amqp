package tools

import "testing"

// NewTestTools returns a logger and statsd instance to use in your tests
func NewTestTools(t *testing.T) (logger, StatsD) {
	logger := TestLogger{T: t}
	statsd, _ := NewStatsD(NewStatsDConfig(false, logger))
	return logger, statsd
}