package runamqp

// StubMessage should be used for your tests to stub out a message coming into your system.
type StubMessage struct {
	message string
	Calls   *StubMessageCalls
}

// StubMessageCalls records message calls such as Ack
type StubMessageCalls struct {
	AckCount int
	Nack     []string
	Requeue  []string
}

// Body returns the message you passed into NewStubMessage
func (s *StubMessage) Body() []byte {
	return []byte(s.message)
}

// Ack ...
func (s *StubMessage) Ack() error {
	s.Calls.AckCount++
	return nil
}

// Nack ...
func (s *StubMessage) Nack(reason string) error {
	s.Calls.Nack = append(s.Calls.Nack, reason)
	return nil
}

// Requeue will obviously not "really" requeue!
func (s *StubMessage) Requeue(reason string) error {
	s.Calls.Requeue = append(s.Calls.Requeue, reason)
	return nil
}

// AckedOnce returns true when ack was called once on this message
func (s *StubMessage) AckedOnce() bool {
	return s.Calls.AckCount == 1
}

// NackedOnce returns true when nack was called once on this message
func (s *StubMessage) NackedOnce() bool {
	return len(s.Calls.Nack) == 1
}

// NewStubMessage returns you a StubMessage. The timeout is for the Get* methods which check to see if functions are called in an asynchonous environment
func NewStubMessage(msg string) *StubMessage {
	s := new(StubMessage)
	s.message = msg
	s.Calls = &StubMessageCalls{}
	return s
}
