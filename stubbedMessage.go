package runamqp

import (
	"fmt"
	"time"
)

type StubMessage struct {
	message                              string
	ackCalled, nackCalled, requeueCalled chan bool
	timeout                              time.Duration
}

func (s *StubMessage) Body() []byte {
	return []byte(s.message)
}

func (s *StubMessage) Ack() error {
	s.ackCalled <- true
	return nil
}

func (s *StubMessage) Nack(reason string) error {
	s.nackCalled <- true
	return nil
}

func (s *StubMessage) Requeue(reason string) error {
	s.requeueCalled <- true
	return nil
}

func NewStubMessage(msg string, timeout time.Duration) *StubMessage {

	s := new(StubMessage)
	s.message = msg
	s.ackCalled = make(chan bool, 1)
	s.nackCalled = make(chan bool, 1)
	s.requeueCalled = make(chan bool, 1)
	s.timeout = timeout
	return s
}

func (s *StubMessage) GetAckCalled() (bool, error) {
	select {
	case acked := <-s.ackCalled:
		return acked, nil
	case <-time.After(s.timeout):
		return false, fmt.Errorf("Timed out waiting for acked")
	}
}

func (s *StubMessage) GetNackCalled() (bool, error) {
	select {
	case nacked := <-s.nackCalled:
		return nacked, nil
	case <-time.After(s.timeout):
		return false, fmt.Errorf("Timed out waiting for nack")
	}
}

func (s *StubMessage) GetRequeueCalled() (bool, error) {
	select {
	case requeued := <-s.requeueCalled:
		return requeued, nil
	case <-time.After(s.timeout):
		return false, fmt.Errorf("Timed out waiting for requeue")
	}
}
