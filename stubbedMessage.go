package runamqp

import (
	"fmt"
	"time"
)

// StubMessage should be used for your tests to stub out a message coming into your system.
type StubMessage struct {
	message                              string
	ackCalled, nackCalled, requeueCalled chan bool
	timeout                              time.Duration
}

// Body returns the message you passed into NewStubMessage
func (s *StubMessage) Body() []byte {
	return []byte(s.message)
}

// Ack ...
func (s *StubMessage) Ack() error {
	s.ackCalled <- true
	return nil
}

// Nack ...
func (s *StubMessage) Nack(reason string) error {
	s.nackCalled <- true
	return nil
}

// Requeue will obviously not "really" requeue!
func (s *StubMessage) Requeue(reason string) error {
	s.requeueCalled <- true
	return nil
}

// NewStubMessage returns you a StubMessage. The timeout is for the Get* methods which check to see if functions are called in an asynchonous environment
func NewStubMessage(msg string, timeout time.Duration) *StubMessage {

	s := new(StubMessage)
	s.message = msg
	s.ackCalled = make(chan bool, 1)
	s.nackCalled = make(chan bool, 1)
	s.requeueCalled = make(chan bool, 1)
	s.timeout = timeout
	return s
}

// GetAckCalled tells you if Ack was called within the timeout you set in NewStubMessage. This is blocking
func (s *StubMessage) GetAckCalled() (bool, error) {
	select {
	case acked := <-s.ackCalled:
		return acked, nil
	case <-time.After(s.timeout):
		return false, fmt.Errorf("Timed out waiting for acked")
	}
}

// GetNackCalled tells you if Nack was called within the timeout you set in NewStubMessage. This is blocking
func (s *StubMessage) GetNackCalled() (bool, error) {
	select {
	case nacked := <-s.nackCalled:
		return nacked, nil
	case <-time.After(s.timeout):
		return false, fmt.Errorf("Timed out waiting for nack")
	}
}

// GetRequeueCalled tells you if Requeue was called within the timeout you set in NewStubMessage. This is blocking
func (s *StubMessage) GetRequeueCalled() (bool, error) {
	select {
	case requeued := <-s.requeueCalled:
		return requeued, nil
	case <-time.After(s.timeout):
		return false, fmt.Errorf("Timed out waiting for requeue")
	}
}
