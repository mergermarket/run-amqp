package runamqp

import (
	"fmt"
)

type logger interface {
	Info(...interface{})
	Error(...interface{})
	Debug(...interface{})
}

type connectionConfig struct {
	URL    string
	Logger logger
}

type exchange struct {
	DLE        string
	Name       string
	RetryNow   string
	RetryLater string
	Type       ExchangeType
}

func (e exchange) String() string {
	return fmt.Sprintf("name %s, type %s", e.Name, e.Type)
}

type queue struct {
	DLQ           string
	MaxPriority   uint8
	Name          string
	Patterns      []string
	PrefetchCount int
	RetryLater    string
	RequeueTTL    int16
	RetryLimit    int
}

// PublisherConfig is used to create a connectionConfig to an exchange for publishing messages to
type PublisherConfig struct {
	connectionConfig
	exchange    exchange
	confirmable bool
}

// ConsumerConfig is used to create a connectionConfig to an exchange with a corresponding queue to listen to messages on
type ConsumerConfig struct {
	connectionConfig
	exchange exchange
	queue    queue
}
type NewPublisherConfig struct {
	URL          string
	exchangeName string
	exchangeType ExchangeType
	confirmable  bool
	logger       logger
}

// NewPublisherConfig returns a PublisherConfig derived from the consumer config. This config can be used to create a Publisher to Publish to this consumer
func (c ConsumerConfig) NewPublisherConfig() PublisherConfig {
	nc := NewPublisherConfig{
		URL:          c.URL,
		exchangeName: c.exchange.Name,
		exchangeType: c.exchange.Type,
		confirmable:  false,
		logger:       c.Logger,
	}
	return nc.Config()
}

// NewPublisherConfig config for establishing a RabbitMq Publisher
func (p *NewPublisherConfig) Config() PublisherConfig {

	return PublisherConfig{
		confirmable: p.confirmable,
		connectionConfig: connectionConfig{
			URL:    p.URL,
			Logger: p.logger,
		},
		exchange: exchange{
			Name: p.exchangeName,
			Type: p.exchangeType,
		},
	}
}

type NewConsumerConfig struct {
	URL          string
	exchangeName string
	exchangeType ExchangeType
	patterns     []string
	logger       logger
	requeueTTL   int16
	requeueLimit int
	serviceName  string
	prefetch     int
	maxPriority  uint8 // Optional
}

// NewConsumerConfig config for establishing a RabbitMq consumer
func (p *NewConsumerConfig) Config() ConsumerConfig {

	if len(p.patterns) == 0 {
		p.logger.Info("Executive decision made! You did not supply a pattern so we have added a default of '#'")
		p.patterns = append(p.patterns, "#") //testme
	}

	queueName := fmt.Sprintf("%s-for-%s", p.exchangeName, p.serviceName)

	return ConsumerConfig{
		connectionConfig: connectionConfig{
			URL:    p.URL,
			Logger: p.logger,
		},
		exchange: exchange{
			Name:       p.exchangeName,
			RetryNow:   fmt.Sprintf("%s-for-%s-retry-now", p.exchangeName, p.serviceName),
			RetryLater: fmt.Sprintf("%s-for-%s-retry-%dms-later", p.exchangeName, p.serviceName, p.requeueTTL),
			DLE:        fmt.Sprintf("%s-for-%s-dle", p.exchangeName, p.serviceName),
			Type:       p.exchangeType,
		},
		queue: queue{
			Name:          queueName,
			DLQ:           queueName + "-dlq",
			RetryLater:    fmt.Sprintf("%s-retry-%dms-later", queueName, p.requeueTTL),
			RequeueTTL:    p.requeueTTL,
			RetryLimit:    p.requeueLimit,
			Patterns:      p.patterns,
			MaxPriority:   p.maxPriority,
			PrefetchCount: p.prefetch,
		},
	}
}
