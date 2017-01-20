package runamqp

import (
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	poolCount := &workerPoolCount{}
	handler := &testHandler{workerPoolCount:poolCount}
	messages := make(chan Message, 10)
	numberOfJobs := 10
	maxWorkers := 2

	startWorkers(messages, handler, maxWorkers, &testLogger{t})

	for i := 0; i < numberOfJobs; i++ {
		messages <- NewStubMessage("foo", 5*time.Second)
	}

	time.Sleep(1 * time.Second) // this blows

	if handler.invocations != numberOfJobs {
		t.Error("Handler was not called enough times, expect 10 but got", handler.invocations)
	}

	if poolCount.max != maxWorkers{
		t.Error("Expected max workers of", maxWorkers, "but got", poolCount.max)
	}
}

type workerPoolCount struct {
	sync.Mutex
	currentRunning, max int
}

func (w *workerPoolCount) incr() {
	w.Lock()
	defer w.Unlock()

	w.currentRunning++
	if w.currentRunning>=w.max{
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
	invocations int
}

func (*testHandler) Name() string {
	return "test worker"
}

func (w *testHandler) Handle(msg Message) {
	w.workerPoolCount.incr()
	defer w.workerPoolCount.decr()

	w.Lock()
	defer w.Unlock()

	time.Sleep(10 * time.Millisecond)
	w.invocations++
}
