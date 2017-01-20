package runamqp

import (
	"testing"
	"time"
	"sync"
)


type testHandler struct {
	sync sync.RWMutex
	invocations int
}

func (*testHandler) Name() string {
	return "test worker"
}

func (w *testHandler) Handle(msg Message) {
	w.sync.Lock()
	defer w.sync.Unlock()
	w.invocations++
}

//todo: We need to somehow test the max number of workers functionality
func TestWorkerPool(t *testing.T) {
	handler := &testHandler{}
	messages := make(chan Message, 10)
	numberOfJobs := 10

	startWorkers(messages, handler, 2, &testLogger{t})


	for i:=0; i<numberOfJobs; i++ {
		messages <- NewStubMessage("foo", 5*time.Second)
	}

	time.Sleep(1 * time.Second) // this blows

	if handler.invocations != numberOfJobs{
		t.Error("Handler was not called enough times, expect 10 but got", handler.invocations)
	}
}
