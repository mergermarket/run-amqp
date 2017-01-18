package runamqp

import (
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"net/http"
)

// Publisher provides a means of publishing to an exchange and is a http handler providing endpoints of GET /rabbitup, POST /entry
type Publisher struct {
	PublishReady chan bool

	currentAmqpChannel *amqp.Channel
	config             PublisherConfig
	router             *http.ServeMux
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

	p.config.Logger.Info("Published", string(msg))
	return nil
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

func (p *Publisher) entry(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "POST PLZ", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = p.Publish(body, "")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Written body to exchange %s", p.config.exchange.Name)
}

func (p *Publisher) rabbitup(w http.ResponseWriter, r *http.Request) {
	if p.publisherReady {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Rabbit is up!")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Rabbit did not start up!")
	}
}

func newPublisher(channels <-chan *amqp.Channel, config PublisherConfig, publishReady chan bool) *Publisher {
	p := new(Publisher)
	p.config = config
	p.PublishReady = publishReady

	p.router = http.NewServeMux()
	p.router.HandleFunc("/entry", p.entry)
	p.router.HandleFunc("/up", p.rabbitup)

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
