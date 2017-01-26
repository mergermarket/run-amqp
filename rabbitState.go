package runamqp

import (
	"fmt"
	"github.com/streadway/amqp"
	"math"
	"time"
)

type rabbitState struct {
	currentAmqpConnection *amqp.Connection
	currentAmqpChannel    *amqp.Channel
	newlyOpenedChannels   chan *amqp.Channel
	channelErrors         chan *amqp.Error
	config                connection
	exchangeConfig        exchange
}

func makeNewConnectedRabbit(config connection, exchange exchange) *rabbitState {

	r := new(rabbitState)
	r.newlyOpenedChannels = make(chan *amqp.Channel, 1)
	r.config = config
	r.exchangeConfig = exchange

	go r.connect()

	return r
}

func (r *rabbitState) connect() {
	r.currentAmqpConnection = connectToRabbitMQ(r.config.URL, r.config.Logger)
	r.createChannel()
	r.createExchange()
	r.notifyNewChannelOpened()
}

func (r *rabbitState) createExchange() {
	err := makeExchange(r.currentAmqpChannel, r.exchangeConfig.Name, r.exchangeConfig.Type)
	if err != nil {
		err = fmt.Errorf("failed to create exchange %s, %v", r.exchangeConfig, err)
		r.sendError(err)
	}

}

func (r *rabbitState) createChannel() {
	r.config.Logger.Debug("creating channel")
	newChannel, err := r.currentAmqpConnection.Channel()

	r.channelErrors = make(chan *amqp.Error)
	go r.listenForChannelErrors()

	newChannel.NotifyClose(r.channelErrors)
	r.config.Logger.Debug("creating channel - sending error", err)
	r.sendError(err)

	r.currentAmqpChannel = newChannel
}

func (r *rabbitState) notifyNewChannelOpened() {
	r.newlyOpenedChannels <- r.currentAmqpChannel
}

func (r *rabbitState) listenForChannelErrors() {
	r.config.Logger.Debug("listening for channel errors")
	for rabbitErr := range r.channelErrors {
		if rabbitErr != nil {
			r.config.Logger.Error("There was an error with channel", rabbitErr)
			r.cleanupOldResources()
			r.connect()
		}
	}
	r.config.Logger.Debug("Rabbit errors channel closed")
}

func (r *rabbitState) cleanupOldResources() {
	r.config.Logger.Debug("Cleaning old resources before reconnecting")

	if r.currentAmqpChannel != nil {
		r.config.Logger.Debug("Closing channel", r.currentAmqpChannel)
		if err := r.currentAmqpChannel.Close(); err != nil {
			r.config.Logger.Error(err)
		} else {
			r.currentAmqpChannel = nil
			r.config.Logger.Debug("Closed channel")
		}
	}

	if r.currentAmqpConnection != nil {
		r.config.Logger.Debug("Closing connection", r.currentAmqpConnection)
		if err := r.currentAmqpConnection.Close(); err != nil {
			r.config.Logger.Error(err)
		} else {
			r.currentAmqpConnection.ConnectionState()
			r.currentAmqpConnection = nil
			r.config.Logger.Debug("Closed connection")
		}
	}
	r.config.Logger.Debug("Resources cleaned")
}

func (r *rabbitState) sendError(err error) {
	defer func() {
		if re := recover(); re != nil {
			r.config.Logger.Error(fmt.Sprintf("Recovered from %v, with original error: [%v]", re, err))
		}
	}()

	if err != nil {
		if amqpErr, ok := err.(*amqp.Error); ok {
			r.channelErrors <- amqpErr
		} else {
			r.channelErrors <- &amqp.Error{
				Reason: err.Error(),
			}
		}
	}
}

func connectToRabbitMQ(uri string, logger logger) *amqp.Connection {
	attempts := 0
	for {
		logger.Info("Connecting to", uri)
		attempts++
		conn, err := amqp.DialConfig(uri, amqp.Config{
			Heartbeat: 30 * time.Second,
		})

		if err == nil {
			logger.Info("Connected to", uri)
			return conn
		}

		logger.Error(fmt.Errorf("problem connecting to %s, %v", uri, err))
		millis := math.Exp2(float64(attempts))
		sleepDuration := time.Duration(int(millis)) * time.Second
		logger.Info("Trying to reconnect to RabbitMQ at", uri, "after", sleepDuration)
		time.Sleep(sleepDuration)
	}
}
