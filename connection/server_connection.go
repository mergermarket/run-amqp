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
}

type sConnection struct {
	logger              logger
	URL                 string
	openConnection      *amqp.Connection
	connections         chan *amqp.Connection
	isConnectionBlocked chan bool
}

func newServerConnection(URL string, logger logger) serverConnection {
	newConnection := sConnection{
		URL:                 URL,
		logger:              logger,
		connections:         make(chan *amqp.Connection),
		isConnectionBlocked: make(chan bool),
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

		errors := make(chan *amqp.Error)
		c.openConnection.NotifyClose(errors)

		for {
			err, ok := <-errors
			if err != nil && ok {
				c.logger.Error(fmt.Sprintf("there was sConnection error with Code: %d Reason: $s", err.Code, err.Reason))
				c.closeOpenConnection()
				c.connect()
			}

		}
	}()
}

func (c *sConnection) listenForConnectionBlocked() {

	go func() {

		blockings := make(chan amqp.Blocking)
		c.openConnection.NotifyBlocked(blockings)

		for {
			if blocking, ok := <-blockings; ok {
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
