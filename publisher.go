package runamqp

import (
	"fmt"
	"github.com/streadway/amqp"
	"net/http"
)

// Publisher provides a means of publishing to an exchange and is a http handler providing endpoints of GET /rabbitup, POST /entry
type Publisher struct {
	PublishReady chan bool

	currentAmqpChannel *amqp.Channel
	config             PublisherConfig
	router             *publisherServer
	publisherReady     bool
}

// Publish will publish a message to the configured exchange
func (p *Publisher) Publish(msg []byte, pattern string) error {
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

	if pattern != "" {
		p.config.Logger.Info("Published", string(msg), "with pattern", pattern)

	} else {
		p.config.Logger.Info("Published", string(msg))
	}
	return nil
}

func (p *Publisher) IsReady() bool {
	return p.publisherReady;
}

// NewPublisher returns a function to send messages to the exchange defined in your config
func NewPublisher(config PublisherConfig) *Publisher {
	publishReady := make(chan bool)
	rabbitState := makeNewConnectedRabbit(config.connection, config.exchange)
	p := newPublisher(rabbitState.newlyOpenedChannels, config, publishReady)
	return p
}

func (p *Publisher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}


func newPublisher(channels <-chan *amqp.Channel, config PublisherConfig, publishReady chan bool) *Publisher {
	p := new(Publisher)
	p.config = config
	p.PublishReady = publishReady

	p.router = newPublisherServer(p)

	go func() {
		for ch := range channels {
			p.currentAmqpChannel = ch
			p.publisherReady = true
			publishReady <- true
		}
	}()

	config.Logger.Info("Ready to publish")
	return p
}
