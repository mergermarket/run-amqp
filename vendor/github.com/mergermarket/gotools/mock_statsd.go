package tools

import (
	"errors"
	"fmt"
	"testing"
)

// MockStatsD provides a basic mock of the MMG StatsD object. It takes a *testing.T to assert on.
type MockStatsD struct {
	calls []Call
}

// Call is a single call to a StatsD method. It has the method name and the arguments it was called with
type Call struct {
	Method string
	Args   Args
}

// Args are the list of arguments to a single StatsD method
type Args struct {
	Name  string
	Value float64
	Tags  []string
}

// MockStatsDExpectation is called whenever a MockStatsD method is invoked. It will receive the
// pointer to the test object added as Test to MockStatsD, the name of the method called, and
// and the four arguments that a StatsD metric may receive. Assert against these values in the
// body of the function
type MockStatsDExpectation func(t *testing.T, method, name string, value float64, tags []string)

// Histogram is a mock histogrm method
func (msd *MockStatsD) Histogram(name string, value float64, tags ...string) {
	msd.call("Histogram", name, value, tags)
}

// Gauge is a mock histogrm method
func (msd *MockStatsD) Gauge(name string, value float64, tags ...string) {
	msd.call("Gauge", name, value, tags)
}

// Incr is a mock histogrm method
func (msd *MockStatsD) Incr(name string, tags ...string) {
	msd.call("Incr", name, 0, tags)
}

func (msd *MockStatsD) Call() (c Call, err error) {
	fmt.Println(msd.calls)
	if len(msd.calls) == 0 {
		return c, errors.New("No calls made")
	}
	return msd.calls[0], nil
}

func (msd *MockStatsD) call(method string, name string, value float64, tags []string) error {
	msd.calls = append(msd.calls, Call{method, Args{name, value, tags}})
	return nil
}
