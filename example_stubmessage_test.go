package runamqp

import (
	"fmt"
)

// ExampleHandlerToTest is to show how you can test your handler using StubMessage
type ExampleHandlerToTest struct{}

// Handle is how you implement the MessageHandler interface, what you do with it is up to you
func (e *ExampleHandlerToTest) Handle(msg Message) {
	msg.Ack()
}

func (e *ExampleHandlerToTest) Name() string {
	return "Always acking handler"
}

func ExampleStubMessage() {

	// Create the thing you want to test
	handler := &ExampleHandlerToTest{}

	// An example message you want to test your system against
	msg := NewStubMessage("Some payload")

	// Run your handler
	handler.Handle(msg)

	// Check it did what you want
	if msg.AckedOnce() {
		fmt.Print("It acked just like we expected")
	}

	// Output: It acked just like we expected
}
