package runamqp

import (
	"github.com/mergermarket/run-amqp/helpers"
	"testing"
	"time"
)

type alwaysAckingHandler struct {
	messageRecieved string
}

func (a *alwaysAckingHandler) Handle(msg Message) {
	a.messageRecieved = string(msg.Body())
	_ = msg.Ack()
}

func (a *alwaysAckingHandler) Name() string {
	return "Always acking handler"
}

func TestConsumerProcessesMessages(t *testing.T) {

	messages := make(chan Message)
	logger := helpers.NewTestLogger(t)

	consumer := &Consumer{
		Messages: messages,
		config: ConsumerConfig{
			connectionConfig: connectionConfig{
				Logger: logger,
			},
		},
	}

	handler := &alwaysAckingHandler{}

	consumer.Process(handler, 5)

	msg := NewStubMessage("hello, world")
	messages <- msg

	time.Sleep(10 * time.Millisecond)

	if handler.messageRecieved != "hello, world" {
		t.Error("Handler was not called")
	}

}
