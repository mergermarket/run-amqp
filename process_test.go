package runamqp

import (
	"testing"
	"time"
)

type alwaysAckingHandler struct {
	messageRecieved string
}

func (a *alwaysAckingHandler) Handle(msg Message) {
	a.messageRecieved = string(msg.Body())
	msg.Ack()
}

func TestConsumerProcessesMessages(t *testing.T) {

	messages := make(chan Message)

	consumer := &Consumer{
		Messages: messages,
	}

	handler := &alwaysAckingHandler{}

	go consumer.Process(handler, 5)

	msg := NewStubMessage("hello, world", 5*time.Millisecond)
	messages <- msg

	time.Sleep(10 * time.Millisecond)

	if handler.messageRecieved != "hello, world" {
		t.Error("Handler was not called")
	}

}
