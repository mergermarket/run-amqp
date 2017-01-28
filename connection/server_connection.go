package connection

import (
	"github.com/streadway/amqp"
	"time"
	"fmt"
	"math"
)

type logger interface {
	Info(...interface{})
	Error(...interface{})
	Debug(...interface{})
}

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
	connectionErrors    chan *amqp.Error
	connectionBlocking  chan amqp.Blocking
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

		if err == nil {
			c.logger.Info("Connected to", c.URL)

			c.listenForConnectionError()
			c.listenForConnectionBlocked()
			c.openConnection = openConnection
			c.connections <- openConnection

		}

		c.logger.Error(fmt.Errorf("problem connecting to %s, %v", c.URL, err))
		millis := math.Exp2(float64(attempts))
		sleepDuration := time.Duration(int(millis)) * time.Second
		c.logger.Info("Trying to reconnect to RabbitMQ at", c.URL, "after", sleepDuration)
		time.Sleep(sleepDuration)
	}
}

func (c *sConnection) listenForConnectionError() {
	close(c.connectionErrors)
	c.connectionErrors = make(chan *amqp.Error)
	c.openConnection.NotifyClose(c.connectionErrors)

	go func() {

		for {
			select {
			case err, ok := <-c.connectionErrors:
				if err != nil && ok {
					c.logger.Error(fmt.Sprintf("there was sConnection error with Code: %d Reason: $s", err.Code, err.Reason))
					c.closeOpenConnection()
					c.connect()
				}
			default:
				c.logger.Debug("will resume listening for sConnection errors in 3 seconds")
				time.Sleep(3 * time.Second)

			}

		}
	}()
}

func (c *sConnection) listenForConnectionBlocked() {

	close(c.connectionBlocking)
	c.connectionBlocking = make(chan amqp.Blocking)
	c.openConnection.NotifyBlocked(c.connectionBlocking)

	go func() {
		for {
			if blocking, ok := <-c.connectionBlocking; ok {
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
