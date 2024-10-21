package runamqp

import (
	"fmt"
	"io"
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
	err := msg.Ack()
	if err != nil {
		// Handle error.
		return
	}
}

func (e *ExampleHandler) Name() string {
	return "Example Handler"
}

func ExampleConsumer() {

	c := NewConsumerConfig{
		URL:          testRabbitURI,
		ExchangeName: "test-example-exchange",
		ExchangeType: Fanout,
		Patterns:     noPatterns,
		Logger:       &SimpleLogger{io.Discard},
		RequeueTTL:   testRequeueTTL,
		RequeueLimit: testRequeueLimit,
		ServiceName:  serviceName,
		Prefetch:     defaultPrefetch,
	}
	// Create a consumer config
	config := c.Config()

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
	pc := NewPublisherConfig{
		URL:          config.URL,
		ExchangeName: config.exchange.Name,
		ExchangeType: config.exchange.Type,
		Confirmable:  false,
		Logger:       config.Logger,
	}
	publisherConfig := pc.Config()
	publisher, err := NewPublisher(publisherConfig)

	// Let's check the Publisher is ready too
	if err != nil {
		log.Fatal("Problem making publisher", publisher)
	}

	// Publish a message
	if err := publisher.Publish([]byte("Hello, world"), nil); err != nil {
		log.Fatal("Error when Publishing the message")
	}

	// Wait a little bit!
	time.Sleep(5 * time.Millisecond)

	fmt.Print(handler.calledWith)
	// Output: Hello, world
}
