package runamqp

import (
	"fmt"
	"github.com/mergermarket/run-amqp/connection"
	"github.com/streadway/amqp"
	"net/http"
)

// PublishOptions will enable options being sent with the message
type PublishOptions struct {
	// Priority will dictate which messages are processed by the consumers first.  The higher the number, the higher the priority
	Priority uint8
	// PublishToQueue will send the message directly to a specific existing queue and the message will not be routed to any other queue attached to the exchange
	PublishToQueue string
	// Pattern is the routing key between the exchange and queues
	Pattern string
}

// Publisher provides a means of publishing to an exchange and is a http handler providing endpoints of GET /rabbitup, POST /entry
type Publisher struct {
	PublishReady chan bool

	currentAmqpChannel *amqp.Channel
	config             PublisherConfig
	router             *publisherServer
	publishReady       bool
}

// Publish will publish a message to an exchange
func (p *Publisher) Publish(msg []byte, options PublishOptions) error {

	exchangeName := p.config.exchange.Name
	pattern := options.Pattern
	if options.PublishToQueue != "" {
		exchangeName = ""
		pattern = options.PublishToQueue
	}

	err := p.currentAmqpChannel.Publish(
		exchangeName,
		pattern,
		true,
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
		message := fmt.Sprintf(`Published "%s" with pattern "%s" and priority "%d"`, string(msg), pattern, options.Priority)
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
