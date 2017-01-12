package runamqp

import (
	"github.com/streadway/amqp"
)

// Consumer has a channel for receiving messages
type Consumer struct {
	Messages    <-chan Message
	QueuesBound <-chan bool
}

// MessageHandler is something that can process a Message, calling Ack, Nack when appropiate for your domain
type MessageHandler interface {
	Handle(msg Message)
}

// Process creates a worker pool of size numberOfWorkers which will run handler on every message sent to the consumer Messages channel.
func (c *Consumer) Process(handler MessageHandler, numberOfWorkers int) {
	for w := 0; w < numberOfWorkers; w++ {
		go worker(handler, c.Messages)
	}
}

func worker(handler MessageHandler, messages <-chan Message) {
	for msg := range messages{
		handler.Handle(msg)
	}
}

// NewConsumer returns a Consumer
func NewConsumer(config ConsumerConfig) *Consumer {
	msgChannel := make(chan Message)
	qBound := make(chan bool)

	go func() {
		rabbitState := makeNewConnectedRabbit(config.connection, config.exchange)

		for ch := range rabbitState.newlyOpenedChannels {
			err := addMainQueueAlsoDleExchangeAndQueue(ch, config)

			if err != nil {
				qBound <- false
				return
			}

			if config.queue.RetryLimit > 0 {
				err = addRetryExchangesAndQueue(ch, config)

				if err != nil {
					qBound <- false
					return
				}
			}

			err = consumeQueue(ch, config, msgChannel)
			if err != nil {
				qBound <- false
				return
			}
			qBound <- true
		}
	}()

	return &Consumer{msgChannel, qBound}
}

func addMainQueueAlsoDleExchangeAndQueue(ch *amqp.Channel, config ConsumerConfig) error {

	err := addDleExchangeAndQueue(ch, config)

	args := make(map[string]interface{})
	args["x-dead-letter-exchange"] = config.exchange.DLE

	if err != nil {
		config.Logger.Error(err)
		return err
	}

	err = assertAndBindQueue(ch, config.queue.Name, config.exchange.Name, config.queue.Patterns, args)

	if err != nil {
		config.Logger.Error(err)
		return err
	}

	config.Logger.Info("Created queue and bound", config.queue.Name, "to exchange", config.exchange.Name, "with type", config.exchange.Type, "and with routing keys", config.queue.Patterns)

	return nil
}

func addDleExchangeAndQueue(ch *amqp.Channel, config ConsumerConfig) error {

	// make dle/dlq
	err := makeExchange(ch, config.exchange.DLE, config.exchange.Type)

	if err != nil {
		config.Logger.Error(err)
		return err
	}

	err = assertAndBindQueue(ch, config.queue.DLQ, config.exchange.DLE, []string{"#"}, nil)

	if err != nil {
		config.Logger.Error(err)
		return err
	}

	return nil
}

const matchAllPattern = "#"

func addRetryExchangesAndQueue(amqpChannel *amqp.Channel, config ConsumerConfig) error {

	retryNowExchangeName := config.exchange.RetryNow
	retryLaterExchangeName := config.exchange.RetryLater

	// make dle/dlq
	err := makeExchange(amqpChannel, retryNowExchangeName, config.exchange.Type)

	if err != nil {
		config.Logger.Error(err)
		return err
	}

	config.Logger.Info("Created retryNow exchange", retryNowExchangeName, "type of exchange:", config.exchange.Type)

	// make dle/dlq
	err = makeExchange(amqpChannel, retryLaterExchangeName, config.exchange.Type)

	if err != nil {
		config.Logger.Error(err)
		return err
	}

	config.Logger.Info("Created retryLater exchange", retryLaterExchangeName, "type of exchange:", config.exchange.Type)

	requeueArgs := make(map[string]interface{})
	requeueArgs["x-dead-letter-exchange"] = retryNowExchangeName
	requeueArgs["x-message-ttl"] = config.queue.RequeueTTL
	requeueArgs["x-dead-letter-routing-key"] = matchAllPattern
	retryNowPatterns := []string{matchAllPattern}

	err = assertAndBindQueue(amqpChannel, config.queue.RetryLater, retryLaterExchangeName, retryNowPatterns, requeueArgs)

	if err != nil {
		config.Logger.Error(err)
		return err
	}

	config.Logger.Info("Created retry later queue and bound", config.queue.RetryLater, "to exchange", retryNowExchangeName, "with type", config.exchange.Type, "and with routing keys", retryNowPatterns)

	err = amqpChannel.QueueBind(config.queue.Name, matchAllPattern, retryNowExchangeName, false, nil)

	if err != nil {
		return err
	}

	config.Logger.Info("Created the bindings for queue ", config.queue.Name, "to exchange", retryNowExchangeName, "with type", config.exchange.Type, "and with routing keys", retryNowPatterns)

	return nil
}

func consumeQueue(amqpChannel *amqp.Channel, config ConsumerConfig, messageChannel chan<- Message) error {

	msgs, err := amqpChannel.Consume(
		config.queue.Name, // queue
		"",                // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)

	if err != nil {
		config.Logger.Error(err)
		return err
	}

	config.Logger.Info("Queues bound, good to go")

	go func() {
		for d := range msgs {
			messageChannel <- &amqpMessage{
				delivery:               d,
				amqpChannel:            amqpChannel,
				retryLimit:             config.queue.RetryLimit,
				retryLaterExchangeName: config.exchange.RetryLater,
				dleExchangeName:        config.exchange.DLE,
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
