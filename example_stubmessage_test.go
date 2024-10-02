package runamqp

import (
	"fmt"
)

// ExampleHandlerToTest is to show how you can test your handler using stubMessage
type ExampleHandlerToTest struct{}

// Handle is how you implement the MessageHandler interface, what you do with it is up to you
func (e *ExampleHandlerToTest) Handle(msg Message) {

	message := string(msg.Body())

	if message == "AckMessage" {
		err := msg.Ack()
		if err != nil {
			// Handle error
			return
		}
		return
	}

	if message == "NackMessage" {
		reasonForNach := "nackCalls demo!"
		err := msg.Nack(reasonForNach)
		if err != nil {
			// Handle error.
			return
		}
		return
	}

	if message == "RequeueMessage" {
		reasonForRequeue := "requeueCalls demo!"
		err := msg.Requeue(reasonForRequeue)
		if err != nil {
			// Handle error.
			return
		}
		return
	}

}

func (e *ExampleHandlerToTest) Name() string {
	return "Always acking handler"
}

func ExampleStubMessage() {

	// Create the thing you want to test

	handler := &ExampleHandlerToTest{}

	var stubMessage StubMessage

	message := "AckMessage"

	stubMessage = NewStubMessage(message)
	handler.Handle(stubMessage)

	// Check Acked is called once
	if stubMessage.AckCalled() {
		fmt.Print("It Acked just like we expected")
	}

	message = "NackMessage"

	stubMessage = NewStubMessage(message)
	handler.Handle(stubMessage)

	// Check nackCalls is called once
	if stubMessage.NackCalled() {
		fmt.Print("It Nacked just like we expected")
	}

	// Check it Nacked the expected message
	if stubMessage.NackedWith(message) {
		fmt.Print("It Nacked the expected message")
	}

	message = "RequeueMessage"
	stubMessage = NewStubMessage(message)
	handler.Handle(stubMessage)

	// Check requeueCalls is called once
	if stubMessage.RequeueCalled() {
		fmt.Print("It Requeued just like we expected")
	}

	// Check it Requeued the expected message
	if stubMessage.RequeuedWith(message) {
		fmt.Print("It Requeued the expected message")
	}

}
