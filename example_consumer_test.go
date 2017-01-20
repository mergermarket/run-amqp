package runamqp

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

// Example handler is the sort of thing you'll make for your application to process amqp messages
type ExampleHandler struct {
	calledWith string
}

// Handle is how you implement the MessageHandler interface, what you do with it is up to you
func (e *ExampleHandler) Handle(msg Message) {
	e.calledWith = string(msg.Body())
	msg.Ack()
}

func (e *ExampleHandler) Name() string {
	return "Example Handler"
}

func ExampleConsumer() {

	// Create a consumer config
	config := NewConsumerConfig(
		testRabbitURI,
		"test-example-exchange",
		Fanout,
		noPatterns,
		&SimpleLogger{ioutil.Discard},
		testRequeueTTL,
		testRequeueLimit,
		serviceName,
	)

	// Create a consumer, which holds the references to the channel of Messages
	consumer := NewConsumer(config)

	// It's good practice to set a timeout in case we have trouble connecting and configuring rabbit
	select {
	case <-consumer.QueuesBound:
	case <-time.After(10 * time.Second):
		log.Fatal("Timed out waiting to set up rabbit")
	}

	// Create a handler for messages
	handler := &ExampleHandler{}

	// Tell the consumer to process messages using your handler
	numberOfWorkers := 10
	consumer.Process(handler, numberOfWorkers)

	// We can now publish to the same exchange for fun
	publisherConfig := NewPublisherConfig(config.URL, config.exchange.Name, config.exchange.Type, config.Logger)
	publisher := NewPublisher(publisherConfig)

	// Let's check the Publisher is ready too
	select {
	case <-publisher.PublishReady:
	case <-time.After(10 * time.Second):
		log.Fatal("Timed out waiting to set up rabbit")
	}

	// Publish a message
	if err := publisher.Publish([]byte("Hello, world"), ""); err != nil {
		log.Fatal("Error when Publishing the message")
	}

	// Wait a little bit!
	time.Sleep(5 * time.Millisecond)

	fmt.Print(handler.calledWith)
	// Output: Hello, world
}
