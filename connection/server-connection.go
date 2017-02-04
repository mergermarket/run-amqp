package connection

import (
	"fmt"
	"github.com/streadway/amqp"
	"math"
	"time"
)

type serverConnection interface {
	GetConnections() chan *amqp.Connection
	GetConnectionStatus() chan bool
	getErrors() chan *amqp.Error
}

type sConnection struct {
	logger              logger
	URL                 string
	openConnection      *amqp.Connection
	connections         chan *amqp.Connection
	isConnectionBlocked chan bool
	errors              chan *amqp.Error
	blockings           chan amqp.Blocking
}

func newServerConnection(URL string, logger logger) serverConnection {
	newConnection := sConnection{
		URL:                 URL,
		logger:              logger,
		connections:         make(chan *amqp.Connection),
		isConnectionBlocked: make(chan bool),
		errors:              make(chan *amqp.Error),
		blockings:           make(chan amqp.Blocking),
	}

	go newConnection.connect()

	return &newConnection
}

func (c *sConnection) GetConnections() chan *amqp.Connection {
	return c.connections
}

func (c *sConnection) GetConnectionStatus() chan bool {
	return c.isConnectionBlocked
}

func (c *sConnection) getErrors() chan *amqp.Error {
	return c.errors
}

func (c *sConnection) connect() {
	attempts := 0
	for {
		c.logger.Info("Connecting to", c.URL)
		attempts++
		openConnection, err := amqp.DialConfig(c.URL, amqp.Config{
			Heartbeat: 30 * time.Second,
		})

		if err != nil {
			c.logger.Error(fmt.Errorf("problem connecting to %s, %v", c.URL, err))
			millis := math.Exp2(float64(attempts))
			sleepDuration := time.Duration(int(millis)) * time.Second
			c.logger.Info("Trying to reconnect to RabbitMQ at", c.URL, "after", sleepDuration)
			time.Sleep(sleepDuration)
			continue
		}

		c.logger.Info("Connected to", c.URL)

		c.listenForConnectionError()
		c.listenForConnectionBlocked()
		c.openConnection = openConnection
		go func() {
			c.connections <- openConnection
		}()

		return

	}
}

func (c *sConnection) listenForConnectionError() {

	go func() {

		c.errors = c.openConnection.NotifyClose(c.errors)

		for {
			err, ok := <-c.errors
			if err != nil && ok {
				c.logger.Error(fmt.Sprintf("there was sConnection error with Code: %d Reason: %s - will try to re-connect now.", err.Code, err.Reason))
				c.closeOpenConnection()
				c.connect()
			}

		}
	}()
}

func (c *sConnection) listenForConnectionBlocked() {

	go func() {

		c.blockings = c.openConnection.NotifyBlocked(c.blockings)

		for {
			if blocking, ok := <-c.blockings; ok {
				message := fmt.Sprintf("sConnection blocking received with TCP %t ready, with reason: %s", blocking.Active, blocking.Reason)
				c.logger.Info(message)
				c.isConnectionBlocked <- blocking.Active
			}
		}
	}()
}

func (c *sConnection) closeOpenConnection() {
	err := c.openConnection.Close()

	if err != nil {
		c.logger.Error("failed to close the sConnection", err)
	}
}
