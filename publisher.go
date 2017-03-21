package runamqp

import (
	"fmt"
	"net/http"

	"github.com/mergermarket/run-amqp/connection"
	"github.com/streadway/amqp"
	"time"
)

// Publisher provides a means of publishing to an exchange and is a http handler providing endpoints of GET /rabbitup, POST /entry
type Publisher struct {
	currentAmqpChannel *amqp.Channel
	config             PublisherConfig
	router             *publisherServer
	publishReady       bool
}

// Publish will publish a message to an exchange
func (p *Publisher) Publish(msg []byte, options *PublishOptions) error {

	if !p.publishReady {
		return fmt.Errorf("unable to publish %s, not ready to publish, try later", string(msg))
	}

	exchangeName := p.config.exchange.Name

	var pattern string
	var priority uint8

	if options != nil {
		pattern = options.Pattern

		if options.PublishToQueue != "" {
			exchangeName = ""
			pattern = options.PublishToQueue
		}

		priority = options.Priority
	}

	err := p.currentAmqpChannel.Publish(
		exchangeName,
		pattern,
		true,
		false,
		amqp.Publishing{
			Body:     msg,
			Priority: priority,
		},
	)

	if err != nil {
		p.config.Logger.Error(err)
		return fmt.Errorf("failed to publish message with error: %s", err.Error())
	}

	if pattern != "" {
		message := fmt.Sprintf(`Published "%s" to exchange "%s" with options: %s`, string(msg), exchangeName, options)
		p.config.Logger.Debug(message)

	} else {
		message := fmt.Sprintf(`Published "%s"`, string(msg))
		p.config.Logger.Debug(message)
	}

	return nil
}

// IsReady return true when the publisher is ready to Publish
func (p *Publisher) IsReady() bool {
	return p.publishReady
}

// NewPublisher returns a function to send messages to the exchange defined in your config. This will create a managed connection to rabbit, so you should only create this once in your application.
func NewPublisher(config PublisherConfig) (*Publisher, error) {
	p := new(Publisher)
	p.config = config
	p.router = newPublisherServer(p, config.exchange.Name, config.Logger)

	go p.listenForOpenedAMQPChannel()

	select {
	case <-p.waitForReady():
		return p, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timed out waiting to create publisher %+v", config)
	}

}

func (p *Publisher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Publisher) listenForOpenedAMQPChannel() {
	connectionManager := connection.NewConnectionManager(p.config.URL, p.config.Logger)
	for ch := range connectionManager.OpenChannel(p.config.exchange.Name) {
		p.publishReady = false
		setupCurrentChannel(p, ch)
	}
}

func setupCurrentChannel(p *Publisher, ch *amqp.Channel) {
	p.currentAmqpChannel = ch

	err := makeExchange(p.currentAmqpChannel, p.config.exchange.Name, p.config.exchange.Type)

	if err != nil {
		p.config.Logger.Error(fmt.Sprintf(`failed to create the exchange "%s" with error "%+v"`, p.config.exchange.Name, err))
		p.publishReady = false
		return
	}

	p.listenForReturnedMessages()
	if p.config.confirmable {
		p.setupConfirmChannel()
	}
	p.publishReady = true
	p.config.Logger.Info("Ready to publish")
}

func (p *Publisher) setupConfirmChannel() {
	err := p.currentAmqpChannel.Confirm(false)
	if err != nil {
		p.config.Logger.Error(fmt.Sprintf(`failed to set up the channel for "%s" as confirm channel: %v`, p.config.exchange.Name, err))
		return
	}

	go func() {
		confirm := make(chan amqp.Confirmation)
		confirm = p.currentAmqpChannel.NotifyPublish(confirm)

		for res := range confirm {
			msg := fmt.Sprintf(`received a confirmation for a message that was published: "%+v" `, res)
			p.config.Logger.Debug(msg)
		}
	}()
}

func (p *Publisher) listenForReturnedMessages() {
	if p.currentAmqpChannel != nil {
		returnMessage := make(chan amqp.Return)
		p.currentAmqpChannel.NotifyReturn(returnMessage)

		go func() {
			for msg := range returnMessage {
				msg := fmt.Sprintf(`A message that was published but returned, Exchange name: "%s" Routing key: "%s" Reply text: "%s"`, msg.Exchange, msg.RoutingKey, msg.ReplyText)
				p.config.Logger.Info(msg)
			}
		}()
	}

}

func (p *Publisher) waitForReady() chan bool {
	rdy := make(chan bool)
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			if p.IsReady() {
				rdy <- true
				return
			}
		}
	}()
	return rdy
}
