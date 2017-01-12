package runamqp

import (
	"fmt"
)

type logger interface {
	Info(...interface{})
	Error(...interface{})
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

type queue struct {
	Name       string
	DLQ        string
	RetryLater string
	RequeueTTL int16
	RetryLimit int
	Patterns   []string
}

type publisherConfig struct {
	connection
	exchange exchange
}

type consumerConfig struct {
	connection
	exchange exchange
	queue    queue
}

// NewPublisherConfig config for establishing a RabbitMq publisher
func NewPublisherConfig(URL string, exchangeName string, exchangeType ExchangeType, logger logger) publisherConfig {

	return publisherConfig{
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
func NewConsumerConfig(URL string, exchangeName string, exchangeType ExchangeType, queueName string, patterns []string, logger logger, requeueTTL int16, requeueLimit int, serviceName string) consumerConfig {

	if len(patterns) == 0 {
		logger.Info("Executive decision made! You did not supply a pattern so we have added a default of '#'")
		patterns = append(patterns, "#") //testme
	}

	return consumerConfig{
		connection: connection{
			URL:    URL,
			Logger: logger,
		},
		exchange: exchange{
			Name:       exchangeName,
			RetryNow:   fmt.Sprintf("%s-bounced-retry-now", exchangeName),
			RetryLater: fmt.Sprintf("%s-bounced-retry-%dms-later", exchangeName, requeueTTL),
			DLE:        fmt.Sprintf("%s-%s-dle", exchangeName, serviceName),
			Type:       exchangeType,
		},
		queue: queue{
			Name:       queueName,
			DLQ:        queueName + "-dlq",
			RetryLater: fmt.Sprintf("%s-bounced-retry-%dms-later", queueName, requeueTTL),
			RequeueTTL: requeueTTL,
			RetryLimit: requeueLimit,
			Patterns:   patterns,
		},
	}
}
