package runamqp

import (
	"github.com/mergermarket/run-amqp/connection"
	"github.com/streadway/amqp"
)

// Consumer has a channel for receiving messages
type Consumer struct {
	Messages         chan Message
	QueuesBound      chan bool
	config           ConsumerConfig
	consumerChannels *consumerChannels
}

// MessageHandler is something that can process a Message, calling Ack, nackCalls when appropiate for your domain
type MessageHandler interface {
	// Name is a description of your handler for logging purposes
	Name() string
	// Handle will receive messages as they come from AMQP
	Handle(msg Message)
}

// Process creates a worker pool of size numberOfWorkers which will run handler on every message sent to the consumer's Messages channel.
func (c *Consumer) Process(handler MessageHandler, numberOfWorkers int) {
	c.config.Logger.Debug("Stuff", handler.Name(), c.Messages)
	startWorkers(c.Messages, handler, numberOfWorkers, c.config.Logger)
}

// NewConsumer returns a Consumer
func NewConsumer(config ConsumerConfig) *Consumer {

	consumer := Consumer{
		Messages:         make(chan Message),
		QueuesBound:      make(chan bool),
		config:           config,
		consumerChannels: newConsumerChannels(config),
	}

	go consumer.setUpConnection()

	return &consumer
}

func (c *Consumer) setUpConnection() {

	connectionManager := connection.NewConnectionManager(c.config.URL, c.config.Logger)

	if !c.consumerChannels.openChannels(connectionManager) {
		c.QueuesBound <- false
		return
	}

	err := c.consumeQueue()

	if err != nil {
		c.QueuesBound <- false
		return
	}

	c.QueuesBound <- true
}

func (c *Consumer) consumeQueue() error {

	msgs, err := c.consumerChannels.mainChannel.Consume(
		c.config.queue.Name, // queue
		"",                  // consumer
		false,               // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)

	if err != nil {
		return err
	}

	c.config.Logger.Info("Queues bound, good to go")

	go func() {
		for d := range msgs {
			c.Messages <- &amqpMessage{
				delivery:          d,
				dleChannel:        c.consumerChannels.dleChannel,
				retryChannel:      c.consumerChannels.retryChannel,
				retryLimit:        c.config.queue.RetryLimit,
				retryExchangeName: c.config.exchange.RetryLater,
				dleExchangeName:   c.config.exchange.DLE,
			}
		}
	}()

	return nil
}

func assertAndBindQueue(ch *amqp.Channel, queueName, exchangeName string, patterns []string, arguments map[string]interface{}) error {
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // no-wait
		arguments, // arguments
	)

	if err != nil {
		return err
	}

	for _, pattern := range patterns {
		if err := ch.QueueBind(q.Name, pattern, exchangeName, false, nil); err != nil {
			return err
		}
	}

	return nil
}
