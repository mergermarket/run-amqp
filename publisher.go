package runamqp

import (
	"fmt"
	"github.com/streadway/amqp"
)

type publisher struct {
	currentAmqpChannel *amqp.Channel
	config             PublisherConfig
}

func newPublisher(channels <-chan *amqp.Channel, config PublisherConfig, publishReady chan bool) *publisher {
	p := new(publisher)
	p.config = config

	go func() {
		for ch := range channels {
			p.currentAmqpChannel = ch
			publishReady <- true
		}
	}()
	config.Logger.Info("Ready to publish")
	return p
}

// Publish will publish a message to the configured exchange
func (p *publisher) Publish(msg []byte, pattern string) error {
	err := p.currentAmqpChannel.Publish(
		p.config.exchange.Name,
		pattern,
		false,
		false,
		amqp.Publishing{
			Body: msg,
		},
	)

	if err != nil {
		p.config.Logger.Error(err)
		return fmt.Errorf("failed to publish message with error: %s", err.Error())
	}

	p.config.Logger.Info("Published", string(msg))
	return nil
}

// NewPublisher returns a function to send messages to the exchange defined in your config
func NewPublisher(config PublisherConfig) (func(message []byte, pattern string) error, <-chan bool) {
	publishReady := make(chan bool)
	rabbitState := makeNewConnectedRabbit(config.connection, config.exchange)
	p := newPublisher(rabbitState.newlyOpenedChannels, config, publishReady)

	return p.Publish, publishReady
}
