package runamqp

import (
	"github.com/mergermarket/run-amqp/connection"
	"github.com/streadway/amqp"
	"sync"
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

	mainQueueReady := make(chan bool)
	dleQueueReady := make(chan bool)
	retryQueueReady := make(chan bool)

	go c.isExchangeWithQueueReady(connectionManager, mainQueueReady, c.consumerChannels.setUpMainExchangeWithQueue, c.config.queue.Name)
	go c.isExchangeWithQueueReady(connectionManager, dleQueueReady, c.consumerChannels.setUpDeadLetterExchangeWithQueue, c.config.queue.DLQ)
	go c.isExchangeWithQueueReady(connectionManager, retryQueueReady, c.consumerChannels.setUpRetryExchangeWithQueue, c.config.queue.RetryLater)

	isReady := allQueuesReady(mainQueueReady, dleQueueReady, retryQueueReady)

	if !isReady {
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

func allQueuesReady(signals ...<-chan bool) bool {
	var wg sync.WaitGroup
	isAnyUnSuccessful := false
	wg.Add(len(signals))

	for _, s := range signals {
		go func(signal <-chan bool) {
			defer wg.Done()
			if success := <-signal; !success {
				isAnyUnSuccessful = true
			}
		}(s)
	}

	wg.Wait()

	return !isAnyUnSuccessful

}

func (c *Consumer) isExchangeWithQueueReady(connectionManager connection.ConnectionManager, isReady chan bool, setUpExchangeWithQueue func(*amqp.Channel) error, description string) {

	go func() {
		for channel := range connectionManager.OpenChannel(description) {
			err := setUpExchangeWithQueue(channel)
			if err != nil {
				c.config.Logger.Error(err)
			}
			isReady <- err == nil
		}
	}()

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
