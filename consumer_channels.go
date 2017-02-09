package runamqp

import (
	"fmt"
	"github.com/mergermarket/run-amqp/connection"
	"sync"
)

type consumerChannels struct {
	mainChannel  connection.AMQPChannel
	dleChannel   connection.AMQPChannel
	retryChannel connection.AMQPChannel

	config ConsumerConfig
}

func newConsumerChannels(config ConsumerConfig) *consumerChannels {
	return &consumerChannels{
		config: config,
	}
}

func (c *consumerChannels) openChannels(connectionManager connection.ConnectionManager) (ready bool) {

	mainQueueReady := c.isExchangeWithQueueReady(connectionManager, c.setUpMainExchangeWithQueue, c.config.queue.Name)
	dleQueueReady := c.isExchangeWithQueueReady(connectionManager, c.setUpDeadLetterExchangeWithQueue, c.config.queue.DLQ)
	retryQueueReady := c.isExchangeWithQueueReady(connectionManager, c.setUpRetryExchangeWithQueue, c.config.queue.RetryLater)

	return allQueuesReady(mainQueueReady, dleQueueReady, retryQueueReady)
}

func (c *consumerChannels) setUpMainExchangeWithQueue(amqpChannel connection.AMQPChannel) error {

	c.mainChannel = amqpChannel

	c.config.Logger.Debug(fmt.Sprintf(`asserting the exchange: "%s" of type: "%s" and binding the queue: "%s" to it.`, c.config.exchange.Name, c.config.exchange.Type, c.config.queue.Name))

	err := makeExchange(amqpChannel, c.config.exchange.Name, c.config.exchange.Type)

	if err != nil {
		return err
	}

	err = assertAndBindQueue(amqpChannel, c.config.queue.Name, c.config.exchange.Name, c.config.queue.Patterns, nil)

	return err
}

func (c *consumerChannels) setUpDeadLetterExchangeWithQueue(amqpChannel connection.AMQPChannel) error {

	c.dleChannel = amqpChannel

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

func (c *consumerChannels) setUpRetryExchangeWithQueue(amqpChannel connection.AMQPChannel) error {

	c.retryChannel = amqpChannel

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

func (c *consumerChannels) isExchangeWithQueueReady(connectionManager connection.ConnectionManager, setUpExchangeWithQueue func(connection.AMQPChannel) error, description string) chan bool {

	isReady := make(chan bool)

	go func() {
		for channel := range connectionManager.OpenChannel(description) {
			err := setUpExchangeWithQueue(channel)
			if err != nil {
				c.config.Logger.Error(err)
			}
			isReady <- err == nil
		}
	}()

	return isReady

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
