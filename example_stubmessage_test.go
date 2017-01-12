package runamqp

import (
	"fmt"
	"log"
	"time"
)

// ExampleHandlerToTest is to show how you can test your handler using StubMessage
type ExampleHandlerToTest struct{}

// Handle is how you implement the MessageHandler interface, what you do with it is up to you
func (e *ExampleHandlerToTest) Handle(msg Message) {
	msg.Ack()
}

func ExampleStubMessage() {

	// Create the thing you want to test
	handler := &ExampleHandler{}

	// An example message you want to test your system against
	msg := NewStubMessage("Some payload", 50*time.Millisecond)

	// Run your handler
	handler.Handle(msg)

	// Check it did what you want
	ackCalled, timeout := msg.GetAckCalled()

	if timeout != nil {
		log.Fatal(timeout)
	}

	if ackCalled {
		fmt.Print("It acked just like we expected")
	}

	// Output: It acked just like we expected
}
