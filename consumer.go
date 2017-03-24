package runamqp

import (
	"fmt"
	"github.com/mergermarket/run-amqp/connection"
	"github.com/streadway/amqp"
	"sync"
)

type consumerChannels struct {
	mainChannel  *amqp.Channel
	dleChannel   *amqp.Channel
	retryChannel *amqp.Channel
}

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
		consumerChannels: new(consumerChannels),
	}

	go consumer.setUpConnection()

	return &consumer
}

func (c *Consumer) setUpConnection() {

	connectionManager := connection.NewConnectionManager(c.config.URL, c.config.Logger)

	mainQueueReady := make(chan bool)
	dleQueueReady := make(chan bool)
	retryQueueReady := make(chan bool)

	go c.isExchangeWithQueueReady(connectionManager, mainQueueReady, c.setUpMainExchangeWithQueue, c.config.queue.Name)
	go c.isExchangeWithQueueReady(connectionManager, dleQueueReady, c.setUpDeadLetterExchangeWithQueue, c.config.queue.DLQ)
	go c.isExchangeWithQueueReady(connectionManager, retryQueueReady, c.setUpRetryExchangeWithQueue, c.config.queue.RetryLater)

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
			if err := channel.Qos(1, 0, true); err != nil {
				c.config.Logger.Error(fmt.Sprintf("problem setting quality of service when consuming %v", err))
			}

			err := setUpExchangeWithQueue(channel)
			if err != nil {
				c.config.Logger.Error(err)
			}
			isReady <- err == nil
		}
	}()

}

func (c *Consumer) setUpMainExchangeWithQueue(amqpChannel *amqp.Channel) error {

	c.consumerChannels.mainChannel = amqpChannel

	c.config.Logger.Debug(fmt.Sprintf(`asserting the exchange: "%s" of type: "%s" and binding the queue: "%s" to it.`, c.config.exchange.Name, c.config.exchange.Type, c.config.queue.Name))

	err := makeExchange(amqpChannel, c.config.exchange.Name, c.config.exchange.Type)

	if err != nil {
		return err
	}

	err = assertAndBindQueue(amqpChannel, c.config.queue.Name, c.config.exchange.Name, c.config.queue.Patterns, nil)

	return err
}

func (c *Consumer) setUpDeadLetterExchangeWithQueue(amqpChannel *amqp.Channel) error {

	c.consumerChannels.dleChannel = amqpChannel

	c.config.Logger.Debug(fmt.Sprintf(`making DLE exchange: "%s" of type: "%s" with queue: "%s" bounds to it.`, c.config.exchange.DLE, c.config.exchange.Type, c.config.queue.DLQ))

	// make dle/dlq
	err := makeExchange(amqpChannel, c.config.exchange.DLE, c.config.exchange.Type)

	if err != nil {
		return err
	}

	err = assertAndBindQueue(amqpChannel, c.config.queue.DLQ, c.config.exchange.DLE, []string{"#"}, nil)

	return err
}

const matchAllPattern = "#"

func (c *Consumer) setUpRetryExchangeWithQueue(amqpChannel *amqp.Channel) error {

	c.consumerChannels.retryChannel = amqpChannel

	retryNowExchangeName := c.config.exchange.RetryNow
	retryLaterExchangeName := c.config.exchange.RetryLater

	c.config.Logger.Debug(fmt.Sprintf(`making RETRY-LATER exchange: "%s" of type: "%s" bound to RETRY-NOW exchage: "%s" with queue: "%s" bounds to it.`, retryLaterExchangeName, c.config.exchange.Type, retryLaterExchangeName, c.config.queue.Name))

	// make dle/dlq
	err := makeExchange(amqpChannel, retryNowExchangeName, c.config.exchange.Type)

	if err != nil {
		return err
	}

	c.config.Logger.Info("Created retryNow exchange", retryNowExchangeName, "type of exchange:", c.config.exchange.Type)

	// make dle/dlq
	err = makeExchange(amqpChannel, retryLaterExchangeName, c.config.exchange.Type)

	if err != nil {
		return err
	}

	c.config.Logger.Info("Created retryLater exchange", retryLaterExchangeName, "type of exchange:", c.config.exchange.Type)

	requeueArgs := make(map[string]interface{})
	requeueArgs["x-dead-letter-exchange"] = retryNowExchangeName
	requeueArgs["x-message-ttl"] = c.config.queue.RequeueTTL
	requeueArgs["x-dead-letter-routing-key"] = matchAllPattern
	retryNowPatterns := []string{matchAllPattern}

	err = assertAndBindQueue(amqpChannel, c.config.queue.RetryLater, retryLaterExchangeName, retryNowPatterns, requeueArgs)

	if err != nil {
		return err
	}

	c.config.Logger.Info("Created retry later queue and bound", c.config.queue.RetryLater, "to exchange", retryNowExchangeName, "with type", c.config.exchange.Type, "and with routing keys", retryNowPatterns)

	err = amqpChannel.QueueBind(c.config.queue.Name, matchAllPattern, retryNowExchangeName, false, nil)

	if err != nil {
		return err
	}

	c.config.Logger.Info("Created the bindings for queue ", c.config.queue.Name, "to exchange", retryNowExchangeName, "with type", c.config.exchange.Type, "and with routing keys", retryNowPatterns)

	return nil
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
