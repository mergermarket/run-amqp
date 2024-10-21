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
	ExchangeName string
	ExchangeType ExchangeType
	Confirmable  bool
	Logger       logger
}

// NewPublisherConfig returns a PublisherConfig derived from the consumer config. This config can be used to create a Publisher to Publish to this consumer
func (c ConsumerConfig) NewPublisherConfig() PublisherConfig {
	nc := NewPublisherConfig{
		URL:          c.URL,
		ExchangeName: c.exchange.Name,
		ExchangeType: c.exchange.Type,
		Confirmable:  false,
		Logger:       c.Logger,
	}
	return nc.Config()
}

// NewPublisherConfig config for establishing a RabbitMq Publisher
func (p *NewPublisherConfig) Config() PublisherConfig {

	return PublisherConfig{
		confirmable: p.Confirmable,
		connectionConfig: connectionConfig{
			URL:    p.URL,
			Logger: p.Logger,
		},
		exchange: exchange{
			Name: p.ExchangeName,
			Type: p.ExchangeType,
		},
	}
}

type NewConsumerConfig struct {
	URL          string
	ExchangeName string
	ExchangeType ExchangeType
	Patterns     []string
	Logger       logger
	RequeueTTL   int16
	RequeueLimit int
	ServiceName  string
	Prefetch     int
	MaxPriority  uint8 // Optional
}

// NewConsumerConfig config for establishing a RabbitMq consumer
func (p *NewConsumerConfig) Config() ConsumerConfig {

	if len(p.Patterns) == 0 {
		p.Logger.Info("Executive decision made! You did not supply a pattern so we have added a default of '#'")
		p.Patterns = append(p.Patterns, "#") //testme
	}

	queueName := fmt.Sprintf("%s-for-%s", p.ExchangeName, p.ServiceName)

	return ConsumerConfig{
		connectionConfig: connectionConfig{
			URL:    p.URL,
			Logger: p.Logger,
		},
		exchange: exchange{
			Name:       p.ExchangeName,
			RetryNow:   fmt.Sprintf("%s-for-%s-retry-now", p.ExchangeName, p.ServiceName),
			RetryLater: fmt.Sprintf("%s-for-%s-retry-%dms-later", p.ExchangeName, p.ServiceName, p.RequeueTTL),
			DLE:        fmt.Sprintf("%s-for-%s-dle", p.ExchangeName, p.ServiceName),
			Type:       p.ExchangeType,
		},
		queue: queue{
			Name:          queueName,
			DLQ:           queueName + "-dlq",
			RetryLater:    fmt.Sprintf("%s-retry-%dms-later", queueName, p.RequeueTTL),
			RequeueTTL:    p.RequeueTTL,
			RetryLimit:    p.RequeueLimit,
			Patterns:      p.Patterns,
			MaxPriority:   p.MaxPriority,
			PrefetchCount: p.Prefetch,
		},
	}
}
