package runamqp

import (
	"github.com/mergermarket/run-amqp/helpers"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {

	logger := helpers.NewTestLogger(t)
	numberOfJobs := 10

	poolCount := &workerPoolCount{}

	handler := &testHandler{
		workerPoolCount: poolCount,
		finished:        make(chan bool),
		maxInvocations:  numberOfJobs,
	}

	messages := make(chan Message, 10)
	maxWorkers := 2

	startWorkers(messages, handler, maxWorkers, logger)

	for i := 0; i < numberOfJobs; i++ {
		messages <- NewStubMessage("foo")
	}

	<-handler.finished

	if handler.invocations != numberOfJobs {
		t.Error("Handler was not called enough times, expect 10 but got", handler.invocations)
	}

	if poolCount.max != maxWorkers {
		t.Error("Expected max workers of", maxWorkers, "but got", poolCount.max)
	}
}

func TestWorkerPoolPanicDoesntBreakEverything(t *testing.T) {
	logger := helpers.NewTestLogger(t)
	messages := make(chan Message, 2)
	startWorkers(messages, &panicyHandler{}, 1, logger)
	messages <- NewStubMessage("Hi")

	time.Sleep(10 * time.Millisecond)
	t.Log("It didnt die, hurrah")
}

type panicyHandler struct {
}

func (*panicyHandler) Handle(_ Message) {
	panic("Oh nooooo")
}

func (*panicyHandler) Name() string {
	return "I am very jumpy"
}

type workerPoolCount struct {
	sync.Mutex
	currentRunning, max int
}

func (w *workerPoolCount) incr() {
	w.Lock()
	defer w.Unlock()

	w.currentRunning++
	if w.currentRunning >= w.max {
		w.max = w.currentRunning
	}
}

func (w *workerPoolCount) decr() {
	w.Lock()
	defer w.Unlock()

	w.currentRunning--
}

type testHandler struct {
	workerPoolCount *workerPoolCount
	sync.Mutex
	invocations, maxInvocations int
	finished                    chan bool
}

func (*testHandler) Name() string {
	return "test worker"
}

func (w *testHandler) Handle(_ Message) {
	w.workerPoolCount.incr()
	defer w.workerPoolCount.decr()

	w.Lock()
	defer w.Unlock()

	time.Sleep(10 * time.Millisecond)
	w.invocations++

	if w.invocations == w.maxInvocations {
		w.finished <- true
	}
}
