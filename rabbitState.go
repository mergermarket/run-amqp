package runamqp

import (
	"github.com/streadway/amqp"
	"math"
	"time"
)

type rabbitState struct {
	currentAmqpConnection *amqp.Connection
	currentAmqpChannel    *amqp.Channel
	newlyOpenedChannels   chan *amqp.Channel
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

// https://github.com/streadway/amqp/issues/160
func (r *rabbitState) connect() {

	//todo: what do we do when there is a conn err but no chan error?
	//connectionErrors := make(chan *amqp.Error)
	//go r.listenForConnectionErrors(connectionErrors)

	r.cleanupOldResources()

	r.config.Logger.Info("Connecting to", r.config.URL)

	r.currentAmqpConnection = connectToRabbitMQ(r.config.URL, r.config.Logger)
	//r.currentAmqpConnection.NotifyClose(connectionErrors)

	r.config.Logger.Info("Connected to", r.config.URL)

	newChannel, err := r.currentAmqpConnection.Channel()

	channelErrors := make(chan *amqp.Error)
	go r.listenForChannelErrors(channelErrors)

	newChannel.NotifyClose(channelErrors)
	sendError(err, channelErrors)

	err = makeExchange(newChannel, r.exchangeConfig.Name, r.exchangeConfig.Type)

	sendError(err, channelErrors)

	r.currentAmqpChannel = newChannel
	r.newlyOpenedChannels <- newChannel
}

//func (r *rabbitState) listenForConnectionErrors(errors chan *amqp.Error) {
//	for rabbitErr := range errors {
//		if rabbitErr != nil {
//			r.config.Logger.Error("There was a connection problem", rabbitErr)
//			r.connect()
//		}
//	}
//	r.config.Logger.Debug("Rabbit errors channel closed")
//}

func (r *rabbitState) listenForChannelErrors(errors chan *amqp.Error) {
	for rabbitErr := range errors {
		if rabbitErr != nil {
			r.config.Logger.Error("There was an error with channel", rabbitErr)
			r.connect()
		}
	}
	r.config.Logger.Debug("Rabbit errors channel closed")
}

func (r *rabbitState) cleanupOldResources() {
	r.config.Logger.Debug("Cleaning old resources before reconnecting")

	if r.currentAmqpChannel != nil {
		r.config.Logger.Debug("Closing channel", r.currentAmqpChannel)
		if err := r.currentAmqpChannel.Close(); err != nil{
			r.config.Logger.Error(err)
		} else {
			r.currentAmqpChannel = nil;
			r.config.Logger.Debug("Closed channel")
		}
	}

	if r.currentAmqpConnection != nil {
		r.config.Logger.Debug("Closing connection", r.currentAmqpConnection)
		if err := r.currentAmqpConnection.Close(); err != nil{
			r.config.Logger.Error(err)
		} else {
			r.currentAmqpConnection.ConnectionState()
			r.currentAmqpConnection = nil;
			r.config.Logger.Debug("Closed connection")
		}
	}
	r.config.Logger.Debug("Resources cleaned")
}

func connectToRabbitMQ(uri string, logger logger) *amqp.Connection {
	attempts := 0
	for {
		attempts++
		conn, err := amqp.DialConfig(uri, amqp.Config{
			Heartbeat: 30 * time.Second,
		})

		if err == nil {
			logger.Info("Connected to", uri)
			return conn
		}

		logger.Error(err)
		millis := math.Exp2(float64(attempts))
		sleepDuration := time.Duration(int(millis)) * time.Second
		logger.Info("Trying to reconnect to RabbitMQ at", uri, "after", sleepDuration)
		time.Sleep(sleepDuration)
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
