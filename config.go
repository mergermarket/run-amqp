package runamqp

import (
	"fmt"
)

type logger interface {
	Info(...interface{})
	Error(...interface{})
	Debug(...interface{})
}

type connection struct {
	URL    string
	Logger logger
}

type exchange struct {
	Name       string
	RetryNow   string
	RetryLater string
	DLE        string
	Type       ExchangeType
}

func (e exchange) String() string {
	return fmt.Sprintf("name %s, type %s", e.Name, e.Type)
}

type queue struct {
	Name       string
	DLQ        string
	RetryLater string
	RequeueTTL int16
	RetryLimit int
	Patterns   []string
}

// PublisherConfig is used to create a connection to an exchange for publishing messages to
type PublisherConfig struct {
	connection
	exchange exchange
}

// ConsumerConfig is used to create a connection to an exchange with a corresponding queue to listen to messages on
type ConsumerConfig struct {
	connection
	exchange exchange
	queue    queue
}

// NewPublisherConfig returns a PublisherConfig derived from the consumer config. This config can be used to create a Publisher to Publish to this consumer
func (c ConsumerConfig) NewPublisherConfig() PublisherConfig {
	return NewPublisherConfig(c.URL, c.exchange.Name, c.exchange.Type, c.Logger)
}

// NewPublisherConfig config for establishing a RabbitMq Publisher
func NewPublisherConfig(URL string, exchangeName string, exchangeType ExchangeType, logger logger) PublisherConfig {

	return PublisherConfig{
		connection: connection{
			URL:    URL,
			Logger: logger,
		},
		exchange: exchange{
			Name: exchangeName,
			Type: exchangeType,
		},
	}
}

// NewConsumerConfig config for establishing a RabbitMq consumer
func NewConsumerConfig(URL string, exchangeName string, exchangeType ExchangeType, patterns []string, logger logger, requeueTTL int16, requeueLimit int, serviceName string) ConsumerConfig {

	if len(patterns) == 0 {
		logger.Info("Executive decision made! You did not supply a pattern so we have added a default of '#'")
		patterns = append(patterns, "#") //testme
	}

	queueName := fmt.Sprintf("%s-for-%s", exchangeName, serviceName)

	return ConsumerConfig{
		connection: connection{
			URL:    URL,
			Logger: logger,
		},
		exchange: exchange{
			Name:       exchangeName,
			RetryNow:   fmt.Sprintf("%s-for-%s-retry-now", exchangeName, serviceName),
			RetryLater: fmt.Sprintf("%s-for-%s-retry-%dms-later", exchangeName, serviceName, requeueTTL),
			DLE:        fmt.Sprintf("%s-for-%s-dle", exchangeName, serviceName),
			Type:       exchangeType,
		},
		queue: queue{
			Name:       queueName,
			DLQ:        queueName + "-dlq",
			RetryLater: fmt.Sprintf("%s-retry-%dms-later", queueName, requeueTTL),
			RequeueTTL: requeueTTL,
			RetryLimit: requeueLimit,
			Patterns:   patterns,
		},
	}
}
