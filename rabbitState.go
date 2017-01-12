package runamqp

import (
	"github.com/streadway/amqp"
	"time"
)

type rabbitState struct {
	currentAmqpConnection *amqp.Connection
	currentAmqpChannel    *amqp.Channel
	newlyOpenedChannels   chan *amqp.Channel
	errors                chan *amqp.Error
	config                connection
	exchangeConfig        exchange
}

func makeNewConnectedRabbit(config connection, exchange exchange) *rabbitState {

	r := new(rabbitState)
	r.newlyOpenedChannels = make(chan *amqp.Channel, 1)
	r.errors = make(chan *amqp.Error)
	r.config = config
	r.exchangeConfig = exchange

	go r.connect()
	go r.listenForErrors()

	return r
}

func (r *rabbitState) connect() {

	r.cleanupOldResources()

	r.config.Logger.Info("Connecting to", r.config.URL)

	r.currentAmqpConnection = connectToRabbitMQ(r.config.URL, r.config.Logger)
	r.currentAmqpConnection.NotifyClose(r.errors)

	r.config.Logger.Info("Connected to", r.config.URL)

	newChannel, err := r.currentAmqpConnection.Channel()

	r.config.Logger.Info("Opened channel")

	newChannel.NotifyClose(r.errors)
	sendError(err, r.errors)

	err = makeExchange(newChannel, r.exchangeConfig.Name, r.exchangeConfig.Type)

	sendError(err, r.errors)

	r.currentAmqpChannel = newChannel
	r.newlyOpenedChannels <- newChannel
}

func (r *rabbitState) listenForErrors() {
	for rabbitErr := range r.errors {
		if rabbitErr != nil {
			r.config.Logger.Error("There was an error with the rabbit connection", rabbitErr)
			r.connect()
		}
	}
}

func (r *rabbitState) cleanupOldResources() {
	if r.currentAmqpChannel != nil {
		r.currentAmqpChannel.Close()
	}

	if r.currentAmqpConnection != nil {
		r.currentAmqpConnection.Close()
	}
}

func connectToRabbitMQ(uri string, logger logger) *amqp.Connection {
	for {

		conn, err := amqp.DialConfig(uri, amqp.Config{
			Heartbeat: 30 * time.Second,
		})

		if err == nil {
			logger.Info("Connected to", uri)
			return conn
		}

		logger.Error(err)
		logger.Info("Trying to reconnect to RabbitMQ at", uri)
		time.Sleep(500 * time.Millisecond)
	}
}

func sendError(err error, errChan chan *amqp.Error) {
	if err != nil {
		if amqpErr, ok := err.(*amqp.Error); ok {
			errChan <- amqpErr
		} else {
			errChan <- &amqp.Error{
				Reason: err.Error(),
			}
		}
	}
}
