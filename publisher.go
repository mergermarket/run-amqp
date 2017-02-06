package runamqp

import (
	"fmt"
	"github.com/mergermarket/run-amqp/connection"
	"github.com/streadway/amqp"
	"net/http"
)

// PublishOptions will enable options being sen with the message
type PublishOptions struct {
	Priority uint8
}

// Publisher provides a means of publishing to an exchange and is a http handler providing endpoints of GET /rabbitup, POST /entry
type Publisher struct {
	PublishReady chan bool

	currentAmqpChannel *amqp.Channel
	config             PublisherConfig
	router             *publisherServer
	publishReady       bool
}

// Publish will publish a message to the configured exchange
func (p *Publisher) Publish(msg []byte, pattern string) error {
	return p.PublishWithOptions(msg, pattern, PublishOptions{})
}

// PublishWithOptions will publish a message with additional options
func (p *Publisher) PublishWithOptions(msg []byte, pattern string, options PublishOptions) error {
	err := p.currentAmqpChannel.Publish(
		p.config.exchange.Name,
		pattern,
		false,
		false,
		amqp.Publishing{
			Body:     msg,
			Priority: options.Priority,
		},
	)

	if err != nil {
		p.config.Logger.Error(err)
		return fmt.Errorf("failed to publish message with error: %s", err.Error())
	}

	if pattern != "" {
		message := fmt.Sprintf(`Published "%s" with pattern "%s"`, string(msg), pattern)
		p.config.Logger.Info(message)

	} else {
		message := fmt.Sprintf(`Published "%s"`, string(msg))
		p.config.Logger.Info(message)
	}
	return nil
}

// IsReady return true when the publisher is ready to Publish
func (p *Publisher) IsReady() bool {
	return p.publishReady
}

// NewPublisher returns a function to send messages to the exchange defined in your config. This will create a managed connection to rabbit, so you should only create this once in your application.
func NewPublisher(config PublisherConfig) *Publisher {
	p := new(Publisher)
	p.config = config
	p.PublishReady = make(chan bool)
	p.router = newPublisherServer(p, config.exchange.Name, config.Logger)

	go p.listenForOpenedAMQPChannel()
	return p
}

func (p *Publisher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Publisher) listenForOpenedAMQPChannel() {
	connectionManager := connection.NewConnectionManager(p.config.URL, p.config.Logger)
	for ch := range connectionManager.OpenChannel(p.config.exchange.Name) {
		p.currentAmqpChannel = ch
		p.publishReady = true
		p.PublishReady <- true
		p.config.Logger.Info("Ready to publish")
	}
}
