package runamqp

import "fmt"

// StubMessage should be used for your tests to stub out a message coming into your system.
type StubMessage interface {
	// Body - returns the message you passed into NewStubMessage
	Body() []byte

	// Ack - acknowledge the message and return and error on failure
	Ack() error

	// Nack - sends the message to the Deadletter exchange with reason and returns error on failure
	Nack(reason string) error

	// Requeue - re-queues the message with reason and returns error on failure
	Requeue(reason string) error

	// AckCalled- returns true if Ack is called for the first time
	AckCalled() bool

	// NackCalled - returns true when Nack is only called once and the message is not previously Acked or Requeued
	NackCalled() bool

	// NackedWith - returns true if Nack was called with expected value
	NackedWith(expectedValue string) bool

	// RequeueCalled - returns true when Requeue was called only once and the message is not previously Acked or Nacked
	RequeueCalled() bool

	// RequeuedWith - returns true if Requeue was called with expectedValue
	RequeuedWith(expectedValue string) bool
}

// stubMessage should be used for your tests to stub out a message coming into your system.
type stubMessage struct {
	message string
	calls   *stubMessageCalls
}

// stubMessageCalls records message calls such as Ack
type stubMessageCalls struct {
	ackCalled     bool
	nackCalled    bool
	nackReason    string
	requeueCalled bool
	requeueReason string
}

// Body returns the message you passed into NewStubMessage
func (s *stubMessage) Body() []byte {
	return []byte(s.message)
}

// Ack ...
func (s *stubMessage) Ack() error {
	if s.calls.requeueCalled {
		return fmt.Errorf("you cannot Ack a message that is previously been Requeued")
	}
	if s.calls.nackCalled {
		return fmt.Errorf("you cannot Ack a message that is previously been Nacked")
	}
	if s.calls.ackCalled {
		return fmt.Errorf("you cannot Ack a message that is previously been Acked")
	}
	s.calls.ackCalled = true
	return nil
}

// nackCalls ...
func (s *stubMessage) Nack(reason string) error {
	if s.calls.ackCalled {
		return fmt.Errorf("you cannot Nack a message that is previously been Acked")
	}
	if s.calls.requeueCalled {
		return fmt.Errorf("you cannot Nack a message that is previously been Requeued")
	}
	if s.calls.nackCalled {
		return fmt.Errorf("you cannot Nack a message that is previously been Nacked")
	}

	s.calls.nackCalled = true
	s.calls.nackReason = reason
	return nil
}

// requeueCalls will obviously not "really" requeue!
func (s *stubMessage) Requeue(reason string) error {
	if s.calls.ackCalled {
		return fmt.Errorf("you cannot Requeue a message that is previously been Acked")
	}
	if s.calls.nackCalled {
		return fmt.Errorf("you cannot Requeue a message that is previously been Nacked")
	}
	if s.calls.requeueCalled {
		return fmt.Errorf("you cannot Requeue a message that is previously been Re-queued")
	}

	s.calls.requeueCalled = true
	s.calls.requeueReason = reason
	return nil
}

// AckCalled returns true if Ack is called successfully on the message
func (s *stubMessage) AckCalled() bool {
	return s.calls.ackCalled
}

// NackCalled returns true when nack was called once on this message
func (s *stubMessage) NackCalled() bool {
	return s.calls.nackCalled
}

// NackedWith returns true when nack was called with given value
func (s *stubMessage) NackedWith(expectedValue string) bool {
	return s.calls.nackCalled && s.calls.nackReason == expectedValue
}

// RequeueCalled returns true when requeue was called once on this message
func (s *stubMessage) RequeueCalled() bool {
	return s.calls.requeueCalled
}

// RequeuedWith returns true when requeue was called with given value
func (s *stubMessage) RequeuedWith(expectedValue string) bool {
	return s.calls.requeueCalled && s.calls.requeueReason == expectedValue
}

// NewStubMessage returns you a stubMessage. It has methods to help you make assertions on how your program interacts with a message
func NewStubMessage(msg string) StubMessage {
	s := new(stubMessage)
	s.message = msg
	s.calls = &stubMessageCalls{}
	return s
}
